package fs_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestGroupMemberReadWrite(t *testing.T) {
	f := fs.New()
	f.AddUser("bob", []string{"user"})
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("shared"))
	require.NoError(t, s.WriteFile("shared", "ok"))
	require.NoError(t, s.SwitchUser("bob"))
	got, err := s.ReadFile("shared")
	require.NoError(t, err)
	require.Equal(t, "ok", got)
	require.NoError(t, s.WriteFile("shared", "bob"))
}

func TestAddUserCreatesGroup(t *testing.T) {
	f := fs.New()
	f.AddUser("alice", []string{"devs"})
	s := fs.NewSession(f)
	require.NoError(t, s.SwitchUser("alice"))
}

func TestFileHandleWriteSeekRead(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	h, err := s.Open("f", true)
	require.NoError(t, err)
	n, err := h.Write([]byte("hello"))
	require.NoError(t, err)
	require.Equal(t, 5, n)
	_, err = h.Seek(0, 0)
	require.NoError(t, err)
	buf := make([]byte, 5)
	n, err = h.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 5, n)
	require.Equal(t, "hello", string(buf))
	require.NoError(t, h.Close())
}

func TestMultipleReadHandles(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.WriteFile("f", "x"))
	h1, err := s.Open("f", false)
	require.NoError(t, err)
	h2, err := s.Open("f", false)
	require.NoError(t, err)
	require.NoError(t, h1.Close())
	require.NoError(t, h2.Close())
}

func TestCopySymlink(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFileAt("/target"))
	require.NoError(t, s.Symlink("/target", "/link"))
	require.NoError(t, s.Copy("/link", "/link2"))
	_, err := s.ReadFileAt("/link2")
	require.NoError(t, err)
}

func TestCreateFileDuplicate(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("a"))
	err := s.CreateFile("a")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
}

func TestCreateFileAtDuplicate(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFileAt("/a"))
	err := s.CreateFileAt("/a")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
}

func TestMoveFileNotFound(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := s.MoveFile("a", "b")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestMoveFileIsDirectory(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	err := s.MoveFile("d", "x")
	require.ErrorIs(t, err, fs.ErrIsDirectory)
}

func TestMkdirAllIdempotent(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/a/b"))
	require.NoError(t, s.MkdirAll("/a/b"))
}

func TestCopyFileSimple(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFileAt("/src"))
	require.NoError(t, s.WriteFileAt("/src", "data"))
	require.NoError(t, s.Copy("/src", "/dst"))
	got, err := s.ReadFileAt("/dst")
	require.NoError(t, err)
	require.Equal(t, "data", got)
}

func TestMovePathLinkMove(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/a"))
	require.NoError(t, s.MkdirAll("/b"))
	require.NoError(t, s.CreateFileAt("/a/f"))
	require.NoError(t, s.Move("/a/f", "/b/f"))
	_, err := s.ReadFileAt("/b/f")
	require.NoError(t, err)
}

func TestWalkVisitsFiles(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/w/sub"))
	require.NoError(t, s.CreateFileAt("/w/sub/f"))
	var count int
	require.NoError(t, s.Walk(func(_ string, n *fs.WalkNode) error {
		if n.Kind == fs.WalkFile {
			count++
		}
		return nil
	}))
	require.GreaterOrEqual(t, count, 1)
}

func TestRemoveMissing(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := s.Remove("/nope")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestRemoveAllMissing(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := s.RemoveAll("/nope")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestHardlinkDuplicate(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("a"))
	require.NoError(t, s.Hardlink("a", "b"))
	err := s.Hardlink("a", "b")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
}

func TestSymlinkDuplicate(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Symlink("/t", "/l"))
	err := s.Symlink("/t2", "/l")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
}

func TestReadFileAtDirectory(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	_, err := s.ReadFileAt("d")
	require.ErrorIs(t, err, fs.ErrIsDirectory)
}

func TestWriteFileAtDirectory(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	err := s.WriteFileAt("d", "x")
	require.ErrorIs(t, err, fs.ErrIsDirectory)
}

func TestSeekWhenceCurrent(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.WriteFile("f", "abcd"))
	h, err := s.Open("f", false)
	require.NoError(t, err)
	_, _ = h.Read(make([]byte, 2))
	pos, err := h.Seek(1, 1)
	require.NoError(t, err)
	require.Equal(t, int64(3), pos)
	require.NoError(t, h.Close())
}

func TestSeekInvalidWhence(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	h, err := s.Open("f", false)
	require.NoError(t, err)
	_, err = h.Seek(0, 9)
	require.ErrorIs(t, err, fs.ErrInvalidName)
	require.NoError(t, h.Close())
}

