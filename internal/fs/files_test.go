package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestFileOperations(t *testing.T) {
	t.Parallel()
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("notes.txt"))
	require.NoError(t, s.WriteFile("notes.txt", "hello"))
	got, err := s.ReadFile("notes.txt")
	require.NoError(t, err)
	require.Equal(t, "hello", got)
	require.NoError(t, s.MoveFile("notes.txt", "notes2.txt"))
	_, err = s.ReadFile("notes.txt")
	require.ErrorIs(t, err, fs.ErrNotFound)
	got, err = s.ReadFile("notes2.txt")
	require.NoError(t, err)
	require.Equal(t, "hello", got)
}

func TestFindExactName(t *testing.T) {
	t.Parallel()
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("a"))
	require.NoError(t, s.Cd("a"))
	require.NoError(t, s.CreateFile("target"))
	require.NoError(t, s.Mkdir("nested"))
	require.NoError(t, s.Cd("nested"))
	require.NoError(t, s.CreateFile("target"))
	require.NoError(t, s.Cd("/"))
	paths, err := s.Find("target")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"/a/target", "/a/nested/target"}, paths)
}

func TestFileErrors(t *testing.T) {
	t.Parallel()

	t.Run("read missing", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		_, err := s.ReadFile("x")
		require.ErrorIs(t, err, fs.ErrNotFound)
	})

	t.Run("read directory", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.Mkdir("d"))
		_, err := s.ReadFile("d")
		require.ErrorIs(t, err, fs.ErrIsDirectory)
	})

	t.Run("write directory", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.Mkdir("d"))
		err := s.WriteFile("d", "nope")
		require.ErrorIs(t, err, fs.ErrIsDirectory)
	})

	t.Run("move dest exists", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.CreateFile("a"))
		require.NoError(t, s.CreateFile("b"))
		err := s.MoveFile("a", "b")
		require.ErrorIs(t, err, fs.ErrAlreadyExists)
	})

	t.Run("rmdir on file", func(t *testing.T) {
		t.Parallel()
		s := fs.NewSession(fs.New())
		require.NoError(t, s.CreateFile("f"))
		err := s.Rmdir("f")
		require.ErrorIs(t, err, fs.ErrNotDirectory)
	})
}
