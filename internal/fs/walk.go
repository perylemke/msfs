package fs

import (
	"errors"
	"path"
)

// SkipDir skips descending into a directory during Walk.
var SkipDir = errors.New("skip directory")

// WalkNode exposes entry metadata to walk callbacks.
type WalkNode struct {
	Kind WalkKind
}

// WalkKind classifies directory entries during a walk.
type WalkKind int

const (
	WalkDirectory WalkKind = iota
	WalkFile
	WalkSymlink
)

// WalkFunc is invoked for each node in the working directory subtree.
type WalkFunc func(path string, node *WalkNode) error

// Walk traverses the working directory subtree depth-first.
func (s *Session) Walk(fn WalkFunc) error {
	root := s.cwd
	s.fs.mu.RLock()
	dir, err := s.fs.resolveAbsLocked(root, true)
	s.fs.mu.RUnlock()
	if err != nil {
		return err
	}
	if dir.inode.kind != kindDirectory {
		return ErrNotDirectory
	}
	return s.walkDir(root, dir.inode, fn)
}

func (s *Session) walkDir(abs string, dir *inode, fn WalkFunc) error {
	s.fs.mu.RLock()
	names := append([]string(nil), dir.order...)
	s.fs.mu.RUnlock()

	for _, name := range names {
		childAbs := path.Join(abs, name)
		s.fs.mu.RLock()
		parent, err := s.fs.resolveAbsLocked(abs, true)
		if err != nil {
			s.fs.mu.RUnlock()
			return err
		}
		id, ok := parent.inode.entries[name]
		if !ok {
			s.fs.mu.RUnlock()
			continue
		}
		n := s.fs.inodes[id]
		n, resolved, err := s.fs.followLinksLocked(n, childAbs)
		s.fs.mu.RUnlock()
		if err != nil {
			return err
		}
		kind := WalkFile
		switch n.kind {
		case kindDirectory:
			kind = WalkDirectory
		case kindSymlink:
			kind = WalkSymlink
		}
		if err := fn(resolved, &WalkNode{Kind: kind}); err != nil {
			if errors.Is(err, SkipDir) && n.kind == kindDirectory {
				continue
			}
			return err
		}
		if n.kind == kindDirectory {
			if err := s.walkDir(resolved, n, fn); err != nil {
				return err
			}
		}
	}
	return nil
}
