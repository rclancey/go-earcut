package earcut

import (
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
