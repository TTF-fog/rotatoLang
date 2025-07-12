//go:build js

package main

import (
	"fmt"
	"strconv"
	"syscall/js"
)

func runTwistCode(this js.Value, args []js.Value) interface{} {
	if len(args) == 0 {
		fmt.Println("No code provided")
		return nil
	}
	input := args[0].String()
	lexer := NewLexer(input)

	var instructions []Instruction

	for {
		tok := lexer.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == INST {
			arg := 0
			args := false
			argF := 0.0
			str_arg := ""
			argTok := lexer.NextToken()
			for argTok.Type != NEWLINE {
				if argTok.Type == INTEGER {
					arg, _ = strconv.Atoi(argTok.Literal.(string))
				} else if argTok.Type == FLOAT {
					argF, _ = strconv.ParseFloat(argTok.Literal.(string), 64)
				} else if argTok.Type == STRING {
					str_arg = argTok.Literal.(string)
				} else if argTok.Type == ARGS {
					args = true
				}
				argTok = lexer.NextToken()
			}
			instructions = append(instructions, Instruction{Mnemonic: tok.Literal.(string), Argument: arg, ArgumentF: argF, ArgumentStr: str_arg, Args: args})
		}
	}

	vm := &VM{
		C: CWheel{
			data: instructions,
		},
		// Initialize the VM with a global scope (one VWheel on the dataStack).
		dataStack: []VWheel{{dir: 1}},
	}

	vm.Run()
	return nil
}

func main() {
	println("Twist Wasm Initialized")
	js.Global().Set("runTwistCode", js.FuncOf(runTwistCode))
	<-make(chan bool)
}
