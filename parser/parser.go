package parser

import "io/ioutil"
import "github.com/aurelienCastel/stringUtil"

func IsSpace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n'
}

func IsAlpha(char byte) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

func IsDigit(char byte) bool {
	return (char >= '0' && char <= '9')
}

func IsAlphanum(char byte) bool {
	return IsAlpha(char) || IsDigit(char)
}

func IsValidIdentStart(char byte) bool {
	return IsAlpha(char) || char == '_'
}

func IsValidIdentChar(char byte) bool {
	return IsValidIdentStart(char) || IsDigit(char)
}

func IsValidIdentifier(token string) bool {
	if len(token) == 0 || !IsValidIdentStart(token[0]) {
		return false
	}
	for i := 1; i < len(token); i++ {
		if !IsValidIdentChar(token[i]) {
			return false
		}
	}
	return true
}

func SkipSpaces(fileContent string, parserPos int) int {
	for ; parserPos < len(fileContent); parserPos++ {
		if !IsSpace(fileContent[parserPos]) {
			return parserPos
		}
	}
	return parserPos
}

// TODO: This is probably customisable, if it doesn't fit,
//       duplicate it for each language
func SkipString(fileContent string, parserPos int) int {
	var stringStart byte = fileContent[parserPos]

	if stringStart == '"' {
		parserPos++
		for ; parserPos < len(fileContent); parserPos++ {
			// Handle the case of \ in interpreted strings
			if fileContent[parserPos] == '\\' {
				parserPos++
				continue
			}
			if fileContent[parserPos] == stringStart {
				return parserPos + 1
			}
		}
	} else if stringStart == '`' {
		parserPos++
		for ; parserPos < len(fileContent); parserPos++ {
			if fileContent[parserPos] == stringStart {
				return parserPos + 1
			}
		}
	}
	return parserPos
}

// TODO: This is probably customisable, if it doesn't fit,
//       duplicate it for each language
func SkipComment(fileContent string, parserPos int) int {
	commentMap := map[string]string{
		"//": "\n",
		"/*": "*/",
	}
	var commentStart []byte

	if fileContent[parserPos] == '/' {
		commentStart = append(commentStart, fileContent[parserPos])
		parserPos++
		if fileContent[parserPos] == '/' || fileContent[parserPos] == '*' {
			commentStart = append(commentStart, fileContent[parserPos])
			parserPos++
			parserPos, _ = Match(commentMap[string(commentStart)], fileContent, parserPos)
		}
	}
	return parserPos
}

func SkipNonCode(fileContent string, parserPos int) int {
	var newPos int

	newPos = SkipSpaces(fileContent, parserPos)
	newPos = SkipString(fileContent, newPos)
	newPos = SkipComment(fileContent, newPos)

	if parserPos != newPos {
		return SkipNonCode(fileContent, newPos)
	}
	return newPos
}

// Move to the character after the target and return true,
// or move to the end of the file and return false
// (doesn't skip 'non code')
func Match(target string, fileContent string, parserPos int) (int, bool) {
	var i int = 0

	for ; parserPos < len(fileContent); parserPos++ {
		if fileContent[parserPos] == target[i] {
			i++
			if i == len(target) {
				return parserPos + 1, true
			}
		} else {
			i = 0
		}
	}
	return parserPos, false
}

// Move to the character after the target and return true,
// or move to the end of the file and return false
// (look for target only in code, not in strings or comments)
func MatchToken(target string, fileContent string, parserPos int) (int, bool) {
	var token string

	for ; parserPos < len(fileContent); parserPos++ {
		parserPos, token = FindNextToken(fileContent, parserPos)
		if target == token {
			return parserPos, true
		}
	}
	return parserPos, false
}

// You should call this BEFORE the parserPos is on an opening token.
func MatchClosing(opening string, closing string, fileContent string, parserPos int) (int, bool) {
	var i int = 0
	var j int = 0
	var nestingLevel int = 0

	for ; parserPos < len(fileContent); parserPos++ {
		if fileContent[parserPos] == opening[i] {
			i++
			if i == len(opening) {
				nestingLevel++
				i = 0
				j = 0
			}
		} else {
			i = 0
		}
		if fileContent[parserPos] == closing[j] {
			j++
			if j == len(closing) {
				nestingLevel--
				if nestingLevel == 0 {
					return parserPos + 1, true
				}
				i = 0
				j = 0
			}
		} else {
			j = 0
		}
	}
	return parserPos, false
}

func FindNextToken(fileContent string, parserPos int) (int, string) {
	var token []byte

	parserPos = SkipNonCode(fileContent, parserPos)

	if IsValidIdentStart(fileContent[parserPos]) {
		token = append(token, fileContent[parserPos])
		for parserPos++; parserPos < len(fileContent); parserPos++ {
			if !IsValidIdentChar(fileContent[parserPos]) {
				break
			}
			token = append(token, fileContent[parserPos])
		}
	} else if IsDigit(fileContent[parserPos]) {
		token = append(token, fileContent[parserPos])
		for parserPos++; parserPos < len(fileContent); parserPos++ {
			if !IsDigit(fileContent[parserPos]) {
				break
			}
			token = append(token, fileContent[parserPos])
		}
	} else {
		token = append(token, fileContent[parserPos])
		parserPos++
	}
	return parserPos, string(token)
}

func GetFileInfo(fileName string) FileInfo {
	var parserPos int = 0
	var fileInfo FileInfo

	fileContent, err := ioutil.ReadFile(fileName)
	fileContent = append(fileContent, '\x00')

	if err != nil {
		panic(err)
	}

	fileInfo.FileName = fileName
	fileInfo.Language = stringUtil.FileExtension(fileName)
	parserPos, fileInfo.PackageName = FindPackageName(string(fileContent), parserPos)

	var function Callable
	for ; parserPos < len(fileContent); parserPos++ {
		parserPos, function = FindCallable(string(fileContent), parserPos)
		fileInfo.Callables = append(fileInfo.Callables, function)
	}
	fileInfo.Callables = fileInfo.Callables[0 : len(fileInfo.Callables)-1]
	return fileInfo
}
