package extract

import (
	"go/token"
	"go/types"
)

// symbolKey uniquely identifies a symbol for deduplication
type symbolKey struct {
	pkg  string
	name string
	pos  token.Pos
}

// visitedSet tracks visited symbols during traversal
type visitedSet map[symbolKey]bool

// objectInfo wraps types.Object with additional metadata
type objectInfo struct {
	obj   types.Object
	depth int
}
