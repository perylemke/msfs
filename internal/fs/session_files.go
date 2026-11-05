package fs

import "path"

func (s *Session) entryInCWD(name string) (*inode, error) {
	if err := validateName(name); err != nil {
		return nil, err
	}
	s.fs.mu.RLock()
	defer s.fs.mu.RUnlock()
	dir, err := s.fs.resolveAbsLocked(s.cwd, true)
	if err != nil {
		return nil, err
	}
	id, ok := dir.inode.entries[name]
	if !ok {
		return nil, ErrNotFound
	}
	n := s.fs.inodes[id]
	n, _, err = s.fs.followLinksLocked(n, path.Join(s.cwd, name))
	return n, err
}

// CreateFile creates an empty file in the working directory.
func (s *Session) CreateFile(name string) error {
	if err := validateName(name); err != nil {
		return err
	}
	target := path.Join(s.cwd, name)
	if err := s.fs.runBefore(EventCreate, target); err != nil {
		return err
	}
	s.fs.mu.Lock()
	dir, err := s.fs.resolveAbsLocked(s.cwd, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if err := s.checkPerm(dir.inode, _PERMWrite); err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if _, ok := dir.inode.entries[name]; ok {
		s.fs.mu.Unlock()
		return ErrAlreadyExists
	}
	f := s.fs.newInode(kindFile)
	s.fs.inodes[f.id] = f
	dir.inode.entries[name] = f.id
	dir.inode.order = append(dir.inode.order, name)
	s.fs.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventCreate, Path: target})
	return nil
}

// WriteFile replaces the contents of a file in the working directory.
func (s *Session) WriteFile(name, content string) error {
	n, err := s.entryInCWD(name)
	if err != nil {
		return err
	}
	if n.kind == kindDirectory {
		return ErrIsDirectory
	}
	if err := s.checkPerm(n, _PERMWrite); err != nil {
		return err
	}
	target := path.Join(s.cwd, name)
	if err := s.fs.runBefore(EventWrite, target); err != nil {
		return err
	}
	n.mu.Lock()
	n.data = []byte(content)
	n.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventWrite, Path: target})
	return nil
}

// ReadFile returns file contents from the working directory.
func (s *Session) ReadFile(name string) (string, error) {
	n, err := s.entryInCWD(name)
	if err != nil {
		return "", err
	}
	if n.kind == kindDirectory {
		return "", ErrIsDirectory
	}
	if err := s.checkPerm(n, _PERMRead); err != nil {
		return "", err
	}
	n.mu.RLock()
	defer n.mu.RUnlock()
	return string(n.data), nil
}

// MoveFile renames a file within the working directory.
func (s *Session) MoveFile(from, to string) error {
	if err := validateName(from); err != nil {
		return err
	}
	if err := validateName(to); err != nil {
		return err
	}
	fromPath := path.Join(s.cwd, from)
	toPath := path.Join(s.cwd, to)
	if err := s.fs.runBefore(EventRename, fromPath); err != nil {
		return err
	}
	s.fs.mu.Lock()
	dir, err := s.fs.resolveAbsLocked(s.cwd, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	id, ok := dir.inode.entries[from]
	if !ok {
		s.fs.mu.Unlock()
		return ErrNotFound
	}
	n := s.fs.inodes[id]
	if n.kind != kindFile {
		s.fs.mu.Unlock()
		return ErrIsDirectory
	}
	if _, ok := dir.inode.entries[to]; ok {
		s.fs.mu.Unlock()
		return ErrAlreadyExists
	}
	dir.inode.entries[to] = id
	delete(dir.inode.entries, from)
	dir.inode.order = replaceName(dir.inode.order, from, to)
	s.fs.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventRename, Path: toPath})
	_ = fromPath
	return nil
}

func replaceName(order []string, from, to string) []string {
	out := make([]string, len(order))
	for i, n := range order {
		if n == from {
			out[i] = to
		} else {
			out[i] = n
		}
	}
	return out
}

// Find returns absolute paths under the working directory subtree with an exact name match.
func (s *Session) Find(name string) ([]string, error) {
	var out []string
	err := s.Walk(func(p string, n *WalkNode) error {
		if path.Base(p) == name {
			out = append(out, p)
		}
		return nil
	})
	return out, err
}
