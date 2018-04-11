package earcut

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

var epsilon = float64(1.0e-12)

func checkVerts(expected, actual []int) bool {
	if len(expected) != len(actual) {
		return false
	}
	for i, exp := range expected {
		if exp != actual[i] {
			return false
		}
	}
	return true
}

func TestDegenerate(t *testing.T) {
	path := []float64{
		0.0, 0.0,
		1.0, 0.0,
		0.0, 1.0,
		0.0, 0.0,
	}
	holes := []int{}
	dims := 2
	tri, err := Earcut(path, holes, dims)
	if err != nil {
		t.Error("Error making triangles for a simple triangle:", err)
	}
	if len(tri) != 3 {
		t.Errorf("Expected 3 vertex indices, got %d", len(tri))
	}
	if !checkVerts([]int{2, 0, 1}, tri) {
		t.Error("Triangle vertices don't match", tri)
	}
	if d := Deviation(path, holes, dims, tri); d > epsilon {
		t.Errorf(
			"Triangle area not equal to polygon area (%.6f%% deviation",
			d*100.0)
	}
}

func TestSimplePoly(t *testing.T) {
	path := []float64{
		0.0, 0.0,
		1.0, 0.0,
		1.309, 0.951,
		0.5, 1.539,
		-0.309, 0.951,
		0.0, 0.0,
	}
	holes := []int{}
	dims := 2
	tri, err := Earcut(path, holes, dims)
	if err != nil {
		t.Error("Error making triangles for a simple polygon:", err)
	}
	if len(tri) != 9 {
		t.Errorf("Expected 9 vertex indices, got %d", len(tri))
	}
	if !checkVerts([]int{4, 0, 1, 1, 2, 3, 3, 4, 1}, tri) {
		t.Error("Triangle vertices don't match", tri)
	}
	if d := Deviation(path, holes, dims, tri); d > epsilon {
		t.Errorf(
			"Triangle area not equal to polygon area (%.6f%% deviation",
			d*100.0)
	}
}

func TestSimplePolyWithHole(t *testing.T) {
	path := []float64{
		0.0, 0.0,
		1.0, 0.0,
		1.309, 0.951,
		0.5, 1.539,
		-0.309, 0.951,
		0.0, 0.0,
		0.25, 0.25,
		0.30, 0.25,
		0.30, 0.30,
		0.25, 0.30,
		0.25, 0.25,
	}
	holes := []int{6}
	dims := 2
	tri, err := Earcut(path, holes, dims)
	if err != nil {
		t.Error("Error making triangles for a simple polygon with a hole:", err)
	}
	if len(tri) != 27 {
		t.Errorf("Expected 27 vertex indices, got %d", len(tri))
	}
	exp := []int{
		0, 1, 2,
		2, 3, 4,
		10, 4, 0,
		4, 10, 9,
		7, 10, 0,
		2, 4, 9,
		7, 0, 2,
		2, 9, 8,
		8, 7, 2,
	}
	if !checkVerts(exp, tri) {
		t.Error("Triangle vertices don't match", tri)
	}
	if d := Deviation(path, holes, dims, tri); d > epsilon {
		t.Errorf(
			"Triangle area not equal to polygon area (%.6f%% deviation",
			d*100.0)
	}
}

func TestComplexPoly(t *testing.T) {
	path := []float64{
		0.0, 0.0, 1.0,
		11.0, 0.0, 2.0,
		11.0, 9.0, 3.0,
		6.0, 9.0, 4.0,
		6.0, 3.0, 5.0,
		3.0, 3.0, 6.0,
		3.0, 9.0, 7.0,
		2.0, 9.0, 8.0,
		2.0, 2.0, 9.0,
		7.0, 2.0, 10.0,
		7.0, 8.0, 11.0,
		10.0, 8.0, 12.0,
		10.0, 1.0, 13.0,
		1.0, 1.0, 14.0,
		1.0, 10.0, 15.0,
		4.0, 10.0, 16.0,
		4.0, 4.0, 17.0,
		5.0, 4.0, 18.0,
		5.0, 11.0, 19.0,
		0.0, 11.0, 20.0,
		0.0, 0.0, 21.0,
	}
	holes := []int{}
	dims := 3
	tri, err := Earcut(path, holes, dims)
	if err != nil {
		t.Error("Error making triangles for a complex polygon:", err)
	}
	if len(tri) != 54 {
		t.Errorf("Expected 54 vertex indices, got %d", len(tri))
	}
	exp := []int{
		5, 6, 7,
		15, 16, 17,
		5, 7, 8,
		15, 17, 18,
		4, 5, 8,
		14, 15, 18,
		4, 8, 9,
		14, 18, 19,
		3, 4, 9,
		13, 14, 19,
		3, 9, 10,
		13, 19, 0,
		2, 3, 10,
		12, 13, 0,
		2, 10, 11,
		12, 0, 1,
		1, 2, 11,
		11, 12, 1,
	}
	if !checkVerts(exp, tri) {
		t.Error("Triangle vertices don't match", tri)
	}
	if d := Deviation(path, holes, dims, tri); d > epsilon {
		t.Errorf(
			"Triangle area not equal to polygon area (%.6f%% deviation",
			d*100.0)
	}
}

