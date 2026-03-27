package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	var b bytes.Buffer
	err := dirTree(&b, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
	b.WriteTo(out)
}

func dirTree(buf *bytes.Buffer, dirName string, printFiles bool) error {
	depthLevel := 1
	parentPath := ""
	indentStr := ""
	str, err := recurseSubDirectory(parentPath, dirName, printFiles, indentStr, depthLevel, false)
	if err != nil {
		return err
	}
	_, err = buf.WriteString(fmt.Sprint(str))
	return err
}

func recurseSubDirectory(parentPath string, subdir string, printFiles bool, indentStr string, depth int, isLastEntry bool) (string, error) {
	fullDirPath := subdir
	listStr := ""
	// skip this for root directory only
	if parentPath != "" {
		// Render current directory and update path indent string
		fullDirPath = filepath.Join(parentPath, subdir)
		listStr = buildEntryLine(subdir, indentStr, isLastEntry, -1)
		if isLastEntry {
			indentStr += "\t"
		} else {
			indentStr += "│\t"
		}
	}

	c, err := os.ReadDir(fullDirPath)
	if err != nil {
		return listStr, err
	}
	// Filter directories for directoriesOnly mode
	var listEntries []os.DirEntry
	if printFiles {
		listEntries = c
	} else {
		for i := 0; i < len(c); i++ {
			if c[i].IsDir() {
				listEntries = append(listEntries, c[i])
			}
		}
	}
	// Iterate and recurse
	for i := 0; i < len(listEntries); i++ {
		isCurrentLastEntry := i == len(listEntries)-1
		if listEntries[i].IsDir() {
			subDirStr, err := recurseSubDirectory(fullDirPath, listEntries[i].Name(), printFiles, indentStr, depth+1, isCurrentLastEntry)
			if err != nil {
				return listStr, err
			}
			listStr += subDirStr
		} else if printFiles {
			fileSize, err := fileSize(fullDirPath, listEntries[i].Name())
			if err != nil {
				return listStr, err
			}
			listStr += buildEntryLine(listEntries[i].Name(), indentStr, isCurrentLastEntry, fileSize)
		}
	}
	return listStr, nil
}

func fileSize(fullDirPath string, filename string) (int64, error) {

	fi, err := os.Stat(filepath.Join(fullDirPath, filename))
	if err != nil {
		return -2, err
	}
	size := fi.Size()
	return size, nil
}

func buildEntryLine(entryName string, indentStr string, isLastEntry bool, fileSize int64) string {
	symbol := "├───"
	if isLastEntry {
		symbol = "└───"
	}
	entryStr := indentStr + symbol + entryName
	if fileSize > -1 {
		if fileSize == 0 {
			entryStr += " (empty)"
		} else {
			entryStr += fmt.Sprint(" (", fileSize, "b)")
		}
	}
	return entryStr + "\n"
}
