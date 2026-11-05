package main

import (
	"os"

	"github.com/perylemke/ms-test/internal/fs"
)

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
	}
	if len(os.Args) >= 3 && os.Args[1] == "run" {
		store := fs.New()
		session := fs.NewSession(store)
		os.Exit(runScript(session, os.Args[2]))
	}
	os.Exit(runREPL(nil))
}
