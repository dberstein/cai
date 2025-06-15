package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/dberstein/cai/chat"
	"github.com/dberstein/cai/provider"
	"github.com/fatih/color"
)

func main() {
	flag.Parse() // See provider/*.go for flags.
	ctx := context.Background()
	p, err := provider.NewGemini(ctx, nil)
	if err != nil {
		log.Println(color.RedString("%v", err))
		os.Exit(1)
	}
	c := chat.New(p)
	defer c.Close()
	for {
		if err := c.Run(); err != nil {
			log.Println(color.RedString("%v", err))
		}
	}
}
