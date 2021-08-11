package walk

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathHasArray(t *testing.T) {
	path := &Path{
		Name: "object",
		Type: PathTypeObject,
		Next: &Path{
			Name: "array",
			Type: PathTypeArray,
			Next: &Path{
				Type: PathTypeElement,
			},
		},
	}

	assert.True(t, path.HasArray())
	assert.False(t, path.Next.Next.HasArray())
}

func TestPathLastParent(t *testing.T) {
	path := &Path{
		Name: "object",
		Type: PathTypeObject,
		Next: &Path{
			Name: "array",
			Type: PathTypeArray,
			Next: &Path{
				Type: PathTypeElement,
			},
		},
	}

	assert.Equal(t, path.Next, path.LastParent())
	assert.Nil(t, path.Next.Next.LastParent())
}

func TestPathTail(t *testing.T) {
	path := &Path{
		Name: "object",
		Type: PathTypeObject,
		Next: &Path{
			Name: "array",
			Type: PathTypeArray,
			Next: &Path{
				Type: PathTypeElement,
			},
		},
	}
	assert.Equal(t, path.Next.Next, path.Tail())
	assert.Equal(t, path.Next.Next, path.Next.Next.Tail())
}

func TestPathClone(t *testing.T) {
	i := 1
	path := &Path{
		Name: "object",
		Type: PathTypeObject,
		Next: &Path{
			Name:  "array",
			Type:  PathTypeArray,
			Index: &i,
			Next: &Path{
				Type: PathTypeElement,
			},
		},
	}
	clone := path.Clone()
	assert.Equal(t, path, clone)
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

func TestParse(t *testing.T) {
	path, err := Parse("object.array[].field")
	assert.Nil(t, err)
	assert.Equal(t, path, &Path{
		Name: "object",
		Type: PathTypeObject,
		Next: &Path{
			Name: "array",
			Type: PathTypeArray,
			Next: &Path{
				Type: PathTypeObject,
				Next: &Path{
					Name: "field",
					Type: PathTypeElement,
				},
			},
		},
	})

	path, err = Parse("array[][]")
	assert.Nil(t, err)
	assert.Equal(t, path, &Path{
		Name: "array",
		Type: PathTypeArray,
		Next: &Path{
			Type: PathTypeArray,
			Next: &Path{
				Type: PathTypeElement,
			},
		},
	})

	path, err = Parse("object.field")
	assert.Nil(t, err)
	assert.Equal(t, path, &Path{
		Name: "object",
		Type: PathTypeObject,
		Next: &Path{
			Name: "field",
			Type: PathTypeElement,
		},
	})

	path, err = Parse("array[][].field")
	assert.Nil(t, err)
	assert.Equal(t, path, &Path{
		Name: "array",
		Type: PathTypeArray,
		Next: &Path{
			Type: PathTypeArray,
			Next: &Path{
				Type: PathTypeObject,
				Next: &Path{
					Name: "field",
					Type: PathTypeElement,
				},
			},
		},
	})

	path, err = Parse("array[][].field[]")
	assert.Nil(t, err)
	assert.Equal(t, path, &Path{
		Name: "array",
		Type: PathTypeArray,
		Next: &Path{
			Type: PathTypeArray,
			Next: &Path{
				Type: PathTypeObject,
				Next: &Path{
					Name: "field",
					Type: PathTypeArray,
					Next: &Path{
						Type: PathTypeElement,
					},
				},
			},
		},
	})

	path, err = Parse(".invalid[]path")
	assert.Nil(t, path)
	assert.NotNil(t, err)
}

func testWalk(t *testing.T, data map[string]interface{}, p string) []Context {
	matches := make([]Context, 0, 5)
	path, err := Parse(p)

	if !assert.Nil(t, err) {
		assert.FailNow(t, err.Error())
	}

	path.Walk(data, func(c Context) {
		matches = append(matches, c)
	})

	return matches
}

func TestPathWalk(t *testing.T) {
	// object.field
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"field": 5,
		},
	}
	expected := []Context{
		{
			Value:  5,
			Parent: data["object"],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{Name: "field"},
			},
			Name:     "field",
			Index:    -1,
			NotFound: false,
		},
	}
	matches := testWalk(t, data, "object.field")
	assert.Equal(t, expected, matches)

	// array[]
	data = map[string]interface{}{
		"array": []string{"a", "b", "c"},
	}
	i := 0
	j := 1
	k := 2
	l := 3
	expected = []Context{
		{
			Value:  "a",
			Parent: data["array"],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &i,
				Next:  &Path{},
			},
			Name:     "",
			Index:    0,
			NotFound: false,
		},
		{
			Value:  "b",
			Parent: data["array"],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &j,
				Next:  &Path{},
			},
			Name:     "",
			Index:    1,
			NotFound: false,
		},
		{
			Value:  "c",
			Parent: data["array"],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &k,
				Next:  &Path{},
			},
			Name:     "",
			Index:    2,
			NotFound: false,
		},
	}
	matches = testWalk(t, data, "array[]")
	assert.Equal(t, expected, matches)

	// array[][]
	data = map[string]interface{}{
		"array": [][]string{
			{},
			{"a", "b"},
			{"c"},
		},
	}
	expected = []Context{
		{
			Value:  "a",
			Parent: data["array"].([][]string)[1],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &j,
				Next: &Path{
					Type:  PathTypeArray,
					Index: &i,
					Next:  &Path{},
				},
			},
			Name:     "",
			Index:    0,
			NotFound: false,
		},
		{
			Value:  "b",
			Parent: data["array"].([][]string)[1],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &j,
				Next: &Path{
					Type:  PathTypeArray,
					Index: &j,
					Next:  &Path{},
				},
			},
			Name:     "",
			Index:    1,
			NotFound: false,
		},
		{
			Value:  "c",
			Parent: data["array"].([][]string)[2],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &k,
				Next: &Path{
					Type:  PathTypeArray,
					Index: &i,
					Next:  &Path{},
				},
			},
			Name:     "",
			Index:    0,
			NotFound: false,
		},
	}
	matches = testWalk(t, data, "array[][]")
	assert.Equal(t, expected, matches)

	// array[].field[]
	data = map[string]interface{}{
		"array": []map[string]interface{}{
			{"field": []string{}},
			{"field": []string{"a", "b"}},
			{},
			{"field": []string{"c"}},
		},
	}
	expected = []Context{
		{
			Value:  "a",
			Parent: data["array"].([]map[string]interface{})[1]["field"],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &j,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{
						Name:  "field",
						Type:  PathTypeArray,
						Index: &i,
						Next:  &Path{},
					},
				},
			},
			Name:     "",
			Index:    0,
			NotFound: false,
		},
		{
			Value:  "b",
			Parent: data["array"].([]map[string]interface{})[1]["field"],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &j,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{
						Name:  "field",
						Type:  PathTypeArray,
						Index: &j,
						Next:  &Path{},
					},
				},
			},
			Name:     "",
			Index:    1,
			NotFound: false,
		},
		{
			Value:  nil,
			Parent: data["array"].([]map[string]interface{})[2],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &k,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{
						Name: "field",
						Type: PathTypeElement,
					},
				},
			},
			Name:     "field",
			Index:    -1,
			NotFound: true,
		},
		{
			Value:  "c",
			Parent: data["array"].([]map[string]interface{})[3]["field"],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &l,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{
						Name:  "field",
						Type:  PathTypeArray,
						Index: &i,
						Next:  &Path{},
					},
				},
			},
			Name:     "",
			Index:    0,
			NotFound: false,
		},
	}
	matches = testWalk(t, data, "array[].field[]")
	assert.Equal(t, expected, matches)

	// array[].field index check
	expected = []Context{
		{
			Value:  []string{},
			Parent: data["array"].([]map[string]interface{})[0],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &i,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{Name: "field"},
				},
			},
			Name:     "field",
			Index:    -1,
			NotFound: false,
		},
		{
			Value:  []string{"a", "b"},
			Parent: data["array"].([]map[string]interface{})[1],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &j,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{Name: "field"},
				},
			},
			Name:     "field",
			Index:    -1,
			NotFound: false,
		},
		{
			Value:  nil,
			Parent: data["array"].([]map[string]interface{})[2],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &k,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{Name: "field"},
				},
			},
			Name:     "field",
			Index:    -1,
			NotFound: true,
		},
		{
			Value:  []string{"c"},
			Parent: data["array"].([]map[string]interface{})[3],
			Path: &Path{
				Name:  "array",
				Type:  PathTypeArray,
				Index: &l,
				Next: &Path{
					Type: PathTypeObject,
					Next: &Path{Name: "field"},
				},
			},
			Name:     "field",
			Index:    -1,
			NotFound: false,
		},
	}
	matches = testWalk(t, data, "array[].field")
	assert.Equal(t, expected, matches)
}

