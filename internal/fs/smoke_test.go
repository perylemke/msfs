package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestNewSessionPwdRoot(t *testing.T) {
	t.Parallel()
	s := fs.NewSession(fs.New())
	got, err := s.Pwd()
	require.NoError(t, err)
	require.Equal(t, "/", got)
}
