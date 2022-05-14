package index

import "golang.org/x/xerrors"

var (
	// ErrNotFound is returned by the indexer when attempting to look up
	// a document that does not exist.
	ErrNotFound = xerrors.New("not found")

	ErrMissingLinkID = xerrors.New("document does not provide a valid linkID")
)
