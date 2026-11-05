package fs

import "sync"

type inodeID uint64

type inodeKind int

const (
	kindDirectory inodeKind = iota
	kindFile
	kindSymlink
)

type permMode uint8

const (
_PERMRead permMode = 1 << iota
_PERMWrite
)

type inode struct {
	id    inodeID
	kind  inodeKind
	mode  permMode
	owner string
	group string

	mu      sync.RWMutex
	entries map[string]inodeID
	order   []string

	data   []byte
	target string

	writerOpen bool
	readers    int
}
