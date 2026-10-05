package main

import (
	"flag"
	"log"

	"github.com/HualiNox/cn-service-cidrs/internal/generator"
	"github.com/HualiNox/cn-service-cidrs/internal/parser"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	sourcesDir := flag.String("sources-dir", "./sources", "")
	output := flag.String("output", "./output", "")

	sourceFiles, err := parser.Parse(*sourcesDir)
	if err != nil {
		panic(err)
	}

	err = generator.Build(*output, sourceFiles)
	if err != nil {
		panic(err)
	}
}
