package fs

import "testing"

func TestResolvePathAndParent(t *testing.T) {
	s := NewSession(New())
	if err := s.MkdirAll("/a/b"); err != nil {
		t.Fatal(err)
	}
	if err := s.Cd("/a"); err != nil {
		t.Fatal(err)
	}
	res, err := s.resolvePath("b", true)
	if err != nil {
		t.Fatal(err)
	}
	if res.abs != "/a/b" {
		t.Fatalf("got %s", res.abs)
	}
	_, parent, name, err := s.fs.resolveParent("/a/b")
	if err != nil || parent != "/a" || name != "b" {
		t.Fatalf("parent=%q name=%q err=%v", parent, name, err)
	}
	_, _, _, err = s.fs.resolveParent("/no/such/path")
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestEnsureDirParentMissing(t *testing.T) {
	s := NewSession(New())
	err := s.ensureDir("/no/parent/here")
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestFollowLinksPublic(t *testing.T) {
	s := NewSession(New())
	if err := s.CreateFileAt("/f"); err != nil {
		t.Fatal(err)
	}
	res, err := s.fs.resolveAbs("/f", true)
	if err != nil {
		t.Fatal(err)
	}
	n, _, err := s.fs.followLinks(res.inode, "/f")
	if err != nil || n == nil {
		t.Fatalf("err=%v", err)
	}
}

func TestWalkCwdIsFile(t *testing.T) {
	s := NewSession(New())
	if err := s.CreateFileAt("/f"); err != nil {
		t.Fatal(err)
	}
	s.cwd = "/f"
	err := s.Walk(func(string, *WalkNode) error { return nil })
	if err != ErrNotDirectory {
		t.Fatalf("got %v", err)
	}
}

func TestUnlinkAbsMissing(t *testing.T) {
	s := NewSession(New())
	if err := s.unlinkAbs("/nope"); err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestListCwdMissing(t *testing.T) {
	s := NewSession(New())
	s.cwd = "/missing"
	_, err := s.List()
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestLinkMoveDestExists(t *testing.T) {
	s := NewSession(New())
	if err := s.MkdirAll("/p"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateFileAt("/p/a"); err != nil {
		t.Fatal(err)
	}
	if err := s.CreateFileAt("/p/b"); err != nil {
		t.Fatal(err)
	}
	err := s.linkMove("/p/a", "/p/b")
	if err != ErrAlreadyExists {
		t.Fatalf("got %v", err)
	}
}

func TestListCwdNotDirectory(t *testing.T) {
	s := NewSession(New())
	if err := s.CreateFileAt("/f"); err != nil {
		t.Fatal(err)
	}
	s.cwd = "/f"
	_, err := s.List()
	if err != ErrNotDirectory {
		t.Fatalf("got %v", err)
	}
}

func TestResolveAbsLockedSeen(t *testing.T) {
	s := NewSession(New())
	if err := s.CreateFileAt("/f"); err != nil {
		t.Fatal(err)
	}
	_, err := s.fs.resolveAbsLockedSeen("/f/nope", true, make(map[inodeID]struct{}))
	if err != ErrNotDirectory {
		t.Fatalf("got %v", err)
	}
}

func TestResolveParentRoot(t *testing.T) {
	s := NewSession(New())
	_, _, _, err := s.fs.resolveParent("/")
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestResolveParentFileParent(t *testing.T) {
	s := NewSession(New())
	if err := s.CreateFileAt("/f"); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := s.fs.resolveParent("/f/x")
	if err != ErrNotDirectory {
		t.Fatalf("got %v", err)
	}
}

func TestAbsPathVariants(t *testing.T) {
	if got := absPath("/a/b", ".."); got != "/a" {
		t.Fatalf("got %s", got)
	}
	if got := absPath("/a", "/z"); got != "/z" {
		t.Fatalf("got %s", got)
	}
	if got := absPath("/a", ""); got != "/a" {
		t.Fatalf("got %s", got)
	}
}

func TestWalkCwdNotDirectory(t *testing.T) {
	s := NewSession(New())
	s.cwd = "/not-there"
	err := s.Walk(func(string, *WalkNode) error { return nil })
	if err != ErrNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestResolveAbsLockedSeenSymlinkFinal(t *testing.T) {
	s := NewSession(New())
	if err := s.Symlink("/missing", "/link"); err != nil {
		t.Fatal(err)
	}
	seen := make(map[inodeID]struct{})
	res, err := s.fs.resolveAbsLockedSeen("/link", false, seen)
	if err != nil {
		t.Fatal(err)
	}
	if res.inode.kind != kindSymlink {
		t.Fatalf("kind %v", res.inode.kind)
	}
}
