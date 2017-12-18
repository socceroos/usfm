package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/socceroos/usfm/json"
)

// Command Line Flags
type flags struct {
	Input    string
	Output   string
	FmtSrc   string
	FmtDest  string
	KeyStart int
}

func main() {
	fl := new(flags)

	// Command Line Flags definition
	flag.StringVar(&fl.FmtSrc, "src-format", "usfm", "The source format")
	flag.StringVar(&fl.FmtDest, "dest-format", "json", "The destination format")
	flag.StringVar(&fl.Input, "i", "in.usfm", "Input file")
	flag.StringVar(&fl.Output, "o", "", "Output file (defaults to input filename with .json extension)")
	flag.IntVar(&fl.KeyStart, "key-start", 0, "Starting key (root bible map, 0 == beginning)")
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
	var outfile string
	if fl.Output == "" {
		var dir = filepath.Dir(fl.Input)
		var filename = filepath.Base(fl.Input)
		var ext = filepath.Ext(fl.Input)
		var name = filename[0 : len(filename)-len(ext)]
		outfile = filepath.Join(dir, name+".json")
	} else {
		outfile = fl.Output
	}
	out, err := os.Create(outfile)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer out.Close()

	// Render and save
	err = json.Render(out, fl.KeyStart)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Saved to %v", outfile)
}
