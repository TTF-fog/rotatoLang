package main

import (
	"fmt"
)

type Instruction struct {
	Mnemonic    string
	Argument    int
	ArgumentStr string
}

type VWheel struct {
	cursor  int
	data    []int
	dir     int
	CMPFLAG bool
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
			if inst.Argument != 1 && inst.Argument != -1 {
				vm.throwError("Invalid argument", &inst)
			}
			vm.V.dir = inst.Argument
		case "CMP":
			cmp := vm.V.data[vm.V.cursor] > inst.Argument
			vm.V.CMPFLAG = cmp
			println(cmp)
		case "OUT":
			if len(inst.ArgumentStr) > 0 {
				println(inst.ArgumentStr)

			} else {
				println(vm.V.data[vm.V.cursor])
			}

		case "MOVVW":

			moveSteps := inst.Argument
			if vm.V.dir == 1 {
				vm.V.cursor = mod(vm.V.cursor+moveSteps, len(vm.V.data))
			} else {
				vm.V.cursor = mod(vm.V.cursor-moveSteps, len(vm.V.data))
			}
		case "DBGPRINTV":
			fmt.Println("Variable Wheel")
			for index, item := range vm.V.data {
				if index == vm.V.cursor {
					fmt.Println(item, "*")
					continue
				} else {
					fmt.Println(item)
				}

			}
		}

		vm.C.cursor++
	}
}

func (vm *VM) throwError(message string, inst *Instruction) {
	fmt.Printf("%s @ Line %d, instruction %s , argument %d \n", message, vm.C.cursor, inst.Mnemonic, inst.Argument)
	panic("^")
}
