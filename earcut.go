package earcut

// Port of https://github.com/mapbox/earcut
// See LICENSE

import (
	"errors"
	"math"
	"sort"
)

type node struct {
	i       int
	x       float64
	y       float64
	prev    *node
	next    *node
	z       *int
	prevZ   *node
	nextZ   *node
	steiner bool
}

// Earcut returns an int array of vertex indices that make up the triangles
// of the polygon.
//
// The polygon vertices should be passed as a flat array of float64 values.
//
// holeIndices is an array of integers pointing to the vertex index (not
// the data index) of the start of each hole (if any).
//
// dim is the number of values per vertex.  Only the first two values (x & y)
// will be considered when constructing the triangles.
func Earcut(data []float64, holeIndices []int, dim int) ([]int, error) {
	if dim < 2 {
		return nil, errors.New("need at least 2 dimensions")
	}
	hasHoles := len(holeIndices) > 0
	var outerLen int
	if hasHoles {
		outerLen = holeIndices[0] * dim
	} else {
		outerLen = len(data)
	}
	outerNode := linkedList(data, 0, outerLen, dim, true)
	triangles := []int{}
	if outerNode == nil {
		return triangles, nil
	}
	minX := math.Inf(1)
	minY := math.Inf(1)
	maxX := math.Inf(-1)
	maxY := math.Inf(-1)
	var x, y, invSize float64
	if hasHoles {
		outerNode = eliminateHoles(data, holeIndices, outerNode, dim)
	}

	// if the shape is not too simple, we'll use z-order curve hash later;
	// calculate polygon bbox
	if len(data) > 80*dim {
		for i := 0; i < outerLen; i += dim {
			x = data[i]
			y = data[i+1]
			if x < minX {
				minX = x
			}
			if x > maxX {
				maxX = x
			}
			if y < minY {
				minY = y
			}
			if y > maxY {
				maxY = y
			}

			// minX, minY and invSize are later used to transform coords into
			// integers for z-order calculation
			invSize = math.Max(maxX-minX, maxY-minY)
			if invSize != 0.0 {
				invSize = 1.0 / invSize
			}
		}
	}
	earcutLinked(outerNode, &triangles, dim, minX, minY, invSize, 0)
	return triangles, nil
}

// create a circular doubly linked list from polygon points in the specified
// winding order
func linkedList(data []float64, start, end, dim int, clockwise bool) *node {
	var last *node
	if clockwise == (signedArea(data, start, end, dim) > 0.0) {
		for i := start; i < end; i += dim {
			last = insertNode(i, data[i], data[i+1], last)
		}
	} else {
		for i := end - dim; i >= start; i -= dim {
			last = insertNode(i, data[i], data[i+1], last)
		}
	}
	if last != nil && equals(last, last.next) {
		removeNode(last)
		last = last.next
	}
	return last
}

// eliminate colinear or duplicate points
func filterPoints(start, end *node) *node {
	if start == nil {
		return start
	}
	if end == nil {
		end = start
	}
	p := start
	again := false
	for {
		again = false
		if !p.steiner && (equals(p, p.next) || area(p.prev, p, p.next) == 0.0) {
			removeNode(p)
			end = p.prev
			p = p.prev
			if p == p.next {
				break
			}
			again = true
		} else {
			p = p.next
		}
		if !again && p == end {
			break
		}
	}
	return end
}

