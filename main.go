package main

import (
	"log"
	_ "net/http/pprof"

	"github.com/sunshinev/go-space-chat/core"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	core.NewCore().Run()
}
