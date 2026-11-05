package fs_test

import (
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestStreamSeekAfterMove(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	require.NoError(t, s.WriteFile("f", "hello"))
	h, err := s.Open("f", false)
	require.NoError(t, err)
	require.NoError(t, s.Move("f", "g"))
	_, err = h.Seek(1, 0)
	require.NoError(t, err)
	buf := make([]byte, 4)
	n, err := h.Read(buf)
	require.NoError(t, err)
	require.Equal(t, 4, n)
	require.Equal(t, "ello", string(buf))
	require.NoError(t, h.Close())
}

func TestSingleWriter(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.CreateFile("f"))
	_, err := s.Open("f", true)
	require.NoError(t, err)
	_, err = s.Open("f", true)
	require.ErrorIs(t, err, fs.ErrWriterOpen)
}
