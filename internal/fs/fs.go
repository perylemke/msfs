package fs

import "sync"

// FS is the shared in-memory filesystem store.
type FS struct {
	mu     sync.RWMutex
	inodes map[inodeID]*inode
	nextID inodeID

	users  map[string]*User
	groups map[string]*Group

	before []BeforeHook
	after  []AfterHook
}

// New creates an empty filesystem with root / owned by user "user".
func New() *FS {
	f := &FS{
		inodes: make(map[inodeID]*inode),
		users:  make(map[string]*User),
		groups: make(map[string]*Group),
	}
	root := f.newInode(kindDirectory)
	root.owner = "user"
	root.group = "user"
	root.mode = _PERMRead | _PERMWrite
	f.inodes[root.id] = root

	f.users["user"] = &User{Name: "user", Groups: []string{"user"}}
	f.groups["user"] = &Group{Name: "user", Members: []string{"user"}}

	return f
}

func (f *FS) newInode(kind inodeKind) *inode {
	f.nextID++
	n := &inode{
		id:      f.nextID,
		kind:    kind,
		mode:    _PERMRead | _PERMWrite,
		owner:   "user",
		group:   "user",
		entries: make(map[string]inodeID),
	}
	return n
}

func (f *FS) root() *inode {
	return f.inodes[1]
}
