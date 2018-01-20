package main

import (
	"bufio"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/knsh14/diff-highlight-go/highlight"
)

func main() {
	signal.Ignore(syscall.SIGPIPE)

	scanner := bufio.NewScanner(os.Stdin)
	err := highlight.DiffHighlight(scanner)
	if err != nil {
		log.Fatal(err)
	}
}