func TestPathWalkNotFoundInObject(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"field": 5,
		},
	}
	expected := []Context{
		{
			Value:  nil,
			Parent: data["object"],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{Name: "notafield"},
			},
			Name:     "notafield",
			Index:    -1,
			NotFound: true,
		},
	}
	matches := testWalk(t, data, "object.notafield")
	assert.Equal(t, expected, matches)
}

func TestPathWalkNotFoundInArray(t *testing.T) {
	data := map[string]interface{}{
		"array": []map[string]interface{}{},
	}
	expected := []Context{
		{
			Value:  nil,
			Parent: data["array"],
			Path: &Path{
				Name: "array",
				Type: PathTypeArray,
				Next: &Path{},
			},
			Name:     "",
			Index:    -1,
			NotFound: true,
		},
	}
	matches := testWalk(t, data, "array[].field")
	assert.Equal(t, expected, matches)
}

func TestPathWalkSliceExpected(t *testing.T) {
	data := map[string]interface{}{
		"object": map[string]interface{}{
			"field": []string{"a", "b"},
			"array": []interface{}{
				5,
				[]string{"a", "b"},
				map[string]interface{}{"field": "1"},
			},
		},
	}
	i := 0
	j := 1
	k := 2
	expected := []Context{
		{
			Value:  nil,
			Parent: data["object"].(map[string]interface{})["field"],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{
					Name:  "field",
					Type:  PathTypeArray,
					Index: &i,
					Next:  &Path{},
				},
			},
			Name:     "",
			Index:    0,
			NotFound: true,
		},
		{
			Value:  nil,
			Parent: data["object"].(map[string]interface{})["field"],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{
					Name:  "field",
					Type:  PathTypeArray,
					Index: &j,
					Next:  &Path{},
				},
			},
			Name:     "",
			Index:    1,
			NotFound: true,
		},
	}
	matches := testWalk(t, data, "object.field[][]")
	assert.Equal(t, expected, matches)

	expected = []Context{
		{
			Value:  nil,
			Parent: data["object"].(map[string]interface{})["array"],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{
					Name:  "array",
					Type:  PathTypeArray,
					Index: &i,
					Next:  &Path{},
				},
			},
			Name:     "",
			Index:    0,
			NotFound: true,
		},
		{
			Value:  "a",
			Parent: data["object"].(map[string]interface{})["array"].([]interface{})[1],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{
					Name:  "array",
					Type:  PathTypeArray,
					Index: &j,
					Next: &Path{
						Type:  PathTypeArray,
						Index: &i,
						Next:  &Path{},
					},
				},
			},
			Name:     "",
			Index:    0,
			NotFound: false,
		},
		{
			Value:  "b",
			Parent: data["object"].(map[string]interface{})["array"].([]interface{})[1],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{
					Name:  "array",
					Type:  PathTypeArray,
					Index: &j,
					Next: &Path{
						Type:  PathTypeArray,
						Index: &j,
						Next:  &Path{},
					},
				},
			},
			Name:     "",
			Index:    1,
			NotFound: false,
		},
		{
			Value:  nil,
			Parent: data["object"].(map[string]interface{})["array"],
			Path: &Path{
				Name: "object",
				Type: PathTypeObject,
				Next: &Path{
					Name:  "array",
					Type:  PathTypeArray,
					Index: &k,
					Next:  &Path{},
				},
			},
			Name:     "",
			Index:    2,
			NotFound: true,
		},
	}
	matches = testWalk(t, data, "object.array[][]")
	assert.Equal(t, expected, matches)
}

