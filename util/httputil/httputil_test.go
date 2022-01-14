package httputil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"goyave.dev/goyave/v4/util/sliceutil"
)

func TestParseMultiValuesHeader(t *testing.T) {
	expected := []HeaderValue{
		{Value: "text/html", Priority: 0.8},
		{Value: "text/*", Priority: 0.8},
		{Value: "*/*", Priority: 0.8},
	}
	result := ParseMultiValuesHeader("text/html;q=0.8,text/*;q=0.8,*/*;q=0.8")
	assert.True(t, sliceutil.Equal(expected, result))

	result = ParseMultiValuesHeader("*/*;q=0.8,text/*;q=0.8,text/html;q=0.8")
	assert.True(t, sliceutil.Equal(expected, result))

	expected = []HeaderValue{
		{Value: "text/html", Priority: 1},
		{Value: "*/*", Priority: 0.7},
		{Value: "text/*", Priority: 0.5},
	}
	result = ParseMultiValuesHeader("text/html,text/*;q=0.5,*/*;q=0.7")
	assert.True(t, sliceutil.Equal(expected, result))

	expected = []HeaderValue{
		{Value: "fr", Priority: 1},
		{Value: "fr-FR", Priority: 0.8},
		{Value: "en-US", Priority: 0.5},
		{Value: "en-*", Priority: 0.3},
		{Value: "en", Priority: 0.3},
		{Value: "*", Priority: 0.3},
	}
	result = ParseMultiValuesHeader("fr , fr-FR;q=0.8, en-US ;q=0.5, *;q=0.3, en-*;q=0.3, en;q=0.3")
	assert.True(t, sliceutil.Equal(expected, result))

	expected = []HeaderValue{{Value: "fr", Priority: 1}}
	result = ParseMultiValuesHeader("fr")
	assert.True(t, sliceutil.Equal(expected, result))

	expected = []HeaderValue{{Value: "fr", Priority: 0.3}}
	result = ParseMultiValuesHeader("fr;q=0.3")
	assert.True(t, sliceutil.Equal(expected, result))
}