func TestBeforeHookNonCancelError(t *testing.T) {
	f := fs.New()
	f.RegisterBeforeHook(func(fs.EventKind, string) error {
		return fmt.Errorf("blocked")
	})
	s := fs.NewSession(f)
	err := s.Mkdir("x")
	require.Error(t, err)
}

func TestSessionFS(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NotNil(t, s.FS())
}

func TestChmodReadWriteOff(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.Chmod("f", false, false))
}

func TestOtherUserDeniedRead(t *testing.T) {
	f := fs.New()
	f.AddUser("stranger", nil)
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("secret"))
	require.NoError(t, s.SwitchUser("stranger"))
	_, err := s.ReadFile("secret")
	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestMoveNotFound(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := s.Move("/a", "/b")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestCopyNotFound(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := s.Copy("/a", "/b")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestOpenMissingFile(t *testing.T) {
	s := fs.NewSession(fs.New())
	_, err := s.Open("missing", false)
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestFindExactFromNestedCwd(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/a/b"))
	require.NoError(t, s.Cd("/a/b"))
	require.NoError(t, s.CreateFile("needle"))
	require.NoError(t, s.Cd("/"))
	paths, err := s.Find("needle")
	require.NoError(t, err)
	require.Contains(t, paths, "/a/b/needle")
}

func TestRegisterBeforeHookCancelRemove(t *testing.T) {
	f := fs.New()
	f.RegisterBeforeHook(func(k fs.EventKind, _ string) error {
		if k == fs.EventRemove {
			return fs.ErrCanceled
		}
		return nil
	})
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("f"))
	err := s.Remove("f")
	require.ErrorIs(t, err, fs.ErrCanceled)
}

func TestAutoRenameCopy(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("a"))
	require.NoError(t, s.CreateFile("b"))
	require.NoError(t, s.WriteFile("a", "from-a"))
	require.NoError(t, s.Copy("a", "b"))
	got, err := s.ReadFile("b~1")
	require.NoError(t, err)
	require.Equal(t, "from-a", got)
}

func TestResolvePathDot(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Cd("."))
	pwd, err := s.Pwd()
	require.NoError(t, err)
	require.Equal(t, "/", pwd)
}

func TestBrokenSymlinkRead(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Symlink("/missing", "/link"))
	_, err := s.ReadFileAt("/link")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestWalkStopError(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	stop := errors.New("stop walk")
	err := s.Walk(func(string, *fs.WalkNode) error { return stop })
	require.ErrorIs(t, err, stop)
}

func TestCopyMergeIntoExistingDirectory(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/src/sub"))
	require.NoError(t, s.CreateFileAt("/src/sub/f"))
	require.NoError(t, s.MkdirAll("/dst"))
	require.NoError(t, s.Copy("/src", "/dst"))
	_, err := s.ReadFileAt("/dst/sub/f")
	require.NoError(t, err)
}

func TestMoveMergeIntoExistingDirectory(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/src/sub"))
	require.NoError(t, s.CreateFileAt("/src/sub/f"))
	require.NoError(t, s.MkdirAll("/dst"))
	require.NoError(t, s.Move("/src", "/dst"))
	_, err := s.ReadFileAt("/dst/sub/f")
	require.NoError(t, err)
	_, err = s.ReadFileAt("/src/sub/f")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestMoveFileAutoRenamesOnConflict(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/a"))
	require.NoError(t, s.MkdirAll("/b"))
	require.NoError(t, s.CreateFileAt("/a/x"))
	require.NoError(t, s.CreateFileAt("/b/y"))
	require.NoError(t, s.Move("/a/x", "/b/y"))
	data, err := s.ReadFileAt("/b/y~1")
	require.NoError(t, err)
	require.Empty(t, data)
}

func TestMoveFileSuccessInCWD(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("old"))
	require.NoError(t, s.MoveFile("old", "new"))
	_, err := s.ReadFile("new")
	require.NoError(t, err)
}

func TestCreateFileInvalidName(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := s.CreateFile("bad/name")
	require.ErrorIs(t, err, fs.ErrInvalidName)
}

func TestEntryInCWDInvalid(t *testing.T) {
	s := fs.NewSession(fs.New())
	_, err := s.ReadFile("bad/name")
	require.ErrorIs(t, err, fs.ErrInvalidName)
}


func TestWalkDirectoryKind(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/d/sub"))
	var sawDir bool
	require.NoError(t, s.Walk(func(_ string, n *fs.WalkNode) error {
		if n.Kind == fs.WalkDirectory {
			sawDir = true
		}
		return nil
	}))
	require.True(t, sawDir)
}

