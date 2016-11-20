package main

import "fmt"
import "io/ioutil"
import "os"

// TODO: The get_test_info that crate the TestInfo struct add
//       a null callable at the end of the callable array,
//       (because there is no more callable to be found)
//       should be removed.
type Argument struct {
	name  string
	_type string
}

type Function struct {
	name      string
	arguments []Argument
}

func new_function() Function {
	var function Function
	function.arguments = make([]Argument, 0, 2)
	return function
}

func (function Function) call_string() string {
	var call_string string
	var i int = 0

	call_string += function.name + "("
	for _, argument := range function.arguments {
		call_string += argument.name + " " + argument._type
		i++
		if i < len(function.arguments) {
			call_string += ", "
		}
	}
	call_string += ")"
	return call_string
}

type Method struct {
	name      string
	receiver  Argument
	arguments []Argument
}

func new_method() Method {
	var method Method
	method.arguments = make([]Argument, 0, 2)
	return method
}

func (method Method) call_string() string {
	var call_string string
	var i int = 0

	call_string += "(" + method.receiver.name + " " + method.receiver._type + ")"
	call_string += "." + method.name + "("
	for _, argument := range method.arguments {
		call_string += argument.name + " " + argument._type
		i++
		if i < len(method.arguments) {
			call_string += ", "
		}
	}
	call_string += ")"
	return call_string
}

type Callable interface {
	call_string() string
}

type TestInfo struct {
	package_name string
	lang         string
	callables    []Callable
}

func is_space(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n'
}

