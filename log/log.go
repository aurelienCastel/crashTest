package log

import "os"
import "../parser"

func WriteInfos(fileName string, fileInfos []parser.FileInfo) {
	file, err := os.Create(fileName)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	for _, fileInfo := range fileInfos {
		WriteInfo(*file, fileInfo)
	}
}

func WriteInfo(file os.File, fileInfo parser.FileInfo) {
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