// main ear slicing loop which triangulates a polygon (given as a linked list)
func earcutLinked(ear *node, triangles *[]int, dim int, minX, minY, invSize float64, pass int) {
	if ear == nil {
		return
	}

	// interlink polygon nodes in z-order
	if pass != 0 && invSize != 0.0 {
		indexCurve(ear, minX, minY, invSize)
	}

	stop := ear
	var prev, next *node
	var test bool
	// iterate through ears, slicing them one by one
	for ear.prev != ear.next {
		prev = ear.prev
		next = ear.next

		if invSize != 0.0 {
			test = isEarHashed(ear, minX, minY, invSize)
		} else {
			test = isEar(ear)
		}
		if test {
			// cut off the triangle
			*triangles = append(*triangles, prev.i/dim, ear.i/dim, next.i/dim)
			removeNode(ear)

			// skipping the next vertice leads to less sliver triangles
			ear = next.next
			stop = next.next
			continue
		}
		ear = next

		// if we looped through the whole remaining polygon and can't find any
		// more ears
		if ear == stop {
			// try filtering points and slicing again
			if pass == 0 {
				earcutLinked(
					filterPoints(ear, nil),
					triangles,
					dim,
					minX,
					minY,
					invSize,
					1)
				// if this didn't work, try curing all small
				// self-intersections locally
			} else if pass == 1 {
				ear = cureLocalIntersections(ear, triangles, dim)
				earcutLinked(ear, triangles, dim, minX, minY, invSize, 2)
				// as a last resort, try splitting the remaining polygon
				// into two
			} else if pass == 2 {
				splitEarcut(ear, triangles, dim, minX, minY, invSize)
			}
			break
		}
	}
}

// check whether a polygon node forms a valid ear with adjacent nodes
func isEar(ear *node) bool {
	a := ear.prev
	b := ear
	c := ear.next

	if area(a, b, c) >= 0.0 {
		// reflex, can't be an ear
		return false
	}

	// now make sure we don't have other points inside the potential ear
	p := ear.next.next

	for p != ear.prev {
		if pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, p.x, p.y) &&
			area(p.prev, p, p.next) >= 0.0 {
			return false
		}
		p = p.next
	}

	return true
}

func isEarHashed(ear *node, minX, minY, invSize float64) bool {
	a := ear.prev
	b := ear
	c := ear.next
	if area(a, b, c) >= 0.0 {
		// reflex, can't be an ear
		return false
	}

	// triangle bbox; min & max are calculated like this for speed
	minTX := math.Min(a.x, math.Min(b.x, c.x))
	minTY := math.Min(a.y, math.Min(b.y, c.y))
	maxTX := math.Max(a.x, math.Max(b.x, c.x))
	maxTY := math.Max(a.y, math.Max(b.y, c.y))

	// z-order range for the current triangle bbox;
	minZ := zOrder(minTX, minTY, minX, minY, invSize)
	maxZ := zOrder(maxTX, maxTY, minX, minY, invSize)

	p := ear.prevZ
	n := ear.nextZ

	// look for points inside the triangle in both directions
	for p != nil && *p.z >= minZ && n != nil && *n.z <= maxZ {
		if p != ear.prev &&
			p != ear.next &&
			pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, p.x, p.y) &&
			area(p.prev, p, p.next) >= 0.0 {
			return false
		}
		p = p.prevZ

		if n != ear.prev &&
			n != ear.next &&
			pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, n.x, n.y) &&
			area(n.prev, n, n.next) >= 0.0 {
			return false
		}
		n = n.nextZ
	}

	// look for remaining points in decreasing z-order
	for p != nil && *p.z >= minZ {
		if p != ear.prev &&
			p != ear.next &&
			pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, p.x, p.y) &&
			area(p.prev, p, p.next) >= 0.0 {
			return false
		}
		p = p.prevZ
	}

	// look for remaining points in increasing z-order
	for n != nil && *n.z <= maxZ {
		if n != ear.prev &&
			n != ear.next &&
			pointInTriangle(a.x, a.y, b.x, b.y, c.x, c.y, n.x, n.y) &&
			area(n.prev, n, n.next) >= 0.0 {
			return false
		}
		n = n.nextZ
	}

	return true
}

// go through all polygon nodes and cure small local self-intersections
func cureLocalIntersections(start *node, triangles *[]int, dim int) *node {
	p := start
	for {
		a := p.prev
		b := p.next.next

		if !equals(a, b) &&
			intersects(a, p, p.next, b) &&
			locallyInside(a, b) &&
			locallyInside(b, a) {
			*triangles = append(*triangles, a.i/dim, p.i/dim, b.i/dim)

			// remove two nodes involved
			removeNode(p)
			removeNode(p.next)

			p = b
			start = b
		}
		p = p.next
		if p == start {
			break
		}
	}

	return p
}

