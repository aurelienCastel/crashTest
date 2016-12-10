// TODO: Find a way to not try to access an index that is outside the file_content
// without adding a lot of check
// TODO: doesn't work on anonymous functions
// TODO: doesn't work on functions not declared in the outer scope of the file.
// TODO: for now the program assume you don't have syntax errors
// should I add error handling of syntax error in the parsed code?

package main

import "fmt"
import "io/ioutil"
import "os"

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
		var name string = argument.name
		if argument.name == "" {
			name = "-UNNAMED-"
		}
		call_string += name + " " + argument._type
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
		var name string = argument.name
		if argument.name == "" {
			name = "-UNNAMED-"
		}
		call_string += name + " " + argument._type
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

func is_valid_identifier(token string) bool {
	if len(token) == 0 || !is_valid_ident_start(token[0]) {
		return false
	}
	for i := 1; i < len(token); i++ {
		if !is_valid_ident_char(token[i]) {
			return false
		}
	}
	return true
}

func skip_spaces(file_content string, parser_pos int) int {
	for ; parser_pos < len(file_content); parser_pos++ {
		if !is_space(file_content[parser_pos]) {
			return parser_pos
		}
	}
	return parser_pos
}

func skip_string(file_content string, parser_pos int) int {
	if parser_pos >= len(file_content) {
		return parser_pos
	}

	var string_start byte = file_content[parser_pos]

	if string_start == '"' {
		parser_pos++
		for ; parser_pos < len(file_content); parser_pos++ {
			// Handle the case of \ in interpreted strings
			if file_content[parser_pos] == '\\' {
				parser_pos++
				continue
			}
			if file_content[parser_pos] == string_start {
				return parser_pos + 1
			}
		}
	} else if string_start == '`' {
		parser_pos++
		for ; parser_pos < len(file_content); parser_pos++ {
			if file_content[parser_pos] == string_start {
				return parser_pos + 1
			}
		}
	}
	return parser_pos
}

func skip_comment(file_content string, parser_pos int) int {
	if parser_pos >= len(file_content) {
		return parser_pos
	}

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

// Move to the character after the target and return true,
// or move to the end of the file and return false
// (doesn't skip 'non code')
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

// Move to the character after the target and return true,
// or move to the end of the file and return false
// (look for target only in code, not in strings or comments)
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

// You should call this BEFORE the parser_pos is on an opening token.
func match_closing(opening string, closing string, file_content string, parser_pos int) (int, bool) {
	var i int = 0
	var j int = 0
	var nesting_level int = 0

	for ; parser_pos < len(file_content); parser_pos++ {
		if file_content[parser_pos] == opening[i] {
			i++
			if i == len(opening) {
				nesting_level++
				i = 0
				j = 0
			}
		} else {
			i = 0
		}
		if file_content[parser_pos] == closing[j] {
			j++
			if j == len(closing) {
				nesting_level--
				if nesting_level == 0 {
					return parser_pos + 1, true
				}
				i = 0
				j = 0
			}
		} else {
			j = 0
		}
	}
	return parser_pos, false
}

func find_next_token(file_content string, parser_pos int) (int, string) {
	var token []byte

	parser_pos = skip_non_code(file_content, parser_pos)

	if parser_pos >= len(file_content) {
		return parser_pos, ""
	}

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

func find_next_type_recursive(file_content string, parser_pos int) int {
	var token string
	parser_pos, token = find_next_token(file_content, parser_pos)
	switch token {
	case ")", ",":
		return parser_pos - 1
	case "[":
		{
			parser_pos, _ = match_closing(token, "]", file_content, parser_pos-1)
			parser_pos = find_next_type_recursive(file_content, parser_pos)
		}
	case "(":
		{
			parser_pos, _ = match_closing(token, ")", file_content, parser_pos-1)
			parser_pos = find_next_type_recursive(file_content, parser_pos)
		}
	default:
		parser_pos = find_next_type_recursive(file_content, parser_pos)
	}
	return parser_pos
}

func find_next_type(file_content string, parser_pos int) (int, string) {
	parser_pos = skip_spaces(file_content, parser_pos)
	end_pos := find_next_type_recursive(file_content, parser_pos)
	return end_pos, file_content[parser_pos:end_pos]
}

func find_package_name(file_content string, parser_pos int) (int, string) {
	var package_name string = ""
	var new_pos, matched = match_token("package", file_content, parser_pos)

	if matched {
		new_pos, package_name = find_next_token(file_content, new_pos)
	}
	return new_pos, package_name
}

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
		// If argument.name is not an identifier its a type,
		// the argument is unnamed
		if !is_valid_identifier(argument.name) {
			parser_pos -= len(argument.name)
			argument.name = ""
		}
		parser_pos, argument._type = find_next_type(file_content, parser_pos)
		// if argument._type is "" it means the identifier in
		// argument.name is in fact a type and
		// the argument is unnamed
		if argument._type == "" {
			argument._type = argument.name
			argument.name = ""
		}

		function.arguments = append(function.arguments, argument)
	}
	return parser_pos, function
}

func find_method(file_content string, parser_pos int) (int, Method) {
	var method Method = new_method()
	var argument Argument
	var token string

	parser_pos, token = find_next_token(file_content, parser_pos) // Skip '('
	parser_pos, method.receiver.name = find_next_token(file_content, parser_pos)
	parser_pos, method.receiver._type = find_next_type(file_content, parser_pos)

	for {
		parser_pos, token = find_next_token(file_content, parser_pos) // Skip '(' | ',' | ')'
		if token == ")" {
			break // Stop at the end of arguments
		}
		parser_pos, argument.name = find_next_token(file_content, parser_pos)
		if argument.name == ")" {
			break // Stop at the end of arguments
		}
		// If argument.name is not an identifier its a type,
		// the argument is unnamed
		if !is_valid_identifier(argument.name) {
			parser_pos -= len(argument.name)
			argument.name = ""
		}
		parser_pos, argument._type = find_next_type(file_content, parser_pos)
		// if argument._type is "" it means the identifier in
		// argument.name is in fact a type and
		// the argument is unnamed
		if argument._type == "" {
			argument._type = argument.name
		}

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
		if file_content[declaration_start-5] == '\n' {
			_, function_opener = find_next_token(file_content, declaration_start)
			if function_opener == "(" {
				parser_pos, callable = find_method(file_content, declaration_start)
			} else {
				parser_pos, callable = find_function(file_content, declaration_start)
			}
			return parser_pos, callable
		} else {
			return find_callable(file_content, declaration_start)
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
}
