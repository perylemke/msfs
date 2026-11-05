package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestRemoveFileHardlinkKeepsAlias(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("orig"))
	require.NoError(t, s.WriteFile("orig", "data"))
	require.NoError(t, s.Hardlink("orig", "alias"))
	require.NoError(t, s.Remove("orig"))
	got, err := s.ReadFile("alias")
	require.NoError(t, err)
	require.Equal(t, "data", got)
}

func TestRemoveDirectoryRequiresRemoveAll(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("d"))
	err := s.Remove("d")
	require.ErrorIs(t, err, fs.ErrIsDirectory)
	require.NoError(t, s.RemoveAll("d"))
	list, err := s.List()
	require.NoError(t, err)
	require.Empty(t, list)
}
