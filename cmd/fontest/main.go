package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/kenkyu392/fontest"
)

func main() {
	log.SetFlags(0)
	err := fontest.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
	if err != nil && err != flag.ErrHelp {
		log.Fatal(err)
		os.Exit(1)
	}
}
