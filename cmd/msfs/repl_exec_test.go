package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/perylemke/ms-test/internal/fs"
	"github.com/stretchr/testify/require"
)

func TestExecLineCommands(t *testing.T) {
	s := fs.NewSession(fs.New())

	require.NoError(t, execLine(s, "pwd"))
	require.NoError(t, execLine(s, "mkdir school"))
	require.NoError(t, execLine(s, "cd school"))
	require.NoError(t, execLine(s, "mkdir-p /school/deep/n"))
	require.NoError(t, execLine(s, "touch /school/f.txt"))
	require.NoError(t, execLine(s, "write /school/f.txt hi"))
	require.NoError(t, execLine(s, "cat /school/f.txt"))
	require.NoError(t, execLine(s, "ls"))
	require.NoError(t, execLine(s, "find f.txt"))
	require.NoError(t, execLine(s, "findre f\\.txt"))
	require.NoError(t, execLine(s, "cp /school/f.txt /school/f2.txt"))
	require.NoError(t, execLine(s, "mv /school/f2.txt /school/moved.txt"))
	require.NoError(t, execLine(s, "ln /school/moved.txt /school/hard"))
	require.NoError(t, execLine(s, "ln-s /school/moved.txt /school/soft"))
	require.NoError(t, execLine(s, "chmod /school/moved.txt rw"))
	require.NoError(t, execLine(s, "user add guest"))
	require.NoError(t, execLine(s, "su user"))
	require.NoError(t, execLine(s, "rm /school/hard"))
	require.NoError(t, execLine(s, "rm -r deep"))
}

func TestRunScriptFile(t *testing.T) {
	dir := t.TempDir()
	script := filepath.Join(dir, "script.txt")
	require.NoError(t, os.WriteFile(script, []byte("pwd\nmkdir x\nls\n"), 0o600))
	s := fs.NewSession(fs.New())
	require.Equal(t, 0, runScript(s, script))
	list, err := s.List()
	require.NoError(t, err)
	require.Contains(t, list, "x")
}

func TestHelpUnknownCommand(t *testing.T) {
	s := fs.NewSession(fs.New())
	err := execLine(s, "help unknown-cmd")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown command")
}

func TestExecLineUsageErrors(t *testing.T) {
	s := fs.NewSession(fs.New())
	require.Error(t, execLine(s, "cd"))
	require.Error(t, execLine(s, "write a"))
	require.Error(t, execLine(s, "mv a"))
	require.Error(t, execLine(s, "user"))
	require.Error(t, execLine(s, "unknown"))
}
