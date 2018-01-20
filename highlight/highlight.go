package highlight

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
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

// DiffHighlight is main program for diff-highlight
func DiffHighlight(scanner *bufio.Scanner) error {
	dc := &DiffContext{InHunk: false, Added: []string{}, Removed: []string{}}

	for scanner.Scan() {
		t := scanner.Text()
		ul := unescapedLine(t)
		err := dc.handleLine(ul)
		if err != nil {
			return errors.Wrap(err, "failed handle line")
		}
	}
	if err := scanner.Err(); err != nil {
		return errors.Wrap(err, "failed scan")
	}
	dc.ShowHunk()
	return nil
}

func unescapedLine(s string) string {
	quoted := strconv.Quote(s)
	return quoted[1 : len(quoted)-1]
}

func (dc *DiffContext) handleLine(input string) error {
	if !dc.InHunk {
		err := printQuotedLine(input)
		if err != nil {
			return errors.Wrap(err, "print failed in outside of hunk")
		}
		dc.InHunk = hunkPattern.MatchString(input)
		return nil
	}
	if removedLinePattern.MatchString(input) {
		dc.Removed = append(dc.Removed, input)
		return nil
	}
	if addedLinePattern.MatchString(input) {
		dc.Added = append(dc.Added, input)
		return nil
	}
	err := dc.ShowHunk()
	if err != nil {
		return errors.Wrap(err, "print failed in outside of hunk")
	}
	err = printQuotedLine(input)
	if err != nil {
		return errors.Wrap(err, "print failed in hunk")
	}
	dc.Added = []string{}
	dc.Removed = []string{}
	dc.InHunk = nextHunkPattern.MatchString(input)
	return nil
}

// ShowHunk shows lien in dc
func (dc *DiffContext) ShowHunk() error {
	if len(dc.Added) == 0 || len(dc.Removed) == 0 {
		for _, v := range dc.Removed {
			err := printQuotedLine(v)
			if err != nil {
				return errors.Wrap(err, "print failed in hunk has no added line")
			}
		}
		for _, v := range dc.Added {
			err := printQuotedLine(v)
			if err != nil {
				return errors.Wrap(err, "print failed in hunk has no removed line")
			}
		}
		return nil
	}

	if len(dc.Added) != len(dc.Removed) {
		for _, v := range dc.Removed {
			err := printQuotedLine(v)
			if err != nil {
				return errors.Wrap(err, "print removed line failed in hunk has diffelent length")
			}
		}
		for _, v := range dc.Added {
			err := printQuotedLine(v)
			if err != nil {
				return errors.Wrap(err, "print removed line failed in hunk has diffelent length")
			}
		}
		return nil
	}

	var queue []string
	for i := 0; i < len(dc.Added); i++ {
		a, r := dc.highlighPair(dc.Added[i], dc.Removed[i])
		err := printQuotedLine(r)
		if err != nil {
			return errors.Wrap(err, "print highlighted removed line failed")
		}
		queue = append(queue, a)
	}
	for _, v := range queue {
		err := printQuotedLine(v)
		if err != nil {
			return errors.Wrap(err, "print highlighted added line failed")
		}
	}
	return nil
}

func printQuotedLine(s string) error {
	v, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		return errors.Wrap(err, "failed Unquote text")
	}
	_, err = fmt.Fprintln(os.Stdout, v)
	if err != nil {
		return errors.Wrap(err, "failed print text")
	}
	return nil
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
