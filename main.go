package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/mgutz/ansi"
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
		b, err := regexp.MatchString("[\\@ ]", input)
		if err != nil {
			return err
		}
		dc.InHunk = b
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
	if len(dc.Added) == 0 || len(dc.Removed) == 0 {
		for _, v := range dc.Added {
			fmt.Println(v)
		}
		for _, v := range dc.Removed {
			fmt.Println(v)
		}
		return
	}

	if len(dc.Added) != len(dc.Removed) {
		for _, v := range dc.Added {
			fmt.Println(v)
		}
		for _, v := range dc.Removed {
			fmt.Println(v)
		}
		return
	}

	var queue []string
	for i := 0; i < len(dc.Added); i++ {
		a, r := highlighPair(dc.Added[i], dc.Removed[i])
		fmt.Println(r)
		queue = append(queue, a)
	}
	for _, v := range queue {
		fmt.Println(v)
	}
}

func highlighPair(added, removed string) (string, string) {

	addedRune, removedRune := []rune(added), []rune(removed)
	seenPlusMinus := false
	addedPrefixIndex, removedPrefixIndex := 0, 0
	for addedPrefixIndex < len(addedRune) && removedPrefixIndex < len(removedRune) {
		if addedRune[addedPrefixIndex] == removedRune[removedPrefixIndex] {
			addedPrefixIndex++
			removedPrefixIndex++
		} else if !seenPlusMinus && removedRune[0] == '-' && addedRune[0] == '+' {
			seenPlusMinus = true
			addedPrefixIndex++
			removedPrefixIndex++
		} else {
			break
		}
	}

	addedSuffixIndex, removedSuffixIndex := len(addedRune)-1, len(removedRune)-1
	for addedSuffixIndex > 0 && removedSuffixIndex > 0 {
		if addedRune[addedSuffixIndex] == removedRune[removedSuffixIndex] {
			addedSuffixIndex--
			removedSuffixIndex--
		} else {
			break
		}
	}

	return highlightLine(addedRune, addedPrefixIndex, addedSuffixIndex, "black:green"), highlightLine(removedRune, removedPrefixIndex, removedSuffixIndex, "black:red")
}

func highlightLine(line []rune, prefix, suffix int, theme string) string {
	t := ansi.ColorCode(theme)
	reset := ansi.ColorCode("reset")
	return fmt.Sprint(string(line[:prefix]), t, string(line[prefix:suffix+1]), reset, string(line[suffix+1:]))
}
