package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/socceroos/usfm/json"
)

// Command Line Flags
type flags struct {
	Input     string
	Output    string
	FmtSrc    string
	FmtDest   string
	KeyStart  int
	Directory string
}

func main() {
	fl := new(flags)

	// Command Line Flags definition
	flag.StringVar(&fl.FmtSrc, "src-format", "usfm", "The source format")
	flag.StringVar(&fl.FmtDest, "dest-format", "json", "The destination format")
	flag.StringVar(&fl.Input, "i", "in.usfm", "Input file")
	flag.StringVar(&fl.Output, "o", "", "Output file (defaults to input filename with .json extension)")
	flag.IntVar(&fl.KeyStart, "key-start", 0, "Starting key (root bible map, 0 == beginning)")
	flag.StringVar(&fl.Directory, "d", "", "Generate outputs for all files in the target directory (handles key iteration based on basic sort of directory list).")
	flag.Parse()

	// Options for JSON conversion
	o := json.Options{}

	var files []os.FileInfo
	var dir string
	var key = fl.KeyStart
	if fl.Directory != "" {
		dir = fl.Directory
		var err error
		files, err = ioutil.ReadDir(fl.Directory)
		if err != nil {
			log.Fatalf("Error reading directory at %s: %s", fl.Directory, err)
		}
	} else {
		dir = filepath.Dir(fl.Input)
		fInfo, err := os.Lstat(fl.Input)
		if err != nil {
			log.Fatalf("Error getting info for input file: %s", err)
		}
		files = append(files, fInfo)
	}

	// Go through each file and generate the output.
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".usfm" {
			// Open our source file
			in, err := os.Open(filepath.Join(dir, file.Name()))
			if err != nil {
				log.Fatalf("Error reading input file: %s", err)
			}
			defer in.Close()

			// Create a new renderer
			json := json.NewRenderer(o, in)

			// Create our out-file
			var outfile string
			if fl.Output == "" && fl.Directory != "" {
				var filename = file.Name()
				var ext = filepath.Ext(filename)
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
			key, err = json.Render(out, key)
			if err != nil {
				log.Println(err)
			}
			log.Printf("Saved to %v", outfile)
		}
	}

}
