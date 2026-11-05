package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestSymlinkRead(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFileAt("/f"))
	require.NoError(t, s.WriteFileAt("/f", "x"))
	require.NoError(t, s.Symlink("/f", "/link"))
	got, err := s.ReadFileAt("/link")
	require.NoError(t, err)
	require.Equal(t, "x", got)
}

func TestHardlinkSurvivesMove(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("orig"))
	require.NoError(t, s.WriteFile("orig", "z"))
	require.NoError(t, s.Hardlink("orig", "alias"))
	require.NoError(t, s.Move("orig", "moved"))
	got, err := s.ReadFile("alias")
	require.NoError(t, err)
	require.Equal(t, "z", got)
	got, err = s.ReadFile("moved")
	require.NoError(t, err)
	require.Equal(t, "z", got)
}
