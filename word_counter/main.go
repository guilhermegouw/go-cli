package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	lines := flag.Bool("l", false, "Count lines")
	bytes := flag.Bool("b", false, "Count bytes")
	flag.Parse()
	count, err := count(os.Stdin, *lines, *bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(count)
}

func count(r io.Reader, countLines, countBytes bool) (int, error) {
	scanner := bufio.NewScanner(r)

	switch {
	case countLines:
		scanner.Split(bufio.ScanLines)
	case countBytes:
		scanner.Split(bufio.ScanBytes)
	default:
		scanner.Split(bufio.ScanWords)
	}
	count := 0

	for scanner.Scan() {
		count++
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}
