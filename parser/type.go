package parser

type Argument struct {
	Name string
	Type string
}

type Function struct {
	Name      string
	Arguments []Argument
}

func NewFunction() Function {
	var function Function
	function.Arguments = make([]Argument, 0, 2)
	return function
}

func (function Function) CallString() string {
	var callString string
	var i int = 0

	callString += function.Name + "("
	for _, argument := range function.Arguments {
		var name string = argument.Name
		if argument.Name == "" {
			name = "-UNNAMED-"
		}
		callString += name + " " + argument.Type
		i++
		if i < len(function.Arguments) {
			callString += ", "
		}
	}
	callString += ")"
	return callString
}

type Method struct {
	Name      string
	Receiver  Argument
	Arguments []Argument
}

func NewMethod() Method {
	var method Method
	method.Arguments = make([]Argument, 0, 2)
	return method
}

func (method Method) CallString() string {
	var callString string
	var i int = 0

	callString += "(" + method.Receiver.Name + " " + method.Receiver.Type + ")"
	callString += "." + method.Name + "("
	for _, argument := range method.Arguments {
		var name string = argument.Name
		if argument.Name == "" {
			name = "-UNNAMED-"
		}
		callString += name + " " + argument.Type
		i++
		if i < len(method.Arguments) {
			callString += ", "
		}
	}
	callString += ")"
	return callString
}

type Callable interface {
	CallString() string
}

type FileInfo struct {
	FileName    string
	PackageName string
	Language    string
	Callables   []Callable
}
