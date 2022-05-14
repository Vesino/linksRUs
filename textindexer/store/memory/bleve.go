package memory

import (
	"sync"
	"time"

	"github.com/Vesino/linksRUs/textindexer/index"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/search/query"
	"github.com/google/uuid"
	"golang.org/x/xerrors"
)

// The size of each page of results that is cached locally by the iterator.
const bacthSize = 10

// Compile-time check to ensure InMemoryBleveIndexer implements Indexer.
var _ index.Indexer = (*InMemoryBleveIndexer)(nil)

// An obvious caveat of this approach is that since bleve stores a partial view of the document data, we cannot recreate the original document from the result
// list returned by bleve after executing a search query. To solve this problem, the in-memory indexer maintains a map where keys are the document link IDs and v
// alues are immutable copies of the documents that are processed by the indexer. When processing a result list, the returned document IDs are used to index the map
// and to recover the original document. To ensure that the in- memory indexer is safe for concurrent use, access to the map is guarded with a read/write mutex.
type bleveDoc struct {
	Title    string
	Content  string
	PageRank float64
}

// InMemoryBleveIndexer is an Indexer implementation that uses an in-memory
// bleve instance to catalogue and search documents.
type InMemoryBleveIndexer struct {
	mu   sync.RWMutex
	docs map[string]*index.Document

	idx bleve.Index
}

// InMemoryBleveIndexer is an Indexer implementation that uses an in-memory
// bleve instance to catalogue and search documents.
func NewInMemoryBleveIndexer() (*InMemoryBleveIndexer, error) {
	mapping := bleve.NewIndexMapping()
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}
	return &InMemoryBleveIndexer{
		idx:  idx,
		docs: make(map[string]*index.Document),
	}, nil
}

// Close the indexer and release any allocated resources.
func (i *InMemoryBleveIndexer) Close() error {
	return i.idx.Close()
}

// Index inserts a new document to the index or updates the index entry
// for and existing document.
func (i *InMemoryBleveIndexer) Index(doc *index.Document) error {
	if doc.LinkID == uuid.Nil {
		return xerrors.Errorf("index: %v", index.ErrMissingLinkID)
	}
	doc.IndexedAt = time.Now()
	dcopy := copyDoc(doc)
	key := dcopy.LinkID.String()

	i.mu.Lock()
	// If updating, preserve existing PageRank score
	if orig, exists := i.docs[key]; exists {
		dcopy.PageRank = orig.PageRank
	}
	if err := i.idx.Index(key, makeBlevedoc(doc)); err != nil {
		return xerrors.Errorf("update score: %w", err)
	}
	i.docs[key] = dcopy
	i.mu.Unlock()
	return nil
}

func (i *InMemoryBleveIndexer) FindByID(linkID uuid.UUID) (*index.Document, error) {
	return i.findById(linkID.String())
}

func (i *InMemoryBleveIndexer) findById(linkID string) (*index.Document, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if d, found := i.docs[linkID]; found {
		return copyDoc(d), nil
	}

	return nil, xerrors.Errorf("find by ID: %w", index.ErrNotFound)
}

func (i *InMemoryBleveIndexer) Search(q index.Query) (index.Iterator, error) {
	var bq query.Query
	switch q.Type {
	case index.QueryTypePhrase:
		bq = bleve.NewMatchPhraseQuery(q.Expression)
	default:
		bq = bleve.NewMatchQuery(q.Expression)
	}

	searchReq := bleve.NewSearchRequest(bq)
	searchReq.SortBy([]string{"-PageRank", "-_score"})
	searchReq.Size = bacthSize
	searchReq.From = int(q.Offset)
	rs, err := i.idx.Search(searchReq)
	if err != nil {
		return nil, xerrors.Errorf("search: %w", err)
	}
	return &bleveIterator{idx: i, searchReq: searchReq, rs: rs, cumIdx: q.Offset}, nil
}

// UpdateScore updates the PageRank score for a document with the specified
// link ID. If no such document exists, a placeholder document with the
// provided score will be created.
func (i *InMemoryBleveIndexer) UpdateScore(linkID uuid.UUID, score float64) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	key := linkID.String()
	doc, found := i.docs[key]
	if !found {
		doc = &index.Document{LinkID: linkID}
		i.docs[key] = doc
	}
	doc.PageRank = score
	if err := i.idx.Index(key, makeBlevedoc(doc)); err != nil {
		return xerrors.Errorf("update score: %w", err)
	}
	return nil
}

func copyDoc(d *index.Document) *index.Document {
	dcopy := new(index.Document)
	*dcopy = *d
	return dcopy
}

func makeBlevedoc(doc *index.Document) bleveDoc {
	return bleveDoc{
		Title:    doc.Title,
		Content:  doc.Content,
		PageRank: doc.PageRank,
	}
}
