package main

import (
	"go-space-chat/core"
	"log"
	_ "net/http/pprof"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	c := &core.Core{}
	c.Run()
}
