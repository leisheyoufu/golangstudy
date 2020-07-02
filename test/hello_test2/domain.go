package hello_test2

import (
	. "gopkg.in/check.v1"
	"gotest.tools/assert"
)

type domain_transfer struct {
}

func (s *domain_transfer) SetUpSuite(c *C) {

	print("do some preparation here")

}

func (s *domain_transfer) TearDownSuite(c *C) {

}

func (s *domain_transfer) TestDomainTest(c *C) {
	result := true
	assert.Equal(c, true, result)

}
