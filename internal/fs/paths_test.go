package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestPathsMkdirAllAndCd(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/a/b/c"))
	require.NoError(t, s.Cd("/a/b"))
	pwd, err := s.Pwd()
	require.NoError(t, err)
	require.Equal(t, "/a/b", pwd)
	require.NoError(t, s.Cd(".."))
	pwd, err = s.Pwd()
	require.NoError(t, err)
	require.Equal(t, "/a", pwd)
}

func TestPathsReadWriteAt(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/x/y"))
	require.NoError(t, s.CreateFileAt("/x/y/f.txt"))
	require.NoError(t, s.WriteFileAt("/x/y/f.txt", "data"))
	got, err := s.ReadFileAt("/x/y/f.txt")
	require.NoError(t, err)
	require.Equal(t, "data", got)
}

func TestPathThroughFile(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	err := s.Cd("f")
	require.ErrorIs(t, err, fs.ErrNotDirectory)
}
