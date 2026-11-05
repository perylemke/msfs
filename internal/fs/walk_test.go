package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestWalkSkipDir(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/a/skip/b"))
	require.NoError(t, s.MkdirAll("/a/keep/c"))
	var seen []string
	err := s.Walk(func(p string, n *fs.WalkNode) error {
		if n.Kind == fs.WalkDirectory && p == "/a/skip" {
			return fs.SkipDir
		}
		seen = append(seen, p)
		return nil
	})
	require.NoError(t, err)
	require.Contains(t, seen, "/a/keep/c")
	require.NotContains(t, seen, "/a/skip/b")
}

func TestFindFirstRegex(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.MkdirAll("/d"))
	require.NoError(t, s.CreateFileAt("/d/foo-1.txt"))
	require.NoError(t, s.CreateFileAt("/d/bar.txt"))
	got, err := s.FindFirstRegex(`foo-\d+\.txt`)
	require.NoError(t, err)
	require.Equal(t, "/d/foo-1.txt", got)
}
