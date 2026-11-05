package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/perylemke/ms-test/internal/fs"
)

func runREPL(args []string) int {
	store := fs.New()
	session := fs.NewSession(store)
	if len(args) > 0 {
		return runScript(session, args[0])
	}
	sc := bufio.NewScanner(os.Stdin)
	fmt.Println("msfs interactive shell (type 'help')")
	for {
		fmt.Print("msfs> ")
		if !sc.Scan() {
			break
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if err := execLine(session, line); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}
	}
	return 0
}

func runScript(s *fs.Session, path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if err := execLine(s, line); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			return 1
		}
	}
	return 0
}

func execLine(s *fs.Session, line string) error {
	fields := splitFields(line)
	if len(fields) == 0 {
		return nil
	}
	cmd := fields[0]
	if handled, err := handleHelpRequest(cmd, fields); handled {
		return err
	}
	switch cmd {
	case "pwd":
		p, err := s.Pwd()
		if err != nil {
			return err
		}
		fmt.Println(p)
	case "cd":
		if len(fields) < 2 {
			return fmt.Errorf("usage: cd <path>")
		}
		return s.Cd(fields[1])
	case "mkdir":
		if len(fields) < 2 {
			return fmt.Errorf("usage: mkdir <name>...")
		}
		var firstErr error
		for _, name := range fields[1:] {
			if err := s.Mkdir(name); err != nil {
				if firstErr == nil {
					firstErr = err
				}
			}
		}
		return firstErr
	case "mkdir-p", "mkdirall":
		if len(fields) < 2 {
			return fmt.Errorf("usage: mkdir-p <path>")
		}
		return s.MkdirAll(fields[1])
	case "ls":
		names, err := s.List()
		if err != nil {
			return err
		}
		fmt.Println(strings.Join(names, " "))
	case "rmdir":
		if len(fields) < 2 {
			return fmt.Errorf("usage: rmdir <name>")
		}
		return s.Rmdir(fields[1])
	case "rm":
		if len(fields) < 2 {
			return fmt.Errorf("usage: rm [-r] <path>...")
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
			return fmt.Errorf("usage: rm [-r] <path>...")
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
	case "touch":
		if len(fields) < 2 {
			return fmt.Errorf("usage: touch <path>")
		}
		return s.CreateFileAt(fields[1])
	case "write":
		if len(fields) < 3 {
			return fmt.Errorf("usage: write <path> <content>")
		}
		return s.WriteFileAt(fields[1], fields[2])
	case "cat":
		if len(fields) < 2 {
			return fmt.Errorf("usage: cat <path>")
		}
		out, err := s.ReadFileAt(fields[1])
		if err != nil {
			return err
		}
		fmt.Println(out)
	case "mv":
		if len(fields) < 3 {
			return fmt.Errorf("usage: mv <src> <dst>")
		}
		return s.Move(fields[1], fields[2])
	case "cp":
		if len(fields) < 3 {
			return fmt.Errorf("usage: cp <src> <dst>")
		}
		return s.Copy(fields[1], fields[2])
	case "find":
		if len(fields) < 2 {
			return fmt.Errorf("usage: find <name>")
		}
		paths, err := s.Find(fields[1])
		if err != nil {
			return err
		}
		for _, p := range paths {
			fmt.Println(p)
		}
	case "findre":
		if len(fields) < 2 {
			return fmt.Errorf("usage: findre <regex>")
		}
		p, err := s.FindFirstRegex(fields[1])
		if err != nil {
			return err
		}
		fmt.Println(p)
	case "ln":
		if len(fields) < 3 {
			return fmt.Errorf("usage: ln <target> <link>")
		}
		return s.Hardlink(fields[1], fields[2])
	case "ln-s":
		if len(fields) < 3 {
			return fmt.Errorf("usage: ln-s <target> <link>")
		}
		return s.Symlink(fields[1], fields[2])
	case "chmod":
		if len(fields) < 3 {
			return fmt.Errorf("usage: chmod <path> <rw|ro>")
		}
		read := true
		write := fields[2] == "rw"
		return s.Chmod(fields[1], read, write)
	case "su":
		if len(fields) < 2 {
			return fmt.Errorf("usage: su <user>")
		}
		return s.SwitchUser(fields[1])
	case "user":
		if len(fields) < 2 {
			return fmt.Errorf("usage: user add <name>")
		}
		if fields[1] == "add" && len(fields) >= 3 {
			s.FS().AddUser(fields[2], nil)
			return nil
		}
		return fmt.Errorf("usage: user add <name>")
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
	return nil
}

func splitFields(line string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	for _, r := range line {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ' ' && !inQuote:
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
