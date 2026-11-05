package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestPDFNavigationScript(t *testing.T) {
	t.Parallel()
	s := fs.NewSession(fs.New())

	require.NoError(t, s.Mkdir("school"))
	require.NoError(t, s.Cd("school"))
	pwd, err := s.Pwd()
	require.NoError(t, err)
	require.Equal(t, "/school", pwd)

	require.NoError(t, s.Mkdir("homework"))
	require.NoError(t, s.Cd("homework"))
	require.NoError(t, s.Mkdir("math"))
	require.NoError(t, s.Mkdir("lunch"))
	require.NoError(t, s.Mkdir("history"))
	require.NoError(t, s.Mkdir("spanish"))
	require.NoError(t, s.Rmdir("lunch"))

	list, err := s.List()
	require.NoError(t, err)
	require.Equal(t, []string{"math", "history", "spanish"}, list)

	pwd, err = s.Pwd()
	require.NoError(t, err)
	require.Equal(t, "/school/homework", pwd)

	require.NoError(t, s.Cd(".."))
	require.NoError(t, s.Mkdir("cheatsheet"))
	list, err = s.List()
	require.NoError(t, err)
	require.Equal(t, []string{"homework", "cheatsheet"}, list)

	require.NoError(t, s.Rmdir("cheatsheet"))
	require.NoError(t, s.Cd(".."))
	pwd, err = s.Pwd()
	require.NoError(t, err)
	require.Equal(t, "/", pwd)
}

func TestNavigationErrors(t *testing.T) {
	t.Parallel()

	t.Run("cd missing", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		err := s.Cd("nope")
		require.ErrorIs(t, err, fs.ErrNotFound)
	})

	t.Run("mkdir duplicate", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.Mkdir("a"))
		err := s.Mkdir("a")
		require.ErrorIs(t, err, fs.ErrAlreadyExists)
	})

	t.Run("invalid name", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		err := s.Mkdir("bad/name")
		require.ErrorIs(t, err, fs.ErrInvalidName)
	})

	t.Run("rmdir not empty", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.Mkdir("d"))
		require.NoError(t, s.Cd("d"))
		require.NoError(t, s.Mkdir("child"))
		require.NoError(t, s.Cd(".."))
		err := s.Rmdir("d")
		require.ErrorIs(t, err, fs.ErrDirectoryNotEmpty)
	})

	t.Run("rmdir not child of cwd", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.Mkdir("a"))
		require.NoError(t, s.Mkdir("b"))
		require.NoError(t, s.Cd("a"))
		err := s.Rmdir("b")
		require.ErrorIs(t, err, fs.ErrNotChild)
	})

	t.Run("cd parent at root", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.Cd(".."))
		pwd, err := s.Pwd()
		require.NoError(t, err)
		require.Equal(t, "/", pwd)
	})
}
