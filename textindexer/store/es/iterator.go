package es

import (
	"github.com/Vesino/linksRUs/textindexer/index"
	"github.com/elastic/go-elasticsearch"
)

type esIterator struct {
	es         *elasticsearch.Client
	searchReq  map[string]interface{}
	cumIdx     uint64
	rsIdx      int
	rs         *esSearchRes
	latchedDoc *index.Document
	lastErr    error
}
