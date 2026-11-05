package fs

import (
	"fmt"
	"path"
)

// MkdirAll creates a directory and any missing parents.
func (s *Session) MkdirAll(p string) error {
	abs := absPath(s.cwd, p)
	if abs == "/" {
		return nil
	}
	parts := splitAbs(abs)
	cur := "/"
	for _, part := range parts {
		cur = path.Join(cur, part)
		if err := s.ensureDir(cur); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) ensureDir(abs string) error {
	name := path.Base(abs)
	parentAbs := path.Dir(abs)
	if err := s.fs.runBefore(EventCreate, abs); err != nil {
		return err
	}
	s.fs.mu.Lock()
	parent, err := s.fs.resolveAbsLocked(parentAbs, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if parent.inode.kind != kindDirectory {
		s.fs.mu.Unlock()
		return ErrNotDirectory
	}
	if _, ok := parent.inode.entries[name]; ok {
		s.fs.mu.Unlock()
		return nil
	}
	child := s.fs.newInode(kindDirectory)
	s.fs.inodes[child.id] = child
	parent.inode.entries[name] = child.id
	parent.inode.order = append(parent.inode.order, name)
	s.fs.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventCreate, Path: abs})
	return nil
}

// CreateFileAt creates an empty file at an absolute or relative path.
func (s *Session) CreateFileAt(p string) error {
	abs := absPath(s.cwd, p)
	name := path.Base(abs)
	parentAbs := path.Dir(abs)
	if err := validateName(name); err != nil {
		return err
	}
	if err := s.fs.runBefore(EventCreate, abs); err != nil {
		return err
	}
	s.fs.mu.Lock()
	parent, err := s.fs.resolveAbsLocked(parentAbs, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if _, ok := parent.inode.entries[name]; ok {
		s.fs.mu.Unlock()
		return ErrAlreadyExists
	}
	f := s.fs.newInode(kindFile)
	s.fs.inodes[f.id] = f
	parent.inode.entries[name] = f.id
	parent.inode.order = append(parent.inode.order, name)
	s.fs.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventCreate, Path: abs})
	return nil
}

// WriteFileAt writes content to a file at a path.
func (s *Session) WriteFileAt(p, content string) error {
	abs := absPath(s.cwd, p)
	res, err := s.fs.resolveAbs(abs, true)
	if err != nil {
		return err
	}
	if res.inode.kind != kindFile {
		return ErrIsDirectory
	}
	if err := s.checkPerm(res.inode, _PERMWrite); err != nil {
		return err
	}
	if err := s.fs.runBefore(EventWrite, abs); err != nil {
		return err
	}
	res.inode.mu.Lock()
	res.inode.data = []byte(content)
	res.inode.mu.Unlock()
	s.fs.runAfter(Event{Kind: EventWrite, Path: abs})
	return nil
}

// ReadFileAt reads a file at a path.
func (s *Session) ReadFileAt(p string) (string, error) {
	abs := absPath(s.cwd, p)
	res, err := s.fs.resolveAbs(abs, true)
	if err != nil {
		return "", err
	}
	if res.inode.kind != kindFile {
		return "", ErrIsDirectory
	}
	if err := s.checkPerm(res.inode, _PERMRead); err != nil {
		return "", err
	}
	res.inode.mu.RLock()
	defer res.inode.mu.RUnlock()
	return string(res.inode.data), nil
}

// Move moves a file or directory to a new path.
func (s *Session) Move(src, dst string) error {
	return s.transfer(src, dst, true)
}

// Copy copies a file or directory to a new path.
func (s *Session) Copy(src, dst string) error {
	return s.transfer(src, dst, false)
}

