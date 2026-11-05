package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestAfterHookRuns(t *testing.T) {
	var count int
	f := fs.New()
	f.RegisterAfterHook(func(ev fs.Event) { count++ })
	s := fs.NewSession(f)
	require.NoError(t, s.Mkdir("a"))
	require.GreaterOrEqual(t, count, 1)
}

func TestFindFirstRegexNotFound(t *testing.T) {
	s := fs.NewSession(fs.New())
	_, err := s.FindFirstRegex(`^none$`)
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestFindFirstRegexInvalid(t *testing.T) {
	s := fs.NewSession(fs.New())
	_, err := s.FindFirstRegex(`[`)
	require.Error(t, err)
}

func TestChmodReadOnlyStillReadable(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.WriteFile("f", "ok"))
	require.NoError(t, s.Chmod("f", true, false))
	got, err := s.ReadFile("f")
	require.NoError(t, err)
	require.Equal(t, "ok", got)
}

func TestNewSessionFrom(t *testing.T) {
	f := fs.New()
	a := fs.NewSession(f)
	b := fs.NewSessionFrom(f)
	require.NoError(t, a.Mkdir("d"))
	require.NoError(t, b.Cd("d"))
	pwd, _ := b.Pwd()
	require.Equal(t, "/d", pwd)
}

func TestCopyDirectoryTree(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/s/n"))
	require.NoError(t, s.CreateFileAt("/s/n/f"))
	require.NoError(t, s.Copy("/s", "/d"))
	_, err := s.ReadFileAt("/d/n/f")
	require.NoError(t, err)
}

func TestSwitchUserMissing(t *testing.T) {
	err := fs.NewSession(fs.New()).SwitchUser("missing")
	require.ErrorIs(t, err, fs.ErrNotFound)
}

func TestSymlinkLoop(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Symlink("/loop", "/loop"))
	_, err := s.ReadFileAt("/loop")
	require.ErrorIs(t, err, fs.ErrSymlinkLoop)
}

func TestHandleSeekFromEnd(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.WriteFile("f", "abc"))
	h, err := s.Open("f", false)
	require.NoError(t, err)
	_, err = h.Seek(-1, 2)
	require.NoError(t, err)
	buf := make([]byte, 1)
	_, err = h.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "c", string(buf))
	require.NoError(t, h.Close())
}

func TestChmodPermissionDenied(t *testing.T) {
	f := fs.New()
	f.AddUser("other", nil)
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("x"))
	require.NoError(t, s.SwitchUser("other"))
	err := s.Chmod("x", true, false)
	require.ErrorIs(t, err, fs.ErrPermission)
}

func TestBeforeHookCancelWrite(t *testing.T) {
	f := fs.New()
	f.RegisterBeforeHook(func(kind fs.EventKind, path string) error {
		if kind == fs.EventWrite {
			return fs.ErrCanceled
		}
		return nil
	})
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("f"))
	err := s.WriteFile("f", "blocked")
	require.ErrorIs(t, err, fs.ErrCanceled)
}

func TestOpenDirectoryFails(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	_, err := s.Open("d", false)
	require.ErrorIs(t, err, fs.ErrIsDirectory)
}

func TestHandleWriteRequiresWriter(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	h, err := s.Open("f", false)
	require.NoError(t, err)
	_, err = h.Write([]byte("x"))
	require.ErrorIs(t, err, fs.ErrPermission)
	require.NoError(t, h.Close())
}

func TestMovePath(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/a"))
	require.NoError(t, s.MkdirAll("/b"))
	require.NoError(t, s.CreateFileAt("/a/f"))
	require.NoError(t, s.Move("/a/f", "/b/f"))
	_, err := s.ReadFileAt("/b/f")
	require.NoError(t, err)
}
