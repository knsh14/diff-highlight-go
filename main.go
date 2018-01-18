package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/mgutz/ansi"
)

var (
	// ColorPattern defines color pattern
	ColorPattern = regexp.MustCompile(`^\\x1b\[[0-9;]*m`)
	// HunkPattern defines hunk
	HunkPattern = regexp.MustCompile(`^(\\x1b\[[0-9;]*m)*\@\@`)

	// AddPattern defines add line pattern
	AddPattern = regexp.MustCompile(`^(\\x1b\[[0-9;]*m)*\\+`)

	// RemovedPattern defines removed line pattern
	RemovedPattern = regexp.MustCompile(`(\\x1b\[[0-9;]*m)*^\\-`)
)

// DiffContext contains infomation about diff
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
		ul := UnescapedLine(t)
		err := dc.handleLine(ul)
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

// UnescapedLine retuns line with color info
func UnescapedLine(s string) string {
	quoted := strconv.Quote(s)
	return strings.Trim(quoted, "\"")
}

func (dc *DiffContext) handleLine(input string) error {
	if !dc.InHunk {
		printQuotedLine(input)
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

// ShowHunk shows lien in dc
func (dc *DiffContext) ShowHunk() {
	if len(dc.Added) == 0 || len(dc.Removed) == 0 {
		for _, v := range dc.Added {
			printQuotedLine(v)
		}
		for _, v := range dc.Removed {
			printQuotedLine(v)
		}
		return
	}

	if len(dc.Added) != len(dc.Removed) {
		for _, v := range dc.Added {
			printQuotedLine(v)
		}
		for _, v := range dc.Removed {
			printQuotedLine(v)
		}
		return
	}

	var queue []string
	for i := 0; i < len(dc.Added); i++ {
		a, r := highlighPair(dc.Added[i], dc.Removed[i])
		printQuotedLine(r)
		queue = append(queue, a)
	}
	for _, v := range queue {
		printQuotedLine(v)
	}
}

func printQuotedLine(s string) {
	v, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		panic(err)
	}
	fmt.Println(v)
}

func highlighPair(added, removed string) (string, string) {

	addedRune, removedRune := []rune(added), []rune(removed)
	seenPlusMinus := false
	addedPrefixIndex, removedPrefixIndex := 0, 0
	for addedPrefixIndex < len(added) && removedPrefixIndex < len(removed) {
		if loc := ColorPattern.FindStringIndex(added[addedPrefixIndex:]); loc != nil {
			addedPrefixIndex += loc[1] - 1
		} else if loc := ColorPattern.FindStringIndex(removed[removedPrefixIndex:]); loc != nil {
			removedPrefixIndex += loc[1] - 1
		} else if added[addedPrefixIndex] == removed[removedPrefixIndex] {
			addedPrefixIndex++
			removedPrefixIndex++
		} else if !seenPlusMinus && removed[removedPrefixIndex+1] == '-' && added[addedPrefixIndex+1] == '+' {
			seenPlusMinus = true
			addedPrefixIndex++
			removedPrefixIndex++
		} else {
			break
		}
	}

	addedSuffixIndex, removedSuffixIndex := len(addedRune)-1, len(removedRune)-1
	for addedSuffixIndex > 0 && removedSuffixIndex > 0 {
		if loc := ColorPattern.FindStringIndex(added[:addedSuffixIndex]); loc != nil {
			addedPrefixIndex = loc[1]
		} else if loc := ColorPattern.FindStringIndex(removed[:removedSuffixIndex]); loc != nil {
			removedPrefixIndex -= loc[1]
		} else if addedRune[addedSuffixIndex] == removedRune[removedSuffixIndex] {
			addedSuffixIndex--
			removedSuffixIndex--
		} else {
			break
		}
	}

	return highlightLine(added, addedPrefixIndex, addedSuffixIndex, "black:green"), highlightLine(removed, removedPrefixIndex, removedSuffixIndex, "black:red")
}

func highlightLine(line string, prefix, suffix int, theme string) string {
	t := ansi.ColorCode(theme)
	reset := ansi.ColorCode("reset")
	return fmt.Sprint(line[:prefix], t, line[prefix:suffix+1], reset, line[suffix+1:])
}
