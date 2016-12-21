// TODO: Add custom errors for when unexisting files / directory / options are specified
// TODO: maybe replace strings by []byte when parsing the file?
// TODO: call the program with package/s as argument/s

// TODO: doesn't work on anonymous functions
// TODO: doesn't work on functions not declared in the outer scope of the file.
// TODO: for now the program assume you don't have syntax errors
// should I add error handling of syntax error in the parsed code?

package main

import "os"
import "fmt"

import "github.com/aurelienCastel/fileDir"
import "github.com/aurelienCastel/stringUtil"
import "github.com/aurelienCastel/errorUtil"

import "github.com/aurelienCastel/crashTest/parser"
import "github.com/aurelienCastel/crashTest/log"

func FileNamesToParse(args []string, extensions []string) []string {
	var fileNames []string
	var file *os.File
	var err error
	var currentDir *os.File = fileDir.CurrentDir()

	if len(args) == 0 {
		fileNames = fileDir.RelativeNamesInDirWithExts(currentDir, extensions)
	} else {
		if args[0] == "rec:" {
			if len(args) == 1 {
				fileNames = fileDir.RelativeNamesInRecDirWithExts(currentDir, extensions)
			} else {
				for _, fileName := range args[1:len(args)] {
					if fileDir.NameIsDir(fileName) {
						file, err = os.Open(fileName)
						errorUtil.Check(err)
						fileNames = append(fileNames, fileDir.RelativeNamesInRecDirWithExts(file, extensions)...)
					} else if stringUtil.HasOneOfSuffixes(fileName, extensions) {
						fileNames = append(fileNames, fileName)
					} else {
						fmt.Printf("%s is written in a language not yet supported.", fileName)
					}
				}
			}
		} else {
			for _, fileName := range args {
				if fileDir.NameIsDir(fileName) {
					file, err = os.Open(fileName)
					errorUtil.Check(err)
					fileNames = append(fileNames, fileDir.RelativeNamesInDirWithExts(file, extensions)...)
				} else if stringUtil.HasOneOfSuffixes(fileName, extensions) {
					fileNames = append(fileNames, fileName)
				} else {
					fmt.Printf("%s is written in a language not yet supported.", fileName)
				}
			}
		}
	}

	return fileNames
}

func main() {
	extensions := []string{".go"}
	var fileNames []string
	var fileInfos []parser.FileInfo = make([]parser.FileInfo, 0, 20)

	fileNames = FileNamesToParse(os.Args[1:len(os.Args)], extensions)

	for _, fileName := range fileNames {
		fileInfos = append(fileInfos, parser.GetFileInfo(fileName))
	}

	log.WriteInfos("crash_results.txt", fileInfos)
}
