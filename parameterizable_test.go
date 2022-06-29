package goyave

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParameterizableTestSuite struct {
	suite.Suite
}

func (suite *ParameterizableTestSuite) TestCompileParameters() {
	regexCache := make(map[string]*regexp.Regexp, 5)
	p := &parameterizable{}
	p.compileParameters("/product/{id:[0-9]+}", true, regexCache)
	suite.Equal([]string{"id"}, p.parameters)
	suite.NotNil(p.regex)
	suite.True(p.regex.MatchString("/product/666"))
	suite.False(p.regex.MatchString("/product/"))
	suite.False(p.regex.MatchString("/product/qwerty"))

	p = &parameterizable{}
	p.compileParameters("/product/{id:[0-9]+}/{name}", true, regexCache)
	suite.Equal([]string{"id", "name"}, p.parameters)
	suite.NotNil(p.regex)
	suite.False(p.regex.MatchString("/product/666"))
	suite.False(p.regex.MatchString("/product//"))
	suite.False(p.regex.MatchString("/product/qwerty"))
	suite.False(p.regex.MatchString("/product/qwerty/test"))
	suite.True(p.regex.MatchString("/product/666/test"))

	suite.Panics(func() { // Empty param, expect error
		p.compileParameters("/product/{}", true, regexCache)
	})
	suite.Panics(func() { // Empty name, expect error
		p.compileParameters("/product/{:[0-9]+}", true, regexCache)
	})
	suite.Panics(func() { // Empty pattern, expect error
		p.compileParameters("/product/{id:}", true, regexCache)
	})
	suite.Panics(func() { // Capturing groups
		p.compileParameters("/product/{name:(.*)}", true, regexCache)
	})
	suite.NotPanics(func() { // Non-capturing groups
		p.compileParameters("/product/{name:(?:.*)}", true, regexCache)
	})
}

func (suite *ParameterizableTestSuite) TestCompileParametersRouter() {
	regexCache := make(map[string]*regexp.Regexp, 5)
	p := &parameterizable{}
	p.compileParameters("/product/{id:[0-9]+}", false, regexCache)
	suite.Equal([]string{"id"}, p.parameters)
	suite.NotNil(p.regex)
	suite.True(p.regex.MatchString("/product/666"))
	suite.True(p.regex.MatchString("/product/666/extra"))
	suite.False(p.regex.MatchString("/product/"))
	suite.False(p.regex.MatchString("/product/qwerty"))
	suite.False(p.regex.MatchString("/product/qwerty/extra"))
}

func (suite *ParameterizableTestSuite) TestBraceIndices() {
	p := &parameterizable{}
	str := "/product/{id:[0-9]+}"
	idxs, err := p.braceIndices(str)
	suite.Nil(err)
	suite.Equal([]int{9, 19}, idxs)

	str = "/product/{id}"
	idxs, err = p.braceIndices(str)
	suite.Nil(err)
	suite.Equal([]int{9, 12}, idxs)

	str = "/product/{id:[0-9]+}/{name}" // Multiple params
	idxs, err = p.braceIndices(str)
	suite.Nil(err)
	suite.Equal([]int{9, 19, 21, 26}, idxs)

	str = "/product/{id}/{name:[\\w]+}"
	idxs, err = p.braceIndices(str)
	suite.Nil(err)
	suite.Equal([]int{9, 12, 14, 25}, idxs)

	str = "/product/{}" // Empty param, expect error
	idxs, err = p.braceIndices(str)
	suite.NotNil(err)
	suite.Equal("empty route parameter in \"/product/{}\"", err.Error())
	suite.Nil(idxs)

	str = "/product/{id:{[0-9]+}" // Unbalanced
	idxs, err = p.braceIndices(str)
	suite.NotNil(err)
	suite.Equal("unbalanced braces in \"/product/{id:{[0-9]+}\"", err.Error())
	suite.Nil(idxs)

	str = "/product/{id:}[0-9]+}" // Unbalanced
	idxs, err = p.braceIndices(str)
	suite.NotNil(err)
	suite.Equal("unbalanced braces in \"/product/{id:}[0-9]+}\"", err.Error())
	suite.Nil(idxs)
}

func (suite *ParameterizableTestSuite) TestMakeParameters() {
	matches := []string{"/product/33/param", "33", "param"}
	names := []string{"id", "name"}

	p := &parameterizable{}
	params := p.makeParameters(matches, names)

	for k := 1; k < len(matches); k++ {
		suite.Equal(matches[k], params[names[k-1]])
	}
}

func (suite *ParameterizableTestSuite) TestRegexCache() {
	regexCache := make(map[string]*regexp.Regexp, 5)
	path := "/product/{id:[0-9]+}"
	regex := "^/product/([0-9]+)$"
	p1 := &parameterizable{}
	p1.compileParameters(path, true, regexCache)
	suite.NotNil(regexCache[regex])

	p2 := &parameterizable{}
	p2.compileParameters(path, true, regexCache)
	suite.Equal(p1.regex, p2.regex)
	suite.Same(p1.regex, p2.regex)
}

func (suite *ParameterizableTestSuite) TestGetParameters() {
	p := &parameterizable{
		parameters: []string{"a", "b"},
	}
	params := p.GetParameters()
	suite.Equal(p.parameters, params)
	suite.NotSame(p.parameters, params)
}

func TestParameterizableTestSuite(t *testing.T) {
	suite.Run(t, new(ParameterizableTestSuite))
}
