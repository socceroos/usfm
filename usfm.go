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
	flag.StringVar(&args.Directory, "d", "./translations", "Generate outputs for all files in the target directory (handles key iteration based on basic sort of directory list).")
	flag.Parse()

	// Options for JSON conversion
	dir := args.Directory

	var folders []os.FileInfo
	startKey := args.KeyStart
	startByte := args.ByteStart

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
		// Recursively scans files in sub-folders
		if folder.IsDir() {
			// Init JSON index
			fatJSON := json.NewJSON()

			// 1. Generate file paths
			fatIndexPath := filepath.Join(dir, folder.Name()+".json")
			fatUsfmPath := filepath.Join(dir, folder.Name()+".usfm")
			translationInfoPath := filepath.Join(dir, folder.Name(), "00-translation.info")
			folderFullPath := filepath.Join(dir, folder.Name())

			fatUsfm, _ := os.OpenFile(fatUsfmPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			// 2. Clean up generated files
			// Try to delete previously generated file
			os.Remove(fatIndexPath)

			// Delete all json files that are previously generated
			jsons, _ := filepath.Glob(filepath.Join(folderFullPath, "*.json"))
			for _, json := range jsons {
				os.Remove(json)
			}

			// 3. Parse usfm files
			// Iterate usfm files path under current sub-folder
			fatJSON.ReadTranslation(translationInfoPath)

			files, _ := filepath.Glob(filepath.Join(folderFullPath, "*.usfm"))
			for _, file := range files {
				// Append text from each book into a single fat usfm
				f, _ := os.Open(file)
				b, _ := ioutil.ReadAll(f)
				fatUsfm.Write(b)

				// Parse and append json data into fat JSON file
				appendJSONData(fatJSON, file, &startKey, &startByte)
			}

			fatJSON.WriteFile(fatIndexPath)
			defer fatUsfm.Close()
		}
	}
}

func appendJSONData(j *json.JSON, inPath string, startKey *int, startByte *int64) {
	// Create our out-file
	ext := filepath.Ext(inPath)
	bookPath := inPath[0:len(inPath)-len(ext)] + ".json"
	log.Printf(bookPath)

	*startKey, *startByte = j.AppendUsfmIndex(inPath, *startKey, *startByte)
}

// handleError if error not null, returns the format string
func handleError(err error, format string, v ...interface{}) {
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		log.Fatalf(format, v...)
	}
}