// try splitting polygon into two and triangulate them independently
func splitEarcut(start *node, triangles *[]int, dim int, minX, minY, invSize float64) {
	// look for a valid diagonal that divides the polygon into two
	a := start
	for {
		b := a.next.next
		for b != a.prev {
			if a.i != b.i && isValidDiagonal(a, b) {
				// split the polygon in two by the diagonal
				c := splitPolygon(a, b)

				// filter colinear points around the cuts
				a = filterPoints(a, a.next)
				c = filterPoints(c, c.next)

				// run earcut on each half
				earcutLinked(a, triangles, dim, minX, minY, invSize, 0)
				earcutLinked(c, triangles, dim, minX, minY, invSize, 0)
				return
			}
			b = b.next
		}
		a = a.next
		if a == start {
			break
		}
	}
}

type sortableQueue []*node

func (q sortableQueue) Len() int           { return len(q) }
func (q sortableQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q sortableQueue) Less(i, j int) bool { return q[i].x < q[j].x }

// link every hole into the outer loop, producing a single-ring polygon
// without holes
func eliminateHoles(data []float64, holeIndices []int, outerNode *node, dim int) *node {
	queue := []*node{}
	var start, end int
	var list *node
	l := len(holeIndices)
	for i := 0; i < l; i++ {
		start = holeIndices[i] * dim
		if i < l-1 {
			end = holeIndices[i+1] * dim
		} else {
			end = len(data)
		}
		list = linkedList(data, start, end, dim, false)
		if list == list.next {
			list.steiner = true
		}
		queue = append(queue, getLeftmost(list))
	}

	sort.Sort(sortableQueue(queue))

	// process holes from left to right
	for i := 0; i < len(queue); i++ {
		eliminateHole(queue[i], outerNode)
		outerNode = filterPoints(outerNode, outerNode.next)
	}

	return outerNode
}

// find a bridge between vertices that connects hole with an outer ring and
// link it
func eliminateHole(hole, outerNode *node) {
	outerNode = findHoleBridge(hole, outerNode)
	if outerNode != nil {
		b := splitPolygon(outerNode, hole)
		filterPoints(b, b.next)
	}
}

// David Eberly's algorithm for finding a bridge between hole and outer polygon
func findHoleBridge(hole, outerNode *node) *node {
	p := outerNode
	hx := hole.x
	hy := hole.y
	qx := math.Inf(-1)
	var m *node

	// find a segment intersected by a ray from the hole's leftmost point
	// to the left; segment's endpoint with lesser x will be potential
	// connection point
	for {
		if hy <= p.y && hy >= p.next.y && p.next.y != p.y {
			x := p.x + (hy-p.y)*(p.next.x-p.x)/(p.next.y-p.x)
			if x <= hx && x > qx {
				qx = x
				if x == hx {
					if hy == p.y {
						return p
					}
					if hy == p.next.y {
						return p.next
					}
				}
				if p.x < p.next.x {
					m = p
				} else {
					m = p.next
				}
			}
		}
		p = p.next
		if p == outerNode {
			break
		}
	}
	if m == nil {
		return nil
	}

	if hx == qx {
		// hole touches outer segment; pick lower endpoint
		return m.prev
	}

	// look for points inside the triangle of hole point, segment
	// intersection and endpoint; if there are no points found, we have a
	// valid connection; otherwise choose the point of the minimum angle
	// with the ray as connection point

	stop := m
	mx := m.x
	my := m.y
	tanMin := math.Inf(1)
	var tan float64

	p = m.next

	var xx float64
	for p != stop {
		if hy < my {
			xx = hx
		} else {
			xx = qx
		}
		if hx >= p.x &&
			p.x >= mx &&
			hx != p.x &&
			pointInTriangle(xx, hy, mx, my, xx, hy, p.x, p.y) {
			tan = math.Abs(hy-p.y) / (hx - p.x) // tangential

			if (tan < tanMin || (tan == tanMin && p.x > m.x)) &&
				locallyInside(p, hole) {
				m = p
				tanMin = tan
			}
		}

		p = p.next
	}

	return m
}

// interlink polygon nodes in z-order
func indexCurve(start *node, minX, minY, invSize float64) {
	p := start
	for {
		if p.z == nil {
			z := zOrder(p.x, p.y, minX, minY, invSize)
			p.z = &z
		}
		p.prevZ = p.prev
		p.nextZ = p.next
		p = p.next
		if p == start {
			break
		}
	}

	p.prevZ.nextZ = nil
	p.prevZ = nil

	sortLinked(p)
}

