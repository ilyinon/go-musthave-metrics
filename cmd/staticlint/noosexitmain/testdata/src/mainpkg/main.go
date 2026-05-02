package a

import (
	"log"
	"os"
)

func main() {
	os.Exit(1)     // want "do not call os.Exit outside main.main"
	log.Fatal("x") // want "do not call log.Fatal outside main.main"
	panic("boom")  // want "do not call panic outside main.main"
}
