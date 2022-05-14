package memory

import (
	"testing"

	"github.com/Vesino/linksRUs/textindexer/index/indextest"
	gc "gopkg.in/check.v1"
)

type InMemoryBleveTestSuite struct {
	indextest.SuiteBase
	idx *InMemoryBleveIndexer
}

var _ = gc.Suite(new(InMemoryBleveIndexer))

// Register our test-suite with go test.
func Test(t *testing.T) { gc.TestingT(t) }

func (s *InMemoryBleveTestSuite) SetUpTest(c *gc.C) {
	idx, err := NewInMemoryBleveIndexer()
	c.Assert(err, gc.IsNil)
	s.SetIndexer(idx)

	// Keep track of the concrete indexer implementation so we can clean up
	// when tearing down the test
	s.idx = idx
}

func (s *InMemoryBleveTestSuite) TearDownTest(c *gc.C) {
	c.Assert(s.idx.Close(), gc.IsNil)
}
