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
	Append    string
	FmtSrc    string
	FmtDest   string
	KeyStart  int
	ByteStart int64
	Directory string
}

type IndexItem struct {
        ID     int    `json:"id"`
        RootID int    `json:"rootID"`
        OSIS   string `json:"osis"`
        Start  int64    `json:"start"`
        End    int64    `json:"end"`
        Type   string `json:"type"`
}

var Index map[string]IndexItem

type Translation struct {
        ShortCode     string
        Name          string
        Revision      string
        DatePublished string
}

type IndexFormat struct {
        Translation Translation       `json:"translation"`
        Index       map[int]IndexItem `json:"index"`
}

func main() {
	fl := new(flags)

	// Command Line Flags definition
	flag.StringVar(&fl.FmtSrc, "src-format", "usfm", "The source format")
	flag.StringVar(&fl.FmtDest, "dest-format", "json", "The destination format")
	flag.StringVar(&fl.Input, "i", "in.usfm", "Input file")
	flag.StringVar(&fl.Output, "o", "", "Output file (defaults to input filename with .json extension)")
	flag.StringVar(&fl.Append, "a", "", "Append output index to an index.json file (filename with .json extension)")
	flag.IntVar(&fl.KeyStart, "key-start", 0, "Starting key (root bible map, 0 == beginning)")
	flag.Int64Var(&fl.ByteStart, "byte-start", 0, "Offset the bytecount start (for calculation of future-conjoined USFM files)")
	flag.StringVar(&fl.Directory, "d", "", "Generate outputs for all files in the target directory (handles key iteration based on basic sort of directory list).")
	flag.Parse()

	// Options for JSON conversion
	o := json.Options{}

	var files []os.FileInfo
	var dir string
	var key = fl.KeyStart
	var byteStart = fl.ByteStart
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
	for i, file := range files {
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

			/*var index IndexFile

			// If we are appending then open and load the JSON
			if fl.Append != "" {
				// Open our source file
				appendFile, err := os.Open(fl.Append)
				if err != nil {
					log.Fatalf("Error reading input file %s: %s", fl.Append, err)
				}
				defer in.Close()

				bytes, _ := ioutil.ReadAll(appendFile)
				var index IndexFile
				json.Unmarshal([]byte(bytes), &IndexFile)
			}*/

			// We'll work out the byteStart if we're converting a directory
			if (fl.Directory != "") {
				if (i > 0) {
					byteStart = byteStart + files[i-1].Size()
				}
			}

			out, err := os.OpenFile(outfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("Error creating output file: %s", err)
			}
			defer out.Close()

			// Render and save
			key, err = json.Render(out, key, byteStart)
			if err != nil {
				log.Println(err)
			}
			log.Printf("Saved to %v", outfile)
		}
	}

}
