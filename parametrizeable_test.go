package goyave

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ParametrizeableTestSuite struct {
	suite.Suite
}

func (suite *ParametrizeableTestSuite) TestCompileParameters() {

	p := &parametrizeable{}
	p.compileParameters("/product/{id:[0-9]+}")
	suite.Equal([]string{"id"}, p.parameters)
	suite.NotNil(p.regex)
	suite.True(p.regex.MatchString("/product/666"))
	suite.False(p.regex.MatchString("/product/"))
	suite.False(p.regex.MatchString("/product/qwerty"))

	p = &parametrizeable{}
	p.compileParameters("/product/{id:[0-9]+}/{name}")
	suite.Equal([]string{"id", "name"}, p.parameters)
	suite.NotNil(p.regex)
	suite.False(p.regex.MatchString("/product/666"))
	suite.False(p.regex.MatchString("/product//"))
	suite.False(p.regex.MatchString("/product/qwerty"))
	suite.False(p.regex.MatchString("/product/qwerty/test"))
	suite.True(p.regex.MatchString("/product/666/test"))

	suite.Panics(func() { // Empty param, expect error
		p.compileParameters("/product/{}")
	})
	suite.Panics(func() { // Empty name, expect error
		p.compileParameters("/product/{:[0-9]+}")
	})
	suite.Panics(func() { // Empty pattern, expect error
		p.compileParameters("/product/{id:}")
	})
	suite.Panics(func() { // Capturing groups
		p.compileParameters("/product/{name:(.*)}")
	})
	suite.NotPanics(func() { // Non-capturing groups
		p.compileParameters("/product/{name:(?:.*)}")
	})
}

func (suite *ParametrizeableTestSuite) TestBraceIndices() {
	p := &parametrizeable{}
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

func (suite *ParametrizeableTestSuite) TestMakeParameters() {
	matches := []string{"33", "param"}
	names := []string{"id", "name"}

	p := &parametrizeable{}
	params := p.makeParameters(matches, names)

	for k := range matches {
		suite.Equal(matches[k], params[names[k]])
	}
}

func TestParametrizeableTestSuite(t *testing.T) {
	suite.Run(t, new(ParametrizeableTestSuite))
}
