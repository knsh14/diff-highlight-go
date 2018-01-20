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
)

var (
	headColorPattern   = regexp.MustCompile(`^\\x1b\[[0-9;]*m`)
	tailColorPattern   = regexp.MustCompile(`\\x1b\[[0-9;]*m$`)
	hunkPattern        = regexp.MustCompile(`^(\\x1b\[[0-9;]*m)*\@\@`)
	addedLinePattern   = regexp.MustCompile(`^(\\x1b\[[0-9;]*m)*\+`)
	removedLinePattern = regexp.MustCompile(`^(\\x1b\[[0-9;]*m)*-`)
	nextHunkPattern    = regexp.MustCompile(`^(\\x1b\[[0-9;]*m)*[\@ ]`)
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
		dc.handleLine(ul)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	dc.ShowHunk()
}

// UnescapedLine retuns line with color info
func UnescapedLine(s string) string {
	quoted := strconv.Quote(s)
	return strings.Trim(quoted, "\"")
}

func (dc *DiffContext) handleLine(input string) {
	if !dc.InHunk {
		printQuotedLine(input)
		dc.InHunk = hunkPattern.MatchString(input)
		return
	}
	if removedLinePattern.MatchString(input) {
		dc.Removed = append(dc.Removed, input)
		return
	}
	if addedLinePattern.MatchString(input) {
		dc.Added = append(dc.Added, input)
		return
	}
	dc.ShowHunk()
	printQuotedLine(input)
	dc.Added = []string{}
	dc.Removed = []string{}
	dc.InHunk = nextHunkPattern.MatchString(input)
}

// ShowHunk shows lien in dc
func (dc *DiffContext) ShowHunk() {
	if len(dc.Added) == 0 || len(dc.Removed) == 0 {
		for _, v := range dc.Removed {
			printQuotedLine(v)
		}
		for _, v := range dc.Added {
			printQuotedLine(v)
		}
		return
	}

	if len(dc.Added) != len(dc.Removed) {
		for _, v := range dc.Removed {
			printQuotedLine(v)
		}
		for _, v := range dc.Added {
			printQuotedLine(v)
		}
		return
	}

	var queue []string
	for i := 0; i < len(dc.Added); i++ {
		a, r := dc.highlighPair(dc.Added[i], dc.Removed[i])
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
	_, err = fmt.Fprintln(os.Stdout, v)
	if err != nil {
		fmt.Println(s)
		panic(err)
	}
}

func (dc *DiffContext) highlighPair(added, removed string) (string, string) {
	seenPlusMinus := false
	apIndex, rpIndex := 0, 0
	for apIndex < len(added) && rpIndex < len(removed) {
		if loc := headColorPattern.FindStringIndex(added[apIndex:]); loc != nil {
			apIndex += loc[1]
		} else if loc := headColorPattern.FindStringIndex(removed[rpIndex:]); loc != nil {
			rpIndex += loc[1]
		} else if added[apIndex] == removed[rpIndex] {
			apIndex++
			rpIndex++
		} else if !seenPlusMinus && removed[rpIndex] == '-' && added[apIndex] == '+' {
			seenPlusMinus = true
			apIndex++
			rpIndex++
		} else {
			break
		}
	}

	asIndex, rsIndex := len(added)-1, len(removed)-1
	for asIndex > apIndex && rsIndex > rpIndex {
		if loc := tailColorPattern.FindStringIndex(added[:asIndex]); loc != nil {
			apIndex = loc[1] - 1
		} else if loc := tailColorPattern.FindStringIndex(removed[:rsIndex]); loc != nil {
			rpIndex = loc[1] - 1
		} else if added[asIndex] == removed[rsIndex] {
			asIndex--
			rsIndex--
		} else {
			break
		}
	}

	return getHighlightedLine(added, apIndex, asIndex), getHighlightedLine(removed, rpIndex, rsIndex)
}

func getHighlightedLine(line string, prefix, suffix int) string {
	return fmt.Sprint(line[:prefix], "\x1b[7m", line[prefix:suffix+1], "\x1b[27m", line[suffix+1:])
}
