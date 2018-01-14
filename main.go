package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"
)

var (
	ColorPattern   = regexp.MustCompile("^(\\x1b\\[[0-9;]*m)+")
	HunkPattern    = regexp.MustCompile("^\\@\\@")
	AddPattern     = regexp.MustCompile("^\\+")
	RemovedPattern = regexp.MustCompile("^\\-")
)

type DiffContext struct {
	InHunk  bool
	Added   []string
	Removed []string
}

func main() {
	signal.Ignore(syscall.SIGPIPE)
	scanner := bufio.NewScanner(os.Stdin)
	dc := &DiffContext{InHunk: false, Added: []string{}, Removed: []string{}}
	for scanner.Scan() {
		t := scanner.Text()
		err := dc.handleLine(t) // Println will add back the final '\n'
		if err != nil {
			log.Fatal(err)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		return
	}
	dc.ShowHunk()
}

func (dc *DiffContext) handleLine(input string) error {
	if !dc.InHunk {
		fmt.Println(input)
		dc.InHunk = HunkPattern.MatchString(input)
	} else if isRemovedLine(input) {
		dc.Removed = append(dc.Removed, input)
	} else if isAddedLine(input) {
		dc.Added = append(dc.Added, input)
	} else {
		dc.ShowHunk()
		dc.Added = []string{}
		dc.Removed = []string{}
	}
	return nil
}

func isInHunk(in string) bool {
	return false
}

func isRemovedLine(in string) bool {
	return RemovedPattern.MatchString(in)
}

func isAddedLine(in string) bool {
	return AddPattern.MatchString(in)
}

func (dc *DiffContext) ShowHunk() {
	fmt.Println("Added lines")
	for _, v := range dc.Added {
		fmt.Println("\t" + v)
	}

	fmt.Println("Remvoed lines")
	for _, v := range dc.Removed {
		fmt.Println("\t" + v)
	}
}
