package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestBeforeHookCancel(t *testing.T) {
	f := fs.New()
	f.RegisterBeforeHook(func(kind fs.EventKind, path string) error {
		if kind == fs.EventRemove {
			return fs.ErrCanceled
		}
		return nil
	})
	s := fs.NewSession(f)
	require.NoError(t, s.Mkdir("d"))
	err := s.Rmdir("d")
	require.ErrorIs(t, err, fs.ErrCanceled)
}

func TestSharedFilesystemSessions(t *testing.T) {
	f := fs.New()
	a := fs.NewSession(f)
	b := fs.NewSessionFrom(f)
	require.NoError(t, a.CreateFile("shared"))
	require.NoError(t, b.Cd("/"))
	got, err := b.ReadFile("shared")
	require.NoError(t, err)
	require.Equal(t, "", got)
	pwd, _ := b.Pwd()
	require.Equal(t, "/", pwd)
}
