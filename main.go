package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	file, err := os.Open("test.whl")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	input := strings.Join(lines, "\n")
	lexer := NewLexer(input)

	var instructions []Instruction
	for {
		tok := lexer.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == INST {
			argTok := lexer.NextToken()
			arg := 0
			if argTok.Type == INTEGER {
				arg, _ = strconv.Atoi(argTok.Literal)
			}
			instructions = append(instructions, Instruction{Mnemonic: tok.Literal, Argument: arg})
		}
	}

	vm := &VM{
		C: CWheel{
			data: instructions,
		},
	}

	vm.Run()

	fmt.Printf("Variable Wheel after NEWV 0: %+v", vm.V)
}
