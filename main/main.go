package main

import (
	"flag"
	"log"
	"os"

	"github.com/dberstein/cai/conversation"
	"github.com/dberstein/cai/provider"
	"github.com/fatih/color"
)

func main() {
	model := flag.String("m", "models/gemini-2.5-pro-preview-06-05", "model to use")
	flag.Parse()

	p, err := provider.New(model)
	if err != nil {
		log.Println(color.RedString("%v", err))
		os.Exit(1)
	}
	c := conversation.New(p)
	defer c.Close()

	for {
		if err := c.Run(model); err != nil {
			log.Println(color.RedString("%v", err))
		}
	}
}
