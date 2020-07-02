package hello_test

import (
	"io"
	"testing"

	. "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})
var _ = Suite(&DirSuite{})

func (s *MySuite) TestHelloWorld(c *C) {
	//c.Assert(42, Equals, "42")
	c.Assert(io.ErrClosedPipe, ErrorMatches, "io: .*on closed pipe")
	c.Check(42, Equals, 42)
}

type DirSuite struct {
	dir string
}

func (s *DirSuite) SetUpTest(c *C) {
	s.dir = c.MkDir()
	// Use s.dir to prepare some data.
}

func (s *DirSuite) TestWithDir(c *C) {
	// Use the data in s.dir in the test.
}

// func (s *DirSuite) BenchmarkLogic(c *C) {
// 	for i := 0; i < c.N; i++ {
// 		// Logic to benchmark
// 	}
// }