func flatten(data [][][2]float64) ([]float64, []int) {
	flat := []float64{}
	holes := []int{}
	j := 0
	for i, ring := range data {
		if i > 0 {
			holes = append(holes, j)
		}
		for _, pt := range ring {
			flat = append(flat, pt[0], pt[1])
			j++
		}
	}
	return flat, holes
}

func loadVertices(name string) ([]float64, []int, error) {
	rawdata, err := ioutil.ReadFile(filepath.Join("fixtures", name+".json"))
	if err != nil {
		return nil, nil, fmt.Errorf("Error reading fixture data: %s", err)
	}
	data := [][][2]float64{}
	err = json.Unmarshal(rawdata, &data)
	if err != nil {
		return nil, nil, fmt.Errorf("Error unmarshaling json fixture data: %s", err)
	}
	flat, holeIndices := flatten(data)
	return flat, holeIndices, nil
}

func testFixture(name string, expTriangles int, expDeviation float64, t *testing.T) {
	flat, holeIndices, err := loadVertices(name)
	if err != nil {
		t.Error(err)
	}
	tri, err := Earcut(flat, holeIndices, 2)
	if err != nil {
		t.Error("Error in earcut:", err)
	}
	d := Deviation(flat, holeIndices, 2, tri)
	if d > expDeviation {
		t.Errorf("Deviation %f greater than expected (%f) for %s", d, expDeviation, name)
	}
	if len(tri)/3 != expTriangles {
		t.Errorf("Expected %d triangles, got %d for fixture %s", expTriangles, len(tri)/3, name)
	}
}

func TestFixtureBuilding(t *testing.T) {
	testFixture("building", 13, epsilon, t)
}

func TestFixtureDude(t *testing.T) {
	testFixture("dude", 106, epsilon, t)
}

func TestFixtureWater(t *testing.T) {
	testFixture("water", 2482, 0.0008, t)
}

func TestFixtureWater2(t *testing.T) {
	testFixture("water2", 1212, epsilon, t)
}

func TestFixtureWater3(t *testing.T) {
	testFixture("water3", 197, epsilon, t)
}

func TestFixtureWater3b(t *testing.T) {
	testFixture("water3b", 25, epsilon, t)
}

func TestFixtureWater4(t *testing.T) {
	testFixture("water4", 705, epsilon, t)
}

func TestFixtureWaterHuge(t *testing.T) {
	testFixture("water-huge", 5174, 0.0011, t)
}

func TestFixtureWaterHuge2(t *testing.T) {
	testFixture("water-huge2", 4461, 0.0028, t)
}

func TestFixtureDegenerate(t *testing.T) {
	testFixture("degenerate", 0, epsilon, t)
}

func TestFixtureBadHole(t *testing.T) {
	testFixture("bad-hole", 42, 0.019, t)
}

func TestFixtureEmptySquare(t *testing.T) {
	testFixture("empty-square", 0, epsilon, t)
}

func TestFixtureIssue16(t *testing.T) {
	testFixture("issue16", 12, epsilon, t)
}

func TestFixtureIssue17(t *testing.T) {
	testFixture("issue17", 11, epsilon, t)
}

func TestFixtureSteiner(t *testing.T) {
	testFixture("steiner", 9, epsilon, t)
}

func TestFixtureIssue29(t *testing.T) {
	testFixture("issue29", 40, epsilon, t)
}

func TestFixtureIssue34(t *testing.T) {
	testFixture("issue34", 139, epsilon, t)
}

func TestFixtureIssue35(t *testing.T) {
	testFixture("issue35", 844, epsilon, t)
}

func TestFixtureSelfTouching(t *testing.T) {
	testFixture("self-touching", 124, 3.4e-14, t)
}

func TestFixtureOutsideRing(t *testing.T) {
	testFixture("outside-ring", 64, epsilon, t)
}

func TestFixtureSimplifiedUSBorder(t *testing.T) {
	testFixture("simplified-us-border", 120, epsilon, t)
}

func TestFixtureTouchingHoles(t *testing.T) {
	testFixture("touching-holes", 57, epsilon, t)
}

func TestFixtureHoleTouchingOuter(t *testing.T) {
	testFixture("hole-touching-outer", 77, epsilon, t)
}

func TestFixtureHilbert(t *testing.T) {
	testFixture("hilbert", 1024, epsilon, t)
}

func TestFixtureIssue45(t *testing.T) {
	testFixture("issue45", 10, epsilon, t)
}

func TestFixtureEberly3(t *testing.T) {
	testFixture("eberly-3", 73, epsilon, t)
}

func TestFixtureEberly6(t *testing.T) {
	testFixture("eberly-6", 1429, epsilon, t)
}

func TestFixtureIssue52(t *testing.T) {
	testFixture("issue52", 109, epsilon, t)
}

func TestFixtureSharedPoints(t *testing.T) {
	testFixture("shared-points", 4, epsilon, t)
}

func TestFixtureBadDiagonals(t *testing.T) {
	testFixture("bad-diagonals", 7, epsilon, t)
}

func TestFixtureIssue83(t *testing.T) {
	testFixture("issue83", 0, 1e-14, t)
}
