package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestMkdirHelpDoesNotCreateDirectory(t *testing.T) {
	s := fs.NewSession(fs.New())
	for _, line := range []string{"mkdir --help", "mkdir -h", "help mkdir"} {
		require.NoError(t, execLine(s, line))
	}
	list, err := s.List()
	require.NoError(t, err)
	require.Empty(t, list)
}

func TestUnknownCommandHelpStillErrors(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := execLine(s, "nope --help")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown command")
}

func TestPrintCommandHelpMkdir(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	require.True(t, printCommandHelp("mkdir"))
	require.NoError(t, w.Close())
	os.Stdout = old
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	out := buf.String()
	require.Contains(t, out, "usage: mkdir")
	require.Contains(t, out, "must not contain")
}

func TestMkdirMultipleDirectories(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, execLine(s, "mkdir pasta1 pasta2 pasta3"))
	list, err := s.List()
	require.NoError(t, err)
	require.Equal(t, []string{"pasta1", "pasta2", "pasta3"}, list)
}

func TestMkdirMultipleContinuesAfterError(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.NoError(t, s.Mkdir("pasta2"))
	err := execLine(s, "mkdir pasta1 pasta2 pasta3")
	require.ErrorIs(t, err, fs.ErrAlreadyExists)
	list, listErr := s.List()
	require.NoError(t, listErr)
	require.Equal(t, []string{"pasta2", "pasta1", "pasta3"}, list)
}

func TestMainHelpFlag(t *testing.T) {
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	printHelp()
	require.NoError(t, w.Close())
	os.Stdout = old
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.True(t, strings.Contains(buf.String(), "mkdir-p"))
}
