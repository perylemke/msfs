package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestPermissionsDenyWrite(t *testing.T) {
	f := fs.New()
	f.AddUser("other", nil)
	s := fs.NewSession(f)
	require.NoError(t, s.CreateFile("secret"))
	require.NoError(t, s.Chmod("secret", true, false))
	require.NoError(t, s.SwitchUser("other"))
	err := s.WriteFile("secret", "no")
	require.ErrorIs(t, err, fs.ErrPermission)
}
