package fs

import "path"

func (f *FS) resolveAbs(abs string, followFinal bool) (*resolved, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.resolveAbsLocked(abs, followFinal)
}

func (f *FS) resolveAbsLocked(abs string, followFinal bool) (*resolved, error) {
	return f.resolveAbsLockedSeen(abs, followFinal, make(map[inodeID]struct{}))
}

func (f *FS) resolveAbsLockedSeen(abs string, followFinal bool, seen map[inodeID]struct{}) (*resolved, error) {
	cur := f.root()
	curPath := "/"
	parts := splitAbs(abs)

	for i, part := range parts {
		if cur.kind != kindDirectory {
			return nil, ErrNotDirectory
		}
		id, ok := cur.entries[part]
		if !ok {
			return nil, ErrNotFound
		}
		next := f.inodes[id]
		isLast := i == len(parts)-1
		if isLast && !followFinal && next.kind == kindSymlink {
			return &resolved{inode: next, abs: path.Join(curPath, part)}, nil
		}
		next, linkPath, err := f.followLinksLockedSeen(next, path.Join(curPath, part), seen)
		if err != nil {
			return nil, err
		}
		cur = next
		curPath = linkPath
	}
	return &resolved{inode: cur, abs: curPath}, nil
}

func (f *FS) followLinks(n *inode, linkAbs string) (*inode, string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.followLinksLocked(n, linkAbs)
}

func (f *FS) followLinksLocked(n *inode, linkAbs string) (*inode, string, error) {
	return f.followLinksLockedSeen(n, linkAbs, make(map[inodeID]struct{}))
}

func (f *FS) followLinksLockedSeen(n *inode, linkAbs string, seen map[inodeID]struct{}) (*inode, string, error) {
	for n.kind == kindSymlink {
		if _, ok := seen[n.id]; ok {
			return nil, "", ErrSymlinkLoop
		}
		seen[n.id] = struct{}{}
		target := n.target
		if target == "" {
			return nil, "", ErrNotFound
		}
		res, err := f.resolveAbsLockedSeen(absPath("/", target), true, seen)
		if err != nil {
			return nil, "", err
		}
		n = res.inode
		linkAbs = res.abs
	}
	return n, linkAbs, nil
}

func (f *FS) resolveParent(abs string) (*inode, string, string, error) {
	if abs == "/" {
		return nil, "", "", ErrNotFound
	}
	parentAbs := path.Dir(abs)
	name := path.Base(abs)
	f.mu.RLock()
	defer f.mu.RUnlock()
	res, err := f.resolveAbsLocked(parentAbs, true)
	if err != nil {
		return nil, "", "", err
	}
	if res.inode.kind != kindDirectory {
		return nil, "", "", ErrNotDirectory
	}
	return res.inode, parentAbs, name, nil
}

func (s *Session) resolvePath(p string, followFinal bool) (*resolved, error) {
	abs := absPath(s.cwd, p)
	return s.fs.resolveAbs(abs, followFinal)
}