func is_alpha(char byte) bool {
	return (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z')
}

func is_digit(char byte) bool {
	return (char >= '0' && char <= '9')
}

func is_alphanum(char byte) bool {
	return is_alpha(char) || is_digit(char)
}

func is_valid_ident_start(char byte) bool {
	return is_alpha(char) || char == '_'
}

func is_valid_ident_char(char byte) bool {
	return is_valid_ident_start(char) || is_digit(char)
}

func skip_spaces(file_content string, parser_pos int) int {
	for ; parser_pos < len(file_content); parser_pos++ {
		if !is_space(file_content[parser_pos]) {
			return parser_pos
		}
	}
	return parser_pos
}

// TODO: Handle the case of \" in strings
// TODO: Should I do something when a string never end?
func skip_string(file_content string, parser_pos int) int {
	var string_start byte = file_content[parser_pos]

	if string_start == '"' || string_start == '`' {
		for parser_pos++; parser_pos < len(file_content); parser_pos++ {
			if file_content[parser_pos] == string_start {
				return parser_pos + 1
			}
		}
	}
	return parser_pos
}

// TODO: Handle nested comments
// TODO: Should I do something when a comment never end?
func skip_comment(file_content string, parser_pos int) int {
	comment_map := map[string]string{
		"//": "\n",
		"/*": "*/",
	}
	var comment_start []byte

	if file_content[parser_pos] == '/' {
		comment_start = append(comment_start, file_content[parser_pos])
		parser_pos++
		if file_content[parser_pos] == '/' || file_content[parser_pos] == '*' {
			comment_start = append(comment_start, file_content[parser_pos])
			parser_pos++
			parser_pos, _ = match(comment_map[string(comment_start)], file_content, parser_pos)
		}
	}
	return parser_pos
}

func skip_non_code(file_content string, parser_pos int) int {
	var new_pos int

	new_pos = skip_spaces(file_content, parser_pos)
	new_pos = skip_string(file_content, new_pos)
	new_pos = skip_comment(file_content, new_pos)

	if parser_pos != new_pos {
		return skip_non_code(file_content, new_pos)
	}
	return new_pos
}

func match(target string, file_content string, parser_pos int) (int, bool) {
	var i int = 0

	for ; parser_pos < len(file_content); parser_pos++ {
		if file_content[parser_pos] == target[i] {
			i++
			if i == len(target) {
				return parser_pos + 1, true
			}
		} else {
			i = 0
		}
	}
	return parser_pos, false
}

func match_token(target string, file_content string, parser_pos int) (int, bool) {
	var token string

	for ; parser_pos < len(file_content); parser_pos++ {
		parser_pos, token = find_next_token(file_content, parser_pos)
		if target == token {
			return parser_pos, true
		}
	}
	return parser_pos, false
}

// You should skip non-code before calling this
func find_next_token(file_content string, parser_pos int) (int, string) {
	var token []byte

	parser_pos = skip_non_code(file_content, parser_pos)
	if is_valid_ident_start(file_content[parser_pos]) {
		token = append(token, file_content[parser_pos])
		for parser_pos++; parser_pos < len(file_content); parser_pos++ {
			if !is_valid_ident_char(file_content[parser_pos]) {
				break
			}
			token = append(token, file_content[parser_pos])
		}
	} else if is_digit(file_content[parser_pos]) {
		token = append(token, file_content[parser_pos])
		for parser_pos++; parser_pos < len(file_content); parser_pos++ {
			if !is_digit(file_content[parser_pos]) {
				break
			}
			token = append(token, file_content[parser_pos])
		}
	} else {
		token = append(token, file_content[parser_pos])
		parser_pos++
	}
	return parser_pos, string(token)
}

func find_package_name(file_content string, parser_pos int) (int, string) {
	var package_name string = ""
	var new_pos, matched = match_token("package", file_content, parser_pos)

	if matched {
		new_pos, package_name = find_next_token(file_content, new_pos)
	}
	return new_pos, package_name
}

// TODO: handle pointers in arguments types
// TODO: handle variadic arguments
func find_function(file_content string, parser_pos int) (int, Callable) {
	var function Function = new_function()
	var argument Argument
	var token string

	parser_pos, function.name = find_next_token(file_content, parser_pos)
	for {
		parser_pos, token = find_next_token(file_content, parser_pos) // Skip '(' | ',' | ')'
		if token == ")" {
			break // Stop at the end of arguments
		}
		parser_pos, argument.name = find_next_token(file_content, parser_pos)
		if argument.name == ")" {
			break // Stop at the end of arguments
		}
		parser_pos, argument._type = find_next_token(file_content, parser_pos)

		function.arguments = append(function.arguments, argument)
	}
	return parser_pos, function
}

// TODO: handle pointers in arguments types
// TODO: handle variadic arguments
func find_method(file_content string, parser_pos int) (int, Method) {
	var method Method = new_method()
	var argument Argument
	var token string

	parser_pos, token = find_next_token(file_content, parser_pos) // Skip '('
	parser_pos, method.receiver.name = find_next_token(file_content, parser_pos)
	parser_pos, method.receiver._type = find_next_token(file_content, parser_pos)

	for {
		parser_pos, token = find_next_token(file_content, parser_pos) // Skip '(' | ',' | ')'
		if token == ")" {
			break // Stop at the end of arguments
		}
		parser_pos, argument.name = find_next_token(file_content, parser_pos)
		if argument.name == ")" {
			break // Stop at the end of arguments
		}
		parser_pos, argument._type = find_next_token(file_content, parser_pos)

		method.arguments = append(method.arguments, argument)
	}
	parser_pos, method.name = find_next_token(file_content, parser_pos)
	return parser_pos, method
}

// If no callable is found return a callable with zero value
func find_callable(file_content string, parser_pos int) (int, Callable) {
	var declaration_start int
	var matched bool
	var function_opener string
	var callable Callable

	declaration_start, matched = match_token("func", file_content, parser_pos)

	if matched {
		_, function_opener = find_next_token(file_content, declaration_start)
		if function_opener == "(" {
			parser_pos, callable = find_method(file_content, declaration_start)
		} else {
			parser_pos, callable = find_function(file_content, declaration_start)
		}
	}
	return declaration_start, callable
}

func get_test_info_from_go(file_content string) TestInfo {
	var parser_pos int = 0
	var test_info TestInfo
	test_info.lang = "go"

	parser_pos, test_info.package_name = find_package_name(file_content, parser_pos)

	var function Callable
	for ; parser_pos < len(file_content); parser_pos++ {
		parser_pos, function = find_callable(file_content, parser_pos)
		test_info.callables = append(test_info.callables, function)
	}
	test_info.callables = test_info.callables[0 : len(test_info.callables)-1]
	return test_info
}

// TODO: switch on file suffix to know which parser to use
func get_test_info_from(file_content string) TestInfo {
	return get_test_info_from_go(file_content)
}

func generate_test(test_info TestInfo) []byte {
	bytes := make([]byte, 5)

	return bytes
}

func main() {
	if len(os.Args) < 2 {
		panic("You need to give a file name as an argument.")
	}

	var file_name string = os.Args[1]

	content, err := ioutil.ReadFile(file_name)

	if err != nil {
		panic(err)
	}

	var test_info = get_test_info_from(string(content))

	fmt.Println("Package : " + test_info.package_name)
	fmt.Println("Lang : " + test_info.lang)
	for _, callable := range test_info.callables {
		fmt.Println(callable.call_string())
	}

	// Create temp dir?

	// if strings.HasSuffix(file_name, ".go") {
	// 	// Maybe use a temp file instead?
	// 	err = ioutil.WriteFile("crash_"+file_name,
	// 		generate_test_file_content(PACKAGE_NAME, FUNCTIONS),
	// 		'w')

	// 	if err != nil {
	// 		panic(err)
	// 	}
	// } else {
	// 	panic("You need to give a .go file name as an argument.")
	// }
}