func TestPathWalkWithIndex(t *testing.T) {
	data := map[string]interface{}{
		"array": []map[string]interface{}{
			{"field": []string{}},
			{"field": []string{"a", "b"}},
			{},
			{"field": []string{"c"}},
			{"field": []string{"d", "e"}},
		},
	}

	i := 1
	path := &Path{
		Name:  "array",
		Type:  PathTypeArray,
		Index: &i,
		Next: &Path{
			Type: PathTypeObject,
			Next: &Path{
				Name:  "field",
				Type:  PathTypeArray,
				Index: &i,
				Next:  &Path{},
			},
		},
	}

	matches := make([]Context, 0, 1)

	path.Walk(data, func(c Context) {
		matches = append(matches, c)
	})

	expected := []Context{
		{
			Value:    "b",
			Parent:   data["array"].([]map[string]interface{})[i]["field"],
			Path:     path,
			Name:     "",
			Index:    i,
			NotFound: false,
		},
	}
	assert.Equal(t, expected, matches)
}

func TestPathWalkWithIndexOutOfBounds(t *testing.T) {
	data := map[string]interface{}{
		"array": []map[string]interface{}{
			{"field": []string{}},
			{"field": []string{"a", "b"}},
			{},
			{"field": []string{"c"}},
			{"field": []string{"d", "e"}},
		},
	}

	i := 1
	j := 5
	path := &Path{
		Name:  "array",
		Type:  PathTypeArray,
		Index: &i,
		Next: &Path{
			Type: PathTypeObject,
			Next: &Path{
				Name:  "field",
				Type:  PathTypeArray,
				Index: &j,
				Next:  &Path{},
			},
		},
	}

	matches := make([]Context, 0, 1)

	path.Walk(data, func(c Context) {
		matches = append(matches, c)
	})

	expected := []Context{
		{
			Value:    nil,
			Parent:   data["array"].([]map[string]interface{})[i]["field"],
			Path:     path,
			Name:     "",
			Index:    j,
			NotFound: true,
		},
	}
	assert.Equal(t, expected, matches)
}
