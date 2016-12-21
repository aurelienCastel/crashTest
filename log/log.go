package log

import "os"
import "github.com/aurelienCastel/crashTest/parser"
import "github.com/aurelienCastel/errorUtil"

func WriteInfos(fileName string, fileInfos []parser.FileInfo) {
	file, err := os.Create(fileName)
	errorUtil.Check(err)

	for _, fileInfo := range fileInfos {
		WriteInfo(*file, fileInfo)
	}

	err = file.Close()
	errorUtil.Check(err)
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
	errorUtil.Check(err)
}
