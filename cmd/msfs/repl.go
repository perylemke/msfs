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
	if err := sc.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}
	return 0
}

func runScript(s *fs.Session, path string) int {
	//nolint:gosec // G304: the script path is the CLI argument the user asked to run
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
	fn, ok := commands[cmd]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd)
	}
	return fn(s, fields)
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
