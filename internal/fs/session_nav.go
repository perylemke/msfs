package fs

import (
	"path"
)

// Cd changes the working directory to a child or parent.
func (s *Session) Cd(p string) error {
	if p == ".." {
		if s.cwd == "/" {
			return nil
		}
		s.cwd = path.Dir(s.cwd)
		return nil
	}
	if p == "." || p == "" {
		return nil
	}
	abs := absPath(s.cwd, p)
	res, err := s.fs.resolveAbs(abs, true)
	if err != nil {
		return err
	}
	if res.inode.kind != kindDirectory {
		return ErrNotDirectory
	}
	s.cwd = res.abs
	return nil
}

// Mkdir creates a directory under the current working directory.
func (s *Session) Mkdir(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	if err := s.fs.runBefore(EventCreate, path.Join(s.cwd, name)); err != nil {
		return err
	}
	s.fs.mu.Lock()
	parent, err := s.fs.resolveAbsLocked(s.cwd, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if _, ok := parent.inode.entries[name]; ok {
		s.fs.mu.Unlock()
		return ErrAlreadyExists
	}
	child := s.fs.newInode(kindDirectory)
	child.owner = parent.inode.owner
	child.group = parent.inode.group
	child.mode = parent.inode.mode
	s.fs.inodes[child.id] = child
	parent.inode.entries[name] = child.id
	parent.inode.order = append(parent.inode.order, name)
	created := path.Join(s.cwd, name)
	s.fs.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventCreate, Path: created})
	return nil
}

// List returns names of children in the working directory.
func (s *Session) List() ([]string, error) {
	s.fs.mu.RLock()
	defer s.fs.mu.RUnlock()
	dir, err := s.fs.resolveAbsLocked(s.cwd, true)
	if err != nil {
		return nil, err
	}
	if dir.inode.kind != kindDirectory {
		return nil, ErrNotDirectory
	}
	out := make([]string, len(dir.inode.order))
	copy(out, dir.inode.order)
	return out, nil
}

// Rmdir removes a directory that is a direct child of the working directory.
func (s *Session) Rmdir(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	targetAbs := path.Join(s.cwd, name)
	if err := s.fs.runBefore(EventRemove, targetAbs); err != nil {
		return err
	}
	s.fs.mu.Lock()
	parent, err := s.fs.resolveAbsLocked(s.cwd, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	id, ok := parent.inode.entries[name]
	if !ok {
		s.fs.mu.Unlock()
		return ErrNotChild
	}
	n := s.fs.inodes[id]
	n, _, err = s.fs.followLinksLocked(n, targetAbs)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if n.kind != kindDirectory {
		s.fs.mu.Unlock()
		return ErrNotDirectory
	}
	if len(n.entries) > 0 {
		s.fs.mu.Unlock()
		return ErrDirectoryNotEmpty
	}
	delete(parent.inode.entries, name)
	parent.inode.order = removeName(parent.inode.order, name)
	delete(s.fs.inodes, n.id)
	s.fs.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventRemove, Path: targetAbs})
	return nil
}

func removeName(order []string, name string) []string {
	out := order[:0]
	for _, n := range order {
		if n != name {
			out = append(out, n)
		}
	}
	return out
}