// Simon Tatham's linked list merge sort algorithm
// http://www.chiark.greenend.org.uk/~sgtatham/algorithms/listsort.html
func sortLinked(list *node) *node {
	var p, q, e, tail *node
	var numMerges, pSize, qSize int
	inSize := 1

	for {
		p = list
		list = nil
		tail = nil
		numMerges = 0

		for p != nil {
			numMerges++
			q = p
			pSize = 0
			for i := 0; i < inSize; i++ {
				pSize++
				q = q.nextZ
				if q == nil {
					break
				}
			}
			qSize = inSize

			for pSize > 0 || (qSize > 0 && q != nil) {

				if pSize != 0 && (qSize == 0 || q == nil || *p.z <= *q.z) {
					e = p
					p = p.nextZ
					pSize--
				} else {
					e = q
					q = q.nextZ
					qSize--
				}

				if tail != nil {
					tail.nextZ = e
				} else {
					list = e
				}

				e.prevZ = tail
				tail = e
			}

			p = q
		}

		tail.nextZ = nil
		inSize *= 2

		if numMerges <= 1 {
			break
		}
	}

	return list
}

// z-order of a point given coords and inverse of the longer side of data bbox
func zOrder(x, y, minX, minY, invSize float64) int {
	// coords are transformed into non-negative 15-bit integer range
	ix := 32767 * int((x-minX)*invSize)
	iy := 32767 * int((y-minY)*invSize)

	ix = (ix | (ix << 8)) & 0x00FF00FF
	ix = (ix | (ix << 4)) & 0x0F0F0F0F
	ix = (ix | (ix << 2)) & 0x33333333
	ix = (ix | (ix << 1)) & 0x55555555

	iy = (iy | (iy << 8)) & 0x00FF00FF
	iy = (iy | (iy << 4)) & 0x0F0F0F0F
	iy = (iy | (iy << 2)) & 0x33333333
	iy = (iy | (iy << 1)) & 0x55555555

	return ix | (iy << 1)
}

// find the leftmost node of a polygon ring
func getLeftmost(start *node) *node {
	p := start
	leftmost := start
	for {
		if p.x < leftmost.x {
			leftmost = p
		}
		p = p.next
		if p == start {
			break
		}
	}

	return leftmost
}

// check if a point lies within a convex triangle
func pointInTriangle(ax, ay, bx, by, cx, cy, px, py float64) bool {
	return (cx-px)*(ay-py)-(ax-px)*(cy-py) >= 0.0 &&
		(ax-px)*(by-py)-(bx-px)*(ay-py) >= 0.0 &&
		(bx-px)*(cy-py)-(cx-px)*(by-py) >= 0.0
}

// check if a diagonal between two polygon nodes is valid (lies in
// polygon interior)
func isValidDiagonal(a, b *node) bool {
	return a.next.i != b.i && a.prev.i != b.i && !intersectsPolygon(a, b) && locallyInside(a, b) && locallyInside(b, a) && middleInside(a, b)
}

// signed area of a triangle
func area(p, q, r *node) float64 {
	return (q.y-p.y)*(r.x-q.x) - (q.x-p.x)*(r.y-q.y)
}

// check if two points are equal
func equals(p1, p2 *node) bool {
	return p1.x == p2.x && p1.y == p2.y
}

// check if two segments intersect
func intersects(p1, q1, p2, q2 *node) bool {
	if (equals(p1, q1) && equals(p2, q2)) ||
		(equals(p1, q2) && equals(p2, q1)) {
		return true
	}
	return (area(p1, q1, p2) > 0.0) != (area(p1, q1, q2) > 0.0) &&
		(area(p2, q2, p1) > 0.0) != (area(p2, q2, q1) > 0.0)
}

// check if a polygon diagonal intersects any polygon segments
func intersectsPolygon(a, b *node) bool {
	p := a
	for {
		if p.i != a.i &&
			p.next.i != a.i &&
			p.i != b.i &&
			p.next.i != b.i &&
			intersects(p, p.next, a, b) {
			return true
		}
		p = p.next
		if p == a {
			break
		}
	}

	return false
}

