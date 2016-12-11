// TODO: maybe replace strings by []byte when parsing the file?
// TODO: call the program without arguments to apply on all go file in the directory
// TODO: call the program with recursive option to apply it to all go file
// in the directory and sub directories
// TODO: call the program with directory/ies as argument/s
// TODO: call the program with package/s as argument/s

// TODO: doesn't work on anonymous functions
// TODO: doesn't work on functions not declared in the outer scope of the file.
// TODO: for now the program assume you don't have syntax errors
// should I add error handling of syntax error in the parsed code?

package main

import "os"
import "./parser"

func WriteFileInfos(fileName string, fileInfos []parser.FileInfo) {
	file, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	for _, fileInfo := range fileInfos {
		WriteFileInfo(*file, fileInfo)
	}
}

func WriteFileInfo(file os.File, fileInfo parser.FileInfo) {
	var infoToWrite string = "--------------------------------\n"
	infoToWrite += "File name : " + fileInfo.FileName + "\n"
	infoToWrite += "Language : " + fileInfo.Language + "\n"
	infoToWrite += "Package : " + fileInfo.PackageName + "\n"
	infoToWrite += "--------------------------------\n"

	for _, callable := range fileInfo.Callables {
		infoToWrite += callable.CallString() + "\n"
	}
	infoToWrite += "\n"

	_, err := file.WriteString(infoToWrite)

	if err != nil {
		panic(err)
	}
}

func main() {
	if len(os.Args) < 2 {
		panic("You need to give a file name as an argument.")
	}

	var fileInfos []parser.FileInfo = make([]parser.FileInfo, 0, 20)
	var fileNames []string = os.Args[1:len(os.Args)]

	for _, fileName := range fileNames {
		fileInfos = append(fileInfos, parser.GetFileInfo(fileName))
	}

	WriteFileInfos("crash_results.txt", fileInfos)
}
