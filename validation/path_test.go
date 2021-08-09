package validation

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathHasArray(t *testing.T) {
	path := &PathItem{
		Name: "object",
		Type: PathTypeObject,
		Next: &PathItem{
			Name: "array",
			Type: PathTypeArray,
			Next: &PathItem{
				Type: PathTypeElement,
			},
		},
	}

	assert.True(t, path.HasArray())
	assert.False(t, path.Next.Next.HasArray())
}

func TestPathLastParent(t *testing.T) {
	path := &PathItem{
		Name: "object",
		Type: PathTypeObject,
		Next: &PathItem{
			Name: "array",
			Type: PathTypeArray,
			Next: &PathItem{
				Type: PathTypeElement,
			},
		},
	}

	assert.Equal(t, path.Next, path.LastParent())
	assert.Nil(t, path.Next.Next.LastParent())
}

func testPathScanner(t *testing.T, path string, expected []string) {
	scanner := createPathScanner(path)
	result := []string{}
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	assert.Equal(t, expected, result)
	assert.Nil(t, scanner.Err())
}

func testPathScannerError(t *testing.T, path string) {
	scanner := createPathScanner(path)
	result := []string{}
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	err := scanner.Err()
	if assert.NotNil(t, err) {
		assert.Contains(t, err.Error(), "Illegal syntax: ")
	} else {
		fmt.Printf("%#v\n", result)
	}
}

func TestPathScanner(t *testing.T) {
	testPathScanner(t, "object.array[].field", []string{"object", ".", "array", "[]", ".", "field"})
	testPathScanner(t, "array[][]", []string{"array", "[]", "[]"})
	testPathScanner(t, "object.field", []string{"object", ".", "field"})

	testPathScannerError(t, "object[].[]")
	testPathScannerError(t, "object..field")
	testPathScannerError(t, "object.")
	testPathScannerError(t, ".object")
	testPathScannerError(t, "array[")
	testPathScannerError(t, "array]")
	testPathScannerError(t, "array[.]")
	testPathScannerError(t, "array[aa]")
	testPathScannerError(t, "array[[]]")
	testPathScannerError(t, "array[a[b]c]")
	testPathScannerError(t, "[]array")
	testPathScannerError(t, "array[]field")
	testPathScannerError(t, "array.[]field")
	testPathScannerError(t, "array.[]field")
	testPathScannerError(t, "")
}

func TestComputePath(t *testing.T) {
	path, err := ComputePath("object.array[].field")
	assert.Nil(t, err)
	assert.Equal(t, path, &PathItem{
		Name: "object",
		Type: PathTypeObject,
		Next: &PathItem{
			Name: "array",
			Type: PathTypeArray,
			Next: &PathItem{
				Type: PathTypeObject,
				Next: &PathItem{
					Name: "field",
					Type: PathTypeElement,
				},
			},
		},
	})

	path, err = ComputePath("array[][]")
	assert.Nil(t, err)
	assert.Equal(t, path, &PathItem{
		Name: "array",
		Type: PathTypeArray,
		Next: &PathItem{
			Type: PathTypeArray,
			Next: &PathItem{
				Type: PathTypeElement,
			},
		},
	})

	path, err = ComputePath("object.field")
	assert.Nil(t, err)
	assert.Equal(t, path, &PathItem{
		Name: "object",
		Type: PathTypeObject,
		Next: &PathItem{
			Name: "field",
			Type: PathTypeElement,
		},
	})

	path, err = ComputePath("array[][].field")
	assert.Nil(t, err)
	assert.Equal(t, path, &PathItem{
		Name: "array",
		Type: PathTypeArray,
		Next: &PathItem{
			Type: PathTypeArray,
			Next: &PathItem{
				Type: PathTypeObject,
				Next: &PathItem{
					Name: "field",
					Type: PathTypeElement,
				},
			},
		},
	})

	path, err = ComputePath("array[][].field[]")
	assert.Nil(t, err)
	assert.Equal(t, path, &PathItem{
		Name: "array",
		Type: PathTypeArray,
		Next: &PathItem{
			Type: PathTypeArray,
			Next: &PathItem{
				Type: PathTypeObject,
				Next: &PathItem{
					Name: "field",
					Type: PathTypeArray,
					Next: &PathItem{
						Type: PathTypeElement,
					},
				},
			},
		},
	})

	path, err = ComputePath(".invalid[]path")
	assert.Nil(t, path)
	assert.NotNil(t, err)
}

func TestPathWalk(t *testing.T) {
	// TODO test path walk
}