func TestReadEmptyFileHandle(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	h, err := s.Open("f", false)
	require.NoError(t, err)
	n, err := h.Read(make([]byte, 10))
	require.NoError(t, err)
	require.Equal(t, 0, n)
	require.NoError(t, h.Close())
}

func TestTransferTypeMismatchCopy(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	require.NoError(t, s.CreateFile("f"))
	err := s.Copy("f", "d")
	require.ErrorIs(t, err, fs.ErrTypeMismatch)
}

func TestRemoveAllNestedSymlink(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFileAt("/t"))
	require.NoError(t, s.MkdirAll("/d"))
	require.NoError(t, s.Symlink("/t", "/d/link"))
	require.NoError(t, s.RemoveAll("/d"))
}

func TestWriteFileAtPermissionOther(t *testing.T) {
	f := fs.New()
	f.AddUser("other", nil)
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.SwitchUser("other"))
	err := s.WriteFileAt("f", "nope")
	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestHardlinkKeepsInodeOnRemove(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFileAt("/f"))
	require.NoError(t, s.Hardlink("/f", "/f2"))
	require.NoError(t, s.Remove("/f"))
	_, err := s.ReadFileAt("/f2")
	require.NoError(t, err)
}

func TestRemoveUnlinkNotFound(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("gone"))
	require.NoError(t, s.Remove("gone"))
	err := s.Remove("gone")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestLinkMoveSameDirRename(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/p"))
	require.NoError(t, s.CreateFileAt("/p/a"))
	require.NoError(t, s.CreateFileAt("/p/b"))
	require.NoError(t, s.Move("/p/a", "/p/b"))
	_, err := s.ReadFileAt("/p/b~1")
	require.NoError(t, err)
}

func TestMoveFileErrors(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("a"))
	require.NoError(t, s.Mkdir("d"))
	err := s.MoveFile("missing", "x")
	require.ErrorIs(t, err, fs.ErrNotFound)
	err = s.MoveFile("d", "x")
	require.ErrorIs(t, err, fs.ErrIsDirectory)
	err = s.MoveFile("a", "a")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
}

func TestRmdirNotEmpty(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	require.NoError(t, s.Cd("d"))
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.Cd(".."))
	err := s.Rmdir("d")
	require.ErrorIs(t, err, fs.ErrDirectoryNotEmpty)
}

func TestWalkBrokenSymlink(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Symlink("/missing", "/bad"))
	err := s.Walk(func(string, *fs.WalkNode) error { return nil })
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestCopyDirectoryHookAbort(t *testing.T) {
	f := fs.New()
	f.RegisterBeforeHook(func(k fs.EventKind, p string) error {
		if k == fs.EventCreate && p == "/d/a/f2" {
			return fs.ErrCanceled
		}
		return nil
	})
	s := fs.NewSession(f)
	require.NoError(t, s.MkdirAll("/s/a"))
	require.NoError(t, s.CreateFileAt("/s/a/f1"))
	require.NoError(t, s.CreateFileAt("/s/a/f2"))
	err := s.Copy("/s", "/d")
	require.ErrorIs(t, err, fs.ErrCanceled)
}

func TestHardlinkAndSymlinkExist(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFileAt("/f"))
	require.NoError(t, s.Hardlink("/f", "/f2"))
	err := s.Hardlink("/f", "/f2")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
	require.NoError(t, s.Symlink("/f", "/s"))
	err = s.Symlink("/f", "/s")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
}

func TestGroupMemberReadOnly(t *testing.T) {
	f := fs.New()
	f.AddUser("bob", []string{"user"})
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("doc"))
	require.NoError(t, s.Chmod("doc", true, false))
	require.NoError(t, s.SwitchUser("bob"))
	err := s.WriteFile("doc", "nope")
	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestCreateFileDirectoryReadOnly(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	require.NoError(t, s.Cd("d"))
	require.NoError(t, s.Chmod(".", true, false))
	err := s.CreateFile("f")
	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestRmdirSymlinkToFile(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.Symlink("f", "link"))
	err := s.Rmdir("link")
	require.ErrorIs(t, err, fs.ErrNotDirectory)
}

func TestMoveFileRename(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("a"))
	require.NoError(t, s.MoveFile("a", "b"))
	_, err := s.ReadFile("b")
	require.NoError(t, err)
}

func TestHardlinkMissingSource(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := s.Hardlink("/nope", "/link")
	require.ErrorIs(t, err, fs.ErrNotFound)
}
