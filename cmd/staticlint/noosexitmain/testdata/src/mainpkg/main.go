package main

import (
	"log"
	"os"
)

func main() {
	os.Exit(1)
	log.Fatal("x")
	panic("boom")
}
