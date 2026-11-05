package main

import (
	"fmt"
	"io"
	"os"
	"sort"
)

//nolint:gosec // G101: "pwd" is a command name, not a credential
var commandHelp = map[string]string{
	"pwd": `usage: pwd
Print the absolute path of the current working directory.`,
	"cd": `usage: cd <path>
Change the working directory. Supports relative paths, .., and absolute paths.`,
	"mkdir": `usage: mkdir <name>...
Create one or more directories in the current working directory. A name must not contain '/'.`,
	"mkdir-p": `usage: mkdir-p <path>
Create a directory and any missing parent directories. Path may be absolute or relative.`,
	"mkdirall": `usage: mkdir-p <path>
Create a directory and any missing parent directories. Path may be absolute or relative.`,
	"ls": `usage: ls
List names of children in the current working directory (creation order).`,
	"rmdir": `usage: rmdir <name>
Remove an empty directory that is a direct child of the current working directory.`,
	"rm": `usage: rm [-r] <path>...
Remove files. With -r, also remove directories and their contents.`,
	"touch": `usage: touch <path>
Create an empty file at the given path (absolute or relative).`,
	"write": `usage: write <path> <content>
Write or replace the contents of a file at the given path.`,
	"cat": `usage: cat <path>
Print the contents of a file at the given path.`,
	"mv": `usage: mv <src> <dst>
Move or rename a file or directory. Supports merge when both paths are directories.`,
	"cp": `usage: cp <src> <dst>
Copy a file or directory. File name collisions get a ~N suffix.`,
	"find": `usage: find <name>
List absolute paths under the working directory whose base name matches exactly.`,
	"findre": `usage: findre <regex>
Print the first file path under the working directory whose base name matches the regex.`,
	"ln": `usage: ln <target> <link>
Create a hard link: another name for the same file or directory inode.`,
	"ln-s": `usage: ln-s <target> <link>
Create a symbolic link at link pointing to target.`,
	"chmod": `usage: chmod <path> <rw|ro>
Set read/write mode on a path. rw allows read and write; ro is read-only. Owner only.`,
	"su": `usage: su <user>
Switch the session user (no permission check; for testing).`,
	"user": `usage: user add <name>
Register a new user on the shared filesystem.`,
	"help": `usage: help [command]
List commands, or show help for command. Same as: command --help`,
}

func isHelpFlag(s string) bool {
	return s == "--help" || s == "-h"
}

func argsHaveHelpFlag(args []string) bool {
	for _, a := range args {
		if isHelpFlag(a) {
			return true
		}
	}
	return false
}

func printCommandHelp(name string) (bool, error) {
	text, ok := commandHelp[name]
	if !ok {
		return false, nil
	}
	_, err := io.WriteString(os.Stdout, text+"\n")
	return true, err
}

func printHelp() error {
	names := make([]string, 0, len(commandHelp))
	seen := make(map[string]struct{})
	for name := range commandHelp {
		if name == "mkdirall" {
			continue
		}
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	sort.Strings(names)
	if _, err := io.WriteString(os.Stdout, "msfs commands (use <command> --help for details):\n"); err != nil {
		return err
	}
	for _, name := range names {
		if _, err := fmt.Fprintf(os.Stdout, "  %s\n", name); err != nil {
			return err
		}
	}
	_, err := io.WriteString(os.Stdout, "\nAliases: mkdirall (same as mkdir-p)\n")
	return err
}

func handleHelpRequest(cmd string, fields []string) (handled bool, err error) {
	if cmd == "help" {
		if len(fields) >= 2 && !isHelpFlag(fields[1]) {
			ok, err := printCommandHelp(fields[1])
			if err != nil {
				return true, err
			}
			if !ok {
				return true, fmt.Errorf("unknown command: %s", fields[1])
			}
			return true, nil
		}
		return true, printHelp()
	}
	if len(fields) > 1 && argsHaveHelpFlag(fields[1:]) {
		ok, err := printCommandHelp(cmd)
		if err != nil {
			return true, err
		}
		if !ok {
			return true, fmt.Errorf("unknown command: %s", cmd)
		}
		return true, nil
	}
	return false, nil
}
