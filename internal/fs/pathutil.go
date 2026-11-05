package fs

import (
	"path"
	"strings"
)

func validateName(name string) error {
	if name == "" || name == "." || name == ".." || strings.Contains(name, "/") {
		return ErrInvalidName
	}
	return nil
}

func absPath(cwd, p string) string {
	if p == "" || p == "." {
		return cwd
	}
	if strings.HasPrefix(p, "/") {
		return path.Clean(p)
	}
	return path.Clean(path.Join(cwd, p))
}

func splitAbs(abs string) []string {
	if abs == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(abs, "/"), "/")
}

type resolved struct {
	inode *inode
	abs   string
}
