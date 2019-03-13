package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/socceroos/usfm/index"
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
	Recursive bool
}

func main() {
	args := new(flags)

	// Parse command line args
	flag.StringVar(&args.FmtSrc, "src-format", "usfm", "The source format")
	flag.StringVar(&args.FmtDest, "dest-format", "json", "The destination format")
	flag.StringVar(&args.Append, "a", "", "Append output index to an index.json file (filename with .json extension)")
	flag.IntVar(&args.KeyStart, "key-start", 0, "Starting key (root bible map, 0 == beginning)")
	flag.Int64Var(&args.ByteStart, "byte-start", 0, "Offset the bytecount start (for calculation of future-conjoined USFM files)")
	flag.StringVar(&args.Directory, "d", ".\\translations", "Generate outputs for all files in the target directory (handles key iteration based on basic sort of directory list).")
	flag.Parse()

	// Options for JSON conversion
	dir := args.Directory

	var folders []os.FileInfo
	// var key = args.KeyStart
	// var byteStart = args.ByteStart

	// Load files from provided directory or input file
	if dir != "" {
		// Read the provided directory
		dir = args.Directory
		var err error
		folders, err = ioutil.ReadDir(dir)

		handleError(err, "Error reading directory at %s", args.Directory)
	} else {
		// Exit program if directory is not provided
		log.Fatalf("Reading directory is not provided")
	}

	// Go through each file and generate the output.
	for _, folder := range folders {
		// Recursive scans files in sub-folders
		if folder.IsDir() {
			folderFullPath := filepath.Join(dir, folder.Name())

			// Delete all json files that are previously generated
			jsons, _ := filepath.Glob(filepath.Join(folderFullPath, "*.json"))
			for _, json := range jsons {
				if err := os.Remove(json); err != nil {
					handleError(err, "Cannot delete file %s", json)
				}
			}

			// Get usfm files path under current sub-folder
			files, _ := filepath.Glob(filepath.Join(folderFullPath, "*.usfm"))

			for _, file := range files {
				renderJSON(file)
			}
		}
	}
	// if filepath.Ext(file.Name()) == ".usfm" {
	// 	// Open our source file
	// 	in, err := os.Open(filepath.Join(dir, file.Name()))
	// 	if err != nil {
	// 		log.Fatalf("Error reading input file: %s", err)
	// 	}
	// 	defer in.Close()
	//
	// 	// Create a new renderer
	// 	json := json.NewRenderer(o, in)
	//
	// 	// Create our out-file
	// 	var outfile string
	// 	if args.Output == "" && args.Directory != "" {
	// 		var filename = file.Name()
	// 		var ext = filepath.Ext(filename)
	// 		var name = filename[0 : len(filename)-len(ext)]
	// 		outfile = filepath.Join(dir, name+".json")
	// 	} else {
	// 		outfile = args.Output
	// 	}

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
	// 	if args.Directory != "" {
	// 		if i > 0 {
	// 			byteStart = byteStart + folders[i-1].Size()
	// 		}
	// 	}
	//
	// 	out, err := os.OpenFile(outfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	// 	if err != nil {
	// 		log.Fatalf("Error creating output file: %s", err)
	// 	}
	// 	defer out.Close()
	//
	// 	// Render and save
	// 	key, err = json.Render(out, key, byteStart)
	// 	if err != nil {
	// 		log.Println(err)
	// 	}
	// 	log.Printf("Saved to %v", outfile)
	// }
}

func renderJSON(inPath string) {
	// Open our source file
	in, err := os.Open(inPath)
	handleError(err, "Error reading input file")

	defer in.Close()

	// Create our out-file
	ext := filepath.Ext(inPath)
	outPath := inPath[0:len(inPath)-len(ext)] + ".json"
	log.Printf(outPath)

	render := index.NewRenderer(in)

	// if args.Directory != "" {
	// 	if i > 0 {
	// 		byteStart = byteStart + folders[i-1].Size()
	// 	}
	// }

	out, err := os.OpenFile(outPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error creating output file: %s", err)
	}
	defer out.Close()

	// Render and save
	_, err = render.Render(out, 0, 0)
	if err != nil {
		log.Println(err)
	}
	log.Printf("Saved to %v", outPath)
}

func handleError(err error, format string, v ...interface{}) {
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		log.Fatalf(format, v...)
	}
}
