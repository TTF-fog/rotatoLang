package main

import (
	"fmt"
)

type Instruction struct {
	Mnemonic string
	Argument int
}

type VWheel struct {
	cursor int
	data   []int
	dir    int
}

type CWheel struct {
	cursor int
	data   []Instruction
}

type VM struct {
	V VWheel
	C CWheel
}

func mod(a, b int) int {
	return (a%b + b) % b
}

func (vm *VM) Run() {
	vm.V.cursor = 0
	for vm.C.cursor < len(vm.C.data) {
		inst := vm.C.data[vm.C.cursor]
		switch inst.Mnemonic {
		case "NEWV":
			vm.V.data = append(vm.V.data, inst.Argument)

		case "WHLDIRV":
			fmt.Println(inst.Argument)
			if inst.Argument != 1 && inst.Argument != -1 {
				vm.throwError("Invalid argument", &inst)
			}

			vm.V.dir = inst.Argument

		case "OUT":
			println(vm.V.data[vm.V.cursor])

		case "MOVVW":

			moveSteps := inst.Argument
			if vm.V.dir == 1 {
				vm.V.cursor = mod(vm.V.cursor+moveSteps, len(vm.V.data))
			} else {
				vm.V.cursor = mod(vm.V.cursor-moveSteps, len(vm.V.data))
			}
		case "CMP":
			
		}

		vm.C.cursor++
	}
}

func (vm *VM) throwError(message string, inst *Instruction) {
	fmt.Printf("%s @ Line %d, instruction %s , argument %d \n", message, vm.C.cursor, inst.Mnemonic, inst.Argument)
	panic("^")
}
