package hello_test2

import (
	"testing"

	gocheck "gopkg.in/check.v1"
)

var AllSuites []interface{}

func RunSuites(suites []interface{}, t *testing.T) {
	for _, suite := range suites {
		var _ = gocheck.Suite(suite)
		AllSuites = append(AllSuites, suite)
	}
	gocheck.TestingT(t)
}

var suites = []interface{}{
	//add test suite
	&domain_transfer{},
}

func Test_run(t *testing.T) {
	RunSuites(suites, t)
}
