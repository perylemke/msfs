package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestMoveCopyMergeAndTilde(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/src/a"))
	require.NoError(t, s.CreateFileAt("/src/a/one"))
	require.NoError(t, s.WriteFileAt("/src/a/one", "1"))
	require.NoError(t, s.MkdirAll("/dst/a"))
	require.NoError(t, s.CreateFileAt("/dst/a/two"))
	require.NoError(t, s.Move("/src/a", "/dst/a"))
	_, err := s.ReadFileAt("/dst/a/one")
	require.NoError(t, err)
	_, err = s.ReadFileAt("/dst/a/two")
	require.NoError(t, err)

	s2 := fs.NewSession(fs.New())
	require.NoError(t, s2.CreateFile("a"))
	require.NoError(t, s2.CreateFile("b"))
	require.NoError(t, s2.Copy("a", "b"))
	_, err = s2.ReadFile("a")
	require.NoError(t, err)
	_, err = s2.ReadFile("b~1")
	require.NoError(t, err)
}

func TestMoveTypeMismatch(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	require.NoError(t, s.CreateFile("f"))
	err := s.Move("f", "d")
	require.ErrorIs(t, err, fs.ErrTypeMismatch)
}
