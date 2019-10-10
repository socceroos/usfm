package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	flag.StringVar(&args.FmtDest, "f", "index.json", "The destination format")
	flag.StringVar(&args.Append, "a", "", "Append output index to an index.json file (filename with .json extension)")
	flag.IntVar(&args.KeyStart, "k", 0, "Starting key (root bible map, 0 == beginning)")
	flag.Int64Var(&args.ByteStart, "b", 0, "Offset the bytecount start (for calculation of future-conjoined USFM files)")
	flag.StringVar(&args.Directory, "d", "./translations", "Generate outputs for all files in the target directory (handles key iteration based on basic sort of directory list).")
	flag.Parse()

	// Options for JSON conversion
	dir := args.Directory
	indexExtension := args.FmtDest

	var folders []os.FileInfo

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
		startKey := args.KeyStart
		startByte := args.ByteStart

		// Recursively scans files in sub-folders
		if folder.IsDir() {
			// Init JSON index
			fatJSON := json.NewJSON()

			// 1. Generate file paths
			fatIndexPath := filepath.Join(dir, folder.Name()+"."+indexExtension)
			fatUsfmPath := filepath.Join(dir, folder.Name()+".usfm")
			translationInfoPath := filepath.Join(dir, folder.Name(), "00-translation.info")
			folderFullPath := filepath.Join(dir, folder.Name())

			// 2. Clean up generated files
			// Try to delete previously generated file
			os.Remove(fatIndexPath)
			os.Remove(fatUsfmPath)
			temps, err := filepath.Glob(filepath.Join(folderFullPath, "*.temp"))
			if err != nil {
				panic(err)
			}
			for _, temp := range temps {
				if err := os.Remove(temp); err != nil {
					panic(err)
				}
			}

			// Init new fat usfm file
			fatUsfm, _ := os.OpenFile(fatUsfmPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

			// 3. Parse usfm files
			// Iterate usfm files path under current sub-folder
			fatJSON.ReadTranslation(translationInfoPath)

			files, _ := filepath.Glob(filepath.Join(folderFullPath, "*.usfm"))
			for _, file := range files {
				// Append text from each book into a single fat usfm
				//f, _ := os.Open(file)
				//b, _ := ioutil.ReadAll(f)
				//
				//fatUsfm.Write(b)
				newFile := formatContent(file, fatUsfm)

				// Parse and append json data into fat JSON file
				startKey, startByte = fatJSON.AppendUsfmIndex(newFile, startKey, startByte)
			}

			defer fatUsfm.Close()

			// Generate fat json
			//startKey, startByte = fatJSON.AppendUsfmIndex(fatUsfmPath, startKey, startByte)
			fatJSON.WriteFile(fatIndexPath)
		}
	}
}

// Format content
func formatContent(filePath string, writer *os.File) string {
	// Log file location
	log.Print(filePath)

	// Create temp usfm file which will be formated
	newFilePath := filePath + ".temp"
	tempUSFM, _ := os.OpenFile(newFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	// Open the original file
	f, err := os.Open(filePath)

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
		return ""
	}

	buffer, _ := ioutil.ReadAll(f)
	content := string(buffer)

	// Replace ¶ with empty string
	content = strings.Replace(content, "¶", "", -1)
	content = strings.Replace(content, "  ", " ", -1)
	content = removeWTag(content)
	content = removeWPlusTag(content)

	writer.Write([]byte(content))
	tempUSFM.Write([]byte(content))

	// Close file
	defer f.Close()
	defer tempUSFM.Close()

	return newFilePath
}

func removeWTag(content string) string {
	start, end := "\\w ", "\\w*" // just replace these with whatever you like...
	sSplits := strings.Split(content, start)
	result := ""

	if len(sSplits) > 1 { // n splits = 1 means start char not found!
		for _, subStr := range sSplits { // check each substring for end

			ixEnd := strings.Index(subStr, end)
			ixSplit := strings.Index(subStr, "|")
			if ixEnd != -1 && ixSplit != -1 {
				result += subStr[0:ixSplit]
				result += subStr[ixEnd+3:(len(subStr))]
			} else {
				result += subStr
			}
		}
	} else {
		return content
	}
	return result
}

func removeWPlusTag(content string) string {
	start, end := "\\+w ", "\\+w*" // just replace these with whatever you like...
	sSplits := strings.Split(content, start)
	result := ""

	if len(sSplits) > 1 { // n splits = 1 means start char not found!
		for _, subStr := range sSplits { // check each substring for end

			ixEnd := strings.Index(subStr, end)
			ixSplit := strings.Index(subStr, "|")
			if ixEnd != -1 && ixSplit != -1 {
				result += subStr[0:ixSplit]
				result += subStr[ixEnd+4:(len(subStr))]
			} else {
				result += subStr
			}
		}
	} else {
		return content
	}
	return result
}

// handleError if error not null, returns the format string
func handleError(err error, format string, v ...interface{}) {
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		log.Fatalf(format, v...)
	}
}
