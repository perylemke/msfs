package main

import (
	"fmt"
	"strings"

	"github.com/perylemke/ms-test/internal/fs"
)

var commands = map[string]func(*fs.Session, []string) error{
	"pwd":      cmdPwd,
	"cd":       cmdCd,
	"mkdir":    cmdMkdir,
	"mkdir-p":  cmdMkdirP,
	"mkdirall": cmdMkdirP,
	"ls":       cmdLs,
	"rmdir":    cmdRmdir,
	"rm":       cmdRm,
	"touch":    cmdTouch,
	"write":    cmdWrite,
	"cat":      cmdCat,
	"mv":       cmdMv,
	"cp":       cmdCp,
	"find":     cmdFind,
	"findre":   cmdFindRe,
	"ln":       cmdLn,
	"ln-s":     cmdLnS,
	"chmod":    cmdChmod,
	"su":       cmdSu,
	"user":     cmdUser,
}

func usage(fields []string, n int, msg string) error {
	if len(fields) < n {
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func cmdPwd(s *fs.Session, _ []string) error {
	p, err := s.Pwd()
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}

func cmdCd(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: cd <path>"); err != nil {
		return err
	}
	return s.Cd(fields[1])
}

func cmdMkdir(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: mkdir <name> [name ...]"); err != nil {
		return err
	}
	var firstErr error
	for _, name := range fields[1:] {
		if err := s.Mkdir(name); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func cmdMkdirP(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: mkdir-p <path>"); err != nil {
		return err
	}
	return s.MkdirAll(fields[1])
}

func cmdLs(s *fs.Session, _ []string) error {
	names, err := s.List()
	if err != nil {
		return err
	}
	fmt.Println(strings.Join(names, " "))
	return nil
}

func cmdRmdir(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: rmdir <name>"); err != nil {
		return err
	}
	return s.Rmdir(fields[1])
}

func cmdRm(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: rm [-r] <path> [path ...]"); err != nil {
		return err
	}
	recursive := false
	var paths []string
	for _, arg := range fields[1:] {
		if arg == "-r" {
			recursive = true
			continue
		}
		paths = append(paths, arg)
	}
	if len(paths) == 0 {
		return fmt.Errorf("usage: rm [-r] <path> [path ...]")
	}
	var firstErr error
	for _, p := range paths {
		var err error
		if recursive {
			err = s.RemoveAll(p)
		} else {
			err = s.Remove(p)
		}
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func cmdTouch(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: touch <path>"); err != nil {
		return err
	}
	return s.CreateFileAt(fields[1])
}

func cmdWrite(s *fs.Session, fields []string) error {
	if err := usage(fields, 3, "usage: write <path> <content>"); err != nil {
		return err
	}
	return s.WriteFileAt(fields[1], fields[2])
}

func cmdCat(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: cat <path>"); err != nil {
		return err
	}
	out, err := s.ReadFileAt(fields[1])
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}

func cmdMv(s *fs.Session, fields []string) error {
	if err := usage(fields, 3, "usage: mv <src> <dst>"); err != nil {
		return err
	}
	return s.Move(fields[1], fields[2])
}

func cmdCp(s *fs.Session, fields []string) error {
	if err := usage(fields, 3, "usage: cp <src> <dst>"); err != nil {
		return err
	}
	return s.Copy(fields[1], fields[2])
}

func cmdFind(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: find <name>"); err != nil {
		return err
	}
	paths, err := s.Find(fields[1])
	if err != nil {
		return err
	}
	for _, p := range paths {
		fmt.Println(p)
	}
	return nil
}

func cmdFindRe(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: findre <regex>"); err != nil {
		return err
	}
	p, err := s.FindFirstRegex(fields[1])
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}

func cmdLn(s *fs.Session, fields []string) error {
	if err := usage(fields, 3, "usage: ln <target> <link>"); err != nil {
		return err
	}
	return s.Hardlink(fields[1], fields[2])
}

func cmdLnS(s *fs.Session, fields []string) error {
	if err := usage(fields, 3, "usage: ln-s <target> <link>"); err != nil {
		return err
	}
	return s.Symlink(fields[1], fields[2])
}

func cmdChmod(s *fs.Session, fields []string) error {
	if err := usage(fields, 3, "usage: chmod <path> <rw|ro>"); err != nil {
		return err
	}
	return s.Chmod(fields[1], true, fields[2] == "rw")
}

func cmdSu(s *fs.Session, fields []string) error {
	if err := usage(fields, 2, "usage: su <user>"); err != nil {
		return err
	}
	return s.SwitchUser(fields[1])
}

func cmdUser(s *fs.Session, fields []string) error {
	if len(fields) >= 3 && fields[1] == "add" {
		s.FS().AddUser(fields[2], nil)
		return nil
	}
	return fmt.Errorf("usage: user add <name>")
}
