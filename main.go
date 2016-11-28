
package main

import (
	"bufio"
	"os"
	"fmt"
)

func main() {
	l := NewLayout("GoRegular")

	err := l.Print(os.Stdout)
	if err != nil {
		panic(err)
	}
	doc, err := parse("-", bufio.NewReader(os.Stdin))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	for _, n := range doc {
		n.Emit(os.Stdout)
	}
	os.Stdout.WriteString("\n")
}
