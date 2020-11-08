package main

import (
	"go-space-chat/core"
	"log"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	core.InitCore()
}
