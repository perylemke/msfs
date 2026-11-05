package main

import (
	"fmt"
	"os"

	"github.com/perylemke/ms-test/internal/fs"
)

func main() {
	s := fs.NewSession(fs.New())
	pwd, err := s.Pwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(pwd)
}
