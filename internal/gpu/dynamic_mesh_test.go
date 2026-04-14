package gpu

import (
	"reflect"
	"testing"
)

func TestMakeQuadIndicesSingle(t *testing.T) {
	result := makeQuadIndices(1)
	expected := []uint32{0, 1, 2, 2, 3, 0}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("makeQuadIndices(1) = %v, want %v", result, expected)
	}
}

func TestMakeQuadIndicesTwo(t *testing.T) {
	result := makeQuadIndices(2)
	expected := []uint32{0, 1, 2, 2, 3, 0, 4, 5, 6, 6, 7, 4}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("makeQuadIndices(2) = %v, want %v", result, expected)
	}
}

func TestMakeQuadIndicesEmpty(t *testing.T) {
	result := makeQuadIndices(0)
	if len(result) != 0 {
		t.Errorf("makeQuadIndices(0) = %v, want empty slice", result)
	}
}
