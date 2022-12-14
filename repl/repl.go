package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/carepollo/esolang/compiler"
	"github.com/carepollo/esolang/lexer"
	"github.com/carepollo/esolang/object"
	"github.com/carepollo/esolang/parser"
	"github.com/carepollo/esolang/vm"
)

const PROMPT = "--> "

// start the interpreter in interactive mode (read, eval, print, loop), basically console mode
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	for {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		lexer := lexer.New(line)
		parser := parser.New(lexer)
		program := parser.ParseProgram()

		if len(parser.Errors()) != 0 {
			printParserErrors(out, parser.Errors())
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation failed:\n %s\n", err)
			continue
		}

		code := comp.Bytecode()
		constants = code.Constants
		machine := vm.NewWithGlobalsStore(code, globals)

		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElement()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

// helper to show internal record of errors
func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, "Error raised")
	for _, message := range errors {
		io.WriteString(out, "\t"+message+"\n")
	}
}
