package main

import (
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/pkg/errors"
)

func die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(-1)
}

func main() {
	if len(os.Args) != 3 {
		die(errors.New("usage: vanityd <config file> <listen addr>"))
	}

	h := HandleHotReloadFile(os.Args[1])
	l, err := net.Listen("tcp", os.Args[2])
	if err != nil {
		die(err)
	}

	err = http.Serve(l, h)
	if err != nil {
		die(err)
	}
}
