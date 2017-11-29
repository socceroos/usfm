package main

import (
	"flag"
	"log"
	"os"

	"github.com/socceroos/usfm/json"
)

// Command Line Flags
type flags struct {
	Input   string
	Output  string
	FmtSrc  string
	FmtDest string
}

func main() {
	fl := new(flags)

	// Command Line Flags definition
	flag.StringVar(&fl.FmtSrc, "src-format", "usfm", "The source format")
	flag.StringVar(&fl.FmtDest, "dest-format", "json", "The destination format")
	flag.StringVar(&fl.Input, "i", "in.usfm", "Input file")
	flag.StringVar(&fl.Output, "o", "out.json", "Output file")
	flag.Parse()

	// Options for JSON conversion
	o := json.Options{}

	// Open our source file
	in, err := os.Open(fl.Input)
	if err != nil {
		log.Fatalf("Error reading input file: %s", err)
	}
	defer in.Close()

	// Create a new renderer
	json := json.NewRenderer(o, in)

	// Create our out-file
	out, err := os.Create(fl.Output)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer out.Close()

	// Render and save
	err = json.Render(out)
	if err != nil {
		log.Println(err)
	}
}
