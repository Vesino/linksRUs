package index

import (
	"github.com/google/uuid"
)

// Indexer is implemented by objects that can index and search documents
// discovered by the Links 'R' Us crawler.
type Indexer interface {
	// Index inserts a new document to the index or updates the index entry
	// for and existing document.
	Index(doc *Document) error

	// FindByID looks up a document by its link ID.
	FindByID(linkID uuid.UUID) (*Document, error)

	// Search the index for a particular query and return back a result
	// iterator.
	Search(query Query) (Iterator, error)

	// UpdateScore updates the PageRank score for a document with the
	// specified link ID. If no such document exists, a placeholder
	// document with the provided score will be created.
	UpdateScore(linkID uuid.UUID, score float64) error
}

// Iterator is implemented by objects that can paginate search results.
type Iterator interface {
	// Close the iterator and release any allocated resources.
	Close() error

	// Next loads the next document matching the search query.
	// It returns false if no more documents are available.
	Next() bool

	// Error returns the last error encountered by the iterator.
	Error() error

	// Document returns the current document from the result set.
	Document() *Document

	// TotalCount returns the approximate number of search results.
	TotalCount() uint64
}

// QueryType describes the types of queries supported by the indexer
// implementations.
type QueryType uint8

const (
	// QueryTypeMatch requests the indexer to match each expression term.
	QueryTypeMatch QueryType = iota

	// QueryTypePhrase searches for an exact phrase match.
	QueryTypePhrase
)

// Query encapsulates a set of parameters to use when searching indexed
// documents.
type Query struct {
	// The way that the indexer should interpret the search expression.
	Type QueryType

	// The search expression
	Expression string

	// The number of search results to skip.
	Offset uint64
}

// The Expression field stores the search query that's entered by the end user. However, its interpretation by the indexer component can vary, depending on the value of
// the Type attribute. As proof of concept, we will only implement two of the most common types of searches:
// Searching for a list of keywords in any order Searching for an exact phrase match
// In the future, we can opt to add support for other types of queries such as boolean-, date- , or domain-based queries.
