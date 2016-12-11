package parser

func FindNextTypeRecursive(fileContent string, parserPos int) int {
	var token string
	parserPos, token = FindNextToken(fileContent, parserPos)
	switch token {
	case ")", ",":
		return parserPos - 1
	case "[":
		{
			parserPos, _ = MatchClosing(token, "]", fileContent, parserPos-1)
			parserPos = FindNextTypeRecursive(fileContent, parserPos)
		}
	case "(":
		{
			parserPos, _ = MatchClosing(token, ")", fileContent, parserPos-1)
			parserPos = FindNextTypeRecursive(fileContent, parserPos)
		}
	default:
		parserPos = FindNextTypeRecursive(fileContent, parserPos)
	}
	return parserPos
}

func FindNextType(fileContent string, parserPos int) (int, string) {
	parserPos = SkipSpaces(fileContent, parserPos)
	endPos := FindNextTypeRecursive(fileContent, parserPos)
	return endPos, fileContent[parserPos:endPos]
}

func FindPackageName(fileContent string, parserPos int) (int, string) {
	var packageName string = ""
	var newPos, matched = MatchToken("package", fileContent, parserPos)

	if matched {
		newPos, packageName = FindNextToken(fileContent, newPos)
	}
	return newPos, packageName
}

func FindFunction(fileContent string, parserPos int) (int, Callable) {
	var function Function = NewFunction()
	var argument Argument
	var token string

	parserPos, function.Name = FindNextToken(fileContent, parserPos)
	for {
		parserPos, token = FindNextToken(fileContent, parserPos) // Skip '(' | ',' | ')'
		if token == ")" {
			break // Stop at the end of arguments
		}
		parserPos, argument.Name = FindNextToken(fileContent, parserPos)
		if argument.Name == ")" {
			break // Stop at the end of arguments
		}
		// If argument.name is not an identifier its a type,
		// the argument is unnamed
		if !IsValidIdentifier(argument.Name) {
			parserPos -= len(argument.Name)
			argument.Name = ""
		}
		parserPos, argument.Type = FindNextType(fileContent, parserPos)
		// if argument.Type is "" it means the identifier in
		// argument.name is in fact a type and
		// the argument is unnamed
		if argument.Type == "" {
			argument.Type = argument.Name
			argument.Name = ""
		}

		function.Arguments = append(function.Arguments, argument)
	}
	return parserPos, function
}

func FindMethod(fileContent string, parserPos int) (int, Method) {
	var method Method = NewMethod()
	var argument Argument
	var token string

	parserPos, token = FindNextToken(fileContent, parserPos) // Skip '('
	parserPos, method.Receiver.Name = FindNextToken(fileContent, parserPos)
	parserPos, method.Receiver.Type = FindNextType(fileContent, parserPos)

	for {
		parserPos, token = FindNextToken(fileContent, parserPos) // Skip '(' | ',' | ')'
		if token == ")" {
			break // Stop at the end of arguments
		}
		parserPos, argument.Name = FindNextToken(fileContent, parserPos)
		if argument.Name == ")" {
			break // Stop at the end of arguments
		}
		// If argument.name is not an identifier its a type,
		// the argument is unnamed
		if !IsValidIdentifier(argument.Name) {
			parserPos -= len(argument.Name)
			argument.Name = ""
		}
		parserPos, argument.Type = FindNextType(fileContent, parserPos)
		// if argument.Type is "" it means the identifier in
		// argument.name is in fact a type and
		// the argument is unnamed
		if argument.Type == "" {
			argument.Type = argument.Name
		}

		method.Arguments = append(method.Arguments, argument)
	}
	parserPos, method.Name = FindNextToken(fileContent, parserPos)
	return parserPos, method
}

// If no callable is found return a callable with zero value
func FindCallable(fileContent string, parserPos int) (int, Callable) {
	var declarationStart int
	var matched bool
	var functionOpener string
	var callable Callable

	declarationStart, matched = MatchToken("func", fileContent, parserPos)

	if matched {
		// Find only function in the outer scope of the file
		if fileContent[declarationStart-5] == '\n' {
			_, functionOpener = FindNextToken(fileContent, declarationStart)
			if functionOpener == "(" {
				parserPos, callable = FindMethod(fileContent, declarationStart)
			} else {
				parserPos, callable = FindFunction(fileContent, declarationStart)
			}
			return parserPos, callable
		} else {
			return FindCallable(fileContent, declarationStart)
		}
	}
	return declarationStart, callable
}