func (s *Session) transfer(src, dst string, move bool) error {
	srcAbs := absPath(s.cwd, src)
	dstAbs := absPath(s.cwd, dst)
	srcRes, err := s.fs.resolveAbs(srcAbs, true)
	if err != nil {
		return err
	}
	dstName := path.Base(dstAbs)
	dstParentAbs := path.Dir(dstAbs)

	s.fs.mu.Lock()
	dstParent, err := s.fs.resolveAbsLocked(dstParentAbs, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if existingID, ok := dstParent.inode.entries[dstName]; ok {
		existing := s.fs.inodes[existingID]
		if srcRes.inode.kind == kindDirectory && existing.kind == kindDirectory {
			s.fs.mu.Unlock()
			if move {
				if err := s.mergeDirectory(srcAbs, dstAbs); err != nil {
					return err
				}
				return s.removePath(srcAbs)
			}
			return s.copyDirectoryContents(srcAbs, dstAbs)
		}
		if srcRes.inode.kind == kindFile && existing.kind == kindFile {
			newName, err := s.autoRename(dstParent.inode, dstName)
			if err != nil {
				s.fs.mu.Unlock()
				return err
			}
			dstName = newName
			dstAbs = path.Join(dstParentAbs, dstName)
		} else {
			s.fs.mu.Unlock()
			return ErrTypeMismatch
		}
	}
	s.fs.mu.Unlock()

	if move {
		return s.linkMove(srcAbs, dstAbs)
	}
	return s.linkCopy(srcAbs, dstAbs)
}

func (s *Session) autoRename(dir *inode, name string) (string, error) {
	for i := 1; i < 1000; i++ {
		candidate := fmt.Sprintf("%s~%d", name, i)
		if _, ok := dir.entries[candidate]; !ok {
			return candidate, nil
		}
	}
	return "", ErrAlreadyExists
}

func (s *Session) linkMove(srcAbs, dstAbs string) error {
	srcParentAbs := path.Dir(srcAbs)
	srcName := path.Base(srcAbs)
	dstName := path.Base(dstAbs)
	dstParentAbs := path.Dir(dstAbs)

	s.fs.mu.Lock()
	defer s.fs.mu.Unlock()
	srcParent, err := s.fs.resolveAbsLocked(srcParentAbs, true)
	if err != nil {
		return err
	}
	id, ok := srcParent.inode.entries[srcName]
	if !ok {
		return ErrNotFound
	}
	dstParent, err := s.fs.resolveAbsLocked(dstParentAbs, true)
	if err != nil {
		return err
	}
	if _, ok := dstParent.inode.entries[dstName]; ok {
		return ErrAlreadyExists
	}
	dstParent.inode.entries[dstName] = id
	dstParent.inode.order = append(dstParent.inode.order, dstName)
	delete(srcParent.inode.entries, srcName)
	srcParent.inode.order = removeName(srcParent.inode.order, srcName)
	return nil
}

func (s *Session) linkCopy(srcAbs, dstAbs string) error {
	srcRes, err := s.fs.resolveAbs(srcAbs, true)
	if err != nil {
		return err
	}
	switch srcRes.inode.kind {
	case kindFile:
		if err := s.CreateFileAt(dstAbs); err != nil {
			return err
		}
		data, err := s.ReadFileAt(srcAbs)
		if err != nil {
			return err
		}
		return s.WriteFileAt(dstAbs, data)
	case kindDirectory:
		if err := s.MkdirAll(dstAbs); err != nil {
			return err
		}
		return s.copyDirectoryContents(srcAbs, dstAbs)
	case kindSymlink:
		s.fs.mu.RLock()
		target := srcRes.inode.target
		s.fs.mu.RUnlock()
		return s.Symlink(target, dstAbs)
	default:
		return ErrNotFound
	}
}

func (s *Session) copyDirectoryContents(srcAbs, dstAbs string) error {
	s.fs.mu.RLock()
	srcDir, err := s.fs.resolveAbsLocked(srcAbs, true)
	if err != nil {
		s.fs.mu.RUnlock()
		return err
	}
	names := append([]string(nil), srcDir.inode.order...)
	s.fs.mu.RUnlock()
	for _, name := range names {
		childSrc := path.Join(srcAbs, name)
		childDst := path.Join(dstAbs, name)
		if err := s.Copy(childSrc, childDst); err != nil {
			return err
		}
	}
	return nil
}

func (s *Session) mergeDirectory(srcAbs, dstAbs string) error {
	return s.copyDirectoryContents(srcAbs, dstAbs)
}

func (s *Session) removePath(abs string) error {
	res, err := s.fs.resolveAbs(abs, true)
	if err != nil {
		return err
	}
	if res.inode.kind == kindDirectory {
		s.fs.mu.RLock()
		names := append([]string(nil), res.inode.order...)
		s.fs.mu.RUnlock()
		for _, name := range names {
			if err := s.removePath(path.Join(abs, name)); err != nil {
				return err
			}
		}
	}
	return s.unlinkAbs(abs)
}

func (s *Session) unlinkAbs(abs string) error {
	name := path.Base(abs)
	parentAbs := path.Dir(abs)
	s.fs.mu.Lock()
	parent, err := s.fs.resolveAbsLocked(parentAbs, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	id, ok := parent.inode.entries[name]
	if !ok {
		s.fs.mu.Unlock()
		return ErrNotFound
	}
	n := s.fs.inodes[id]
	delete(parent.inode.entries, name)
	parent.inode.order = removeName(parent.inode.order, name)
	if n.kind != kindDirectory {
		delete(s.fs.inodes, n.id)
	} else if len(n.entries) == 0 {
		delete(s.fs.inodes, n.id)
	}
	s.fs.mu.Unlock()
	return nil
}

// Symlink creates a symbolic link at linkPath pointing to target.
func (s *Session) Symlink(target, linkPath string) error {
	abs := absPath(s.cwd, linkPath)
	name := path.Base(abs)
	parentAbs := path.Dir(abs)
	if err := validateName(name); err != nil {
		return err
	}
	s.fs.mu.Lock()
	parent, err := s.fs.resolveAbsLocked(parentAbs, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if _, ok := parent.inode.entries[name]; ok {
		s.fs.mu.Unlock()
		return ErrAlreadyExists
	}
	link := s.fs.newInode(kindSymlink)
	link.target = target
	s.fs.inodes[link.id] = link
	parent.inode.entries[name] = link.id
	parent.inode.order = append(parent.inode.order, name)
	s.fs.mu.Unlock()
	return nil
}

// Hardlink creates another name for an existing file or directory.
func (s *Session) Hardlink(existingPath, linkPath string) error {
	srcAbs := absPath(s.cwd, existingPath)
	linkAbs := absPath(s.cwd, linkPath)
	src, err := s.fs.resolveAbs(srcAbs, true)
	if err != nil {
		return err
	}
	name := path.Base(linkAbs)
	parentAbs := path.Dir(linkAbs)
	s.fs.mu.Lock()
	parent, err := s.fs.resolveAbsLocked(parentAbs, true)
	if err != nil {
		s.fs.mu.Unlock()
		return err
	}
	if _, ok := parent.inode.entries[name]; ok {
		s.fs.mu.Unlock()
		return ErrAlreadyExists
	}
	parent.inode.entries[name] = src.inode.id
	parent.inode.order = append(parent.inode.order, name)
	s.fs.mu.Unlock()
	return nil
}

// FindFirstRegex returns the first file path under cwd whose base name matches re.
func (s *Session) FindFirstRegex(re string) (string, error) {
	pattern, err := compileRegex(re)
	if err != nil {
		return "", err
	}
	var found string
	err = s.Walk(func(p string, n *WalkNode) error {
		if n.Kind != WalkFile {
			return nil
		}
		if pattern.MatchString(path.Base(p)) {
			found = p
			return errStopWalk
		}
		return nil
	})
	if err == errStopWalk {
		return found, nil
	}
	if err != nil {
		return "", err
	}
	return "", ErrNotFound
}

var errStopWalk = fmt.Errorf("stop")

func compileRegex(re string) (matcher, error) {
	return regexpCompile(re)
}

type matcher interface {
	MatchString(string) bool
}

// implemented in regex.go using regexp

// Chmod sets read/write mode bits on a path (owner only).
func (s *Session) Chmod(p string, read, write bool) error {
	abs := absPath(s.cwd, p)
	res, err := s.fs.resolveAbs(abs, true)
	if err != nil {
		return err
	}
	if s.user != res.inode.owner {
		return ErrPermission
	}
	var mode permMode
	if read {
		mode |= _PERMRead
	}
	if write {
		mode |= _PERMWrite
	}
	res.inode.mode = mode
	return nil
}

// AddUser registers a user and optional groups.
func (f *FS) AddUser(name string, groups []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.users[name] = &User{Name: name, Groups: groups}
	for _, g := range groups {
		if gr, ok := f.groups[g]; ok {
			gr.Members = append(gr.Members, name)
		} else {
			f.groups[g] = &Group{Name: g, Members: []string{name}}
		}
	}
}

// SwitchUser changes the session user (no permission check).
func (s *Session) SwitchUser(name string) error {
	s.fs.mu.RLock()
	_, ok := s.fs.users[name]
	s.fs.mu.RUnlock()
	if !ok {
		return ErrNotFound
	}
	s.user = name
	return nil
}

// FileHandle supports chunked read/write with seek.
type FileHandle struct {
	inode  *inode
	offset int64
	write  bool
}

// Open opens a file for reading or writing.
func (s *Session) Open(p string, write bool) (*FileHandle, error) {
	abs := absPath(s.cwd, p)
	res, err := s.fs.resolveAbs(abs, true)
	if err != nil {
		return nil, err
	}
	if res.inode.kind != kindFile {
		return nil, ErrIsDirectory
	}
	want := _PERMRead
	if write {
		want = _PERMWrite
	}
	if err := s.checkPerm(res.inode, want); err != nil {
		return nil, err
	}
	res.inode.mu.Lock()
	if write {
		if res.inode.writerOpen {
			res.inode.mu.Unlock()
			return nil, ErrWriterOpen
		}
		res.inode.writerOpen = true
	}
	res.inode.readers++
	res.inode.mu.Unlock()
	return &FileHandle{inode: res.inode, write: write}, nil
}

// Read reads up to len(b) bytes from the handle.
func (h *FileHandle) Read(b []byte) (int, error) {
	h.inode.mu.RLock()
	defer h.inode.mu.RUnlock()
	if h.offset >= int64(len(h.inode.data)) {
		return 0, nil
	}
	n := copy(b, h.inode.data[h.offset:])
	h.offset += int64(n)
	return n, nil
}

// Write writes bytes at the current offset.
func (h *FileHandle) Write(b []byte) (int, error) {
	if !h.write {
		return 0, ErrPermission
	}
	h.inode.mu.Lock()
	defer h.inode.mu.Unlock()
	end := h.offset + int64(len(b))
	if end > int64(len(h.inode.data)) {
		need := make([]byte, end)
		copy(need, h.inode.data)
		h.inode.data = need
	}
	n := copy(h.inode.data[h.offset:], b)
	h.offset += int64(n)
	return n, nil
}

// Seek sets the offset; whence 0=start,1=current,2=end.
func (h *FileHandle) Seek(offset int64, whence int) (int64, error) {
	h.inode.mu.RLock()
	size := int64(len(h.inode.data))
	h.inode.mu.RUnlock()
	switch whence {
	case 0:
		h.offset = offset
	case 1:
		h.offset += offset
	case 2:
		h.offset = size + offset
	default:
		return 0, ErrInvalidName
	}
	if h.offset < 0 {
		h.offset = 0
	}
	return h.offset, nil
}

// Close releases the handle.
func (h *FileHandle) Close() error {
	h.inode.mu.Lock()
	defer h.inode.mu.Unlock()
	if h.write {
		h.inode.writerOpen = false
	}
	if h.inode.readers > 0 {
		h.inode.readers--
	}
	return nil
}

// NewSessionFrom creates another session on the same FS.
func NewSessionFrom(f *FS) *Session {
	return NewSession(f)
}