// check if a polygon diagonal is locally inside the polygon
func locallyInside(a, b *node) bool {
	if area(a.prev, a, a.next) < 0.0 {
		return area(a, b, a.next) >= 0.0 && area(a, a.prev, b) >= 0.0
	}
	return area(a, b, a.prev) < 0.0 || area(a, a.next, b) < 0.0
}

// check if the middle point of a polygon diagonal is inside the polygon
func middleInside(a, b *node) bool {
	p := a
	inside := false
	px := (a.x + b.x) / 2.0
	py := (a.y + b.y) / 2.0
	for {
		if ((p.y > py) != (p.next.y > py)) &&
			p.next.y != p.y &&
			(px < (p.next.x-p.x)*(py-p.y)/(p.next.y-p.y)+p.x) {
			inside = !inside
		}
		p = p.next
		if p == a {
			break
		}
	}

	return inside
}

// link two polygon vertices with a bridge; if the vertices belong to the
// same ring, it splits polygon into two; if one belongs to the outer ring
// and another to a hole, it merges it into a single ring
func splitPolygon(a, b *node) *node {
	a2 := newNode(a.i, a.x, a.y)
	b2 := newNode(b.i, b.x, b.y)
	an := a.next
	bp := b.prev

	a.next = b
	b.prev = a

	a2.next = an
	an.prev = a2

	b2.next = a2
	a2.prev = b2

	bp.next = b2
	b2.prev = bp

	return b2
}

// create a node and optionally link it with previous one (in a circular
// doubly linked list)
func insertNode(i int, x, y float64, last *node) *node {
	p := newNode(i, x, y)

	if last == nil {
		p.prev = p
		p.next = p

	} else {
		p.next = last.next
		p.prev = last
		last.next.prev = p
		last.next = p
	}
	return p
}

func removeNode(p *node) {
	p.next.prev = p.prev
	p.prev.next = p.next

	if p.prevZ != nil {
		p.prevZ.nextZ = p.nextZ
	}
	if p.nextZ != nil {
		p.nextZ.prevZ = p.prevZ
	}
}

func newNode(i int, x, y float64) *node {
	return &node{
		i:       i,
		x:       x,
		y:       y,
		prev:    nil,
		next:    nil,
		z:       nil,
		prevZ:   nil,
		nextZ:   nil,
		steiner: false,
	}
}

func signedArea(data []float64, start, end, dim int) float64 {
	var sum float64 = 0.0
	for i, j := start, end-dim; i < end; i += dim {
		sum += (data[j] - data[i]) * (data[i+1] + data[j+1])
		j = i
	}
	return sum
}

// Deviation returns a percentage difference between the polygon area and
// its triangulation area; used to verify correctness of triangulation
func Deviation(data []float64, holeIndices []int, dim int, triangles []int) float64 {
	hasHoles := holeIndices != nil && len(holeIndices) > 0
	var outerLen int
	if hasHoles {
		outerLen = holeIndices[0] * dim
	} else {
		outerLen = len(data)
	}

	polygonArea := math.Abs(signedArea(data, 0, outerLen, dim))
	var start, end int
	if hasHoles {
		for i, l := 0, len(holeIndices); i < l; i++ {
			start = holeIndices[i] * dim
			if i < l-1 {
				end = holeIndices[i+1] * dim
			} else {
				end = len(data)
			}
			polygonArea -= math.Abs(signedArea(data, start, end, dim))
		}
	}

	var trianglesArea float64 = 0.0
	for i := 0; i < len(triangles); i += 3 {
		a := triangles[i] * dim
		b := triangles[i+1] * dim
		c := triangles[i+2] * dim
		trianglesArea += math.Abs(
			(data[a]-data[c])*(data[b+1]-data[a+1]) -
				(data[a]-data[b])*(data[c+1]-data[a+1]))
	}

	if polygonArea == 0.0 && trianglesArea == 0.0 {
		return 0.0
	}
	if polygonArea == 0.0 {
		return math.Inf(1)
	}
	return math.Abs((trianglesArea - polygonArea) / polygonArea)
}
