package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type Instruction struct {
	Mnemonic    string
	Argument    int
	ArgumentStr string
	Args        bool
}

type VWheel struct {
	cursor  int
	data    []interface{}
	dir     int
	CMPFLAG bool
}

type CWheel struct {
	cursor int
	data   []Instruction
	dir    int
}
type VM struct {
	dataStack []VWheel
	C         CWheel
	callStack []int
}

func mod(a, b int) int {
	return (a%b + b) % b
}

type function struct {
	argument_count int
	line           int
}

func (vm *VM) Run() {
	functions := make(map[string]function)
	var args []interface{}

	for i, inst := range vm.C.data {
		if inst.Mnemonic == "DEF" {
			functions[inst.ArgumentStr] = function{
				line:           i + 1,
				argument_count: inst.Argument, //for now...
			}
		}
	}

	if len(vm.dataStack) > 0 {
		vm.dataStack[0].dir = 1
	}

	for vm.C.cursor < len(vm.C.data) {
		inst := vm.C.data[vm.C.cursor]
		currentVWheel := &vm.dataStack[len(vm.dataStack)-1]

		switch inst.Mnemonic {
		case "DEF":
			searchCursor := vm.C.cursor + 1
			for searchCursor < len(vm.C.data) && vm.C.data[searchCursor].Mnemonic != "RET" {
				searchCursor++
			}
			if searchCursor == len(vm.C.data) {
				vm.throwError("incorrect termination", &inst)
			}
			vm.C.cursor = searchCursor

		case "CALL":
			funcName := inst.ArgumentStr
			if startAddr, found := functions[funcName]; found {
				vm.callStack = append(vm.callStack, vm.C.cursor)
				var popped_args []interface{}
				if inst.Args {
					popped_args, args = pop_args_and_return(functions[funcName].argument_count, args)
				}
				newVWheel := VWheel{
					dir:  1,
					data: popped_args,
				}

				vm.dataStack = append(vm.dataStack, newVWheel)
				vm.C.cursor = startAddr.line
				continue
			} else {
				vm.throwError(fmt.Sprintf("Call to undefined function '%s'", funcName), &inst)
			}
		case "RET":
			if len(vm.callStack) > 0 {
				if len(vm.dataStack) > 1 {
					vm.dataStack = vm.dataStack[:len(vm.dataStack)-1]
				}
				returnAddr := vm.callStack[len(vm.callStack)-1]
				vm.callStack = vm.callStack[:len(vm.callStack)-1]
				vm.C.cursor = returnAddr
			} else {
				os.Exit(0)
			}
		case "NEWV":
			if len(inst.ArgumentStr) > 0 {
				currentVWheel.data = append(currentVWheel.data, inst.ArgumentStr)
			} else {
				currentVWheel.data = append(currentVWheel.data, inst.Argument)
			}
		case "WHLDIRV":
			if inst.Argument != 1 && inst.Argument != -1 {
				vm.throwError("Invalid argument", &inst)
			}
			currentVWheel.dir = inst.Argument
		case "WHLDIRC":
			if inst.Argument != 1 && inst.Argument != -1 {
				vm.throwError("Invalid argument", &inst)
			}
			vm.C.dir = inst.Argument
		case "ADDARG":
			args = append(args, currentVWheel.data[currentVWheel.cursor])
		case "CMP":
			if inst.Args {
				var popped_args []interface{}
				popped_args, args = pop_args_and_return(1, args)
				cursor_data := currentVWheel.data[currentVWheel.cursor]
				switch cursor_data.(type) {
				case int:
					currentVWheel.CMPFLAG = cursor_data.(int) > popped_args[0].(int)
				case string:
					currentVWheel.CMPFLAG = cursor_data.(string) == popped_args[0].(string)

				}
			} else {
				cursor_data := currentVWheel.data[currentVWheel.cursor]
				switch cursor_data.(type) {
				case int:
					currentVWheel.CMPFLAG = cursor_data.(int) > inst.Argument
				case string:
					currentVWheel.CMPFLAG = cursor_data.(string) == inst.ArgumentStr
				}
			}
		case "OUT":
			if len(inst.ArgumentStr) > 0 {
				println(inst.ArgumentStr)
			} else {
				fmt.Printf("%v \n", currentVWheel.data[currentVWheel.cursor])
			}
		case "INP":
			if len(inst.ArgumentStr) > 0 {
				println(inst.ArgumentStr)
			}
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if val, err := strconv.Atoi(input); err == nil {
				currentVWheel.data[currentVWheel.cursor] = val
			} else {
				currentVWheel.data[currentVWheel.cursor] = input
			}
		case "MOVVW":
			moveSteps := inst.Argument
			if len(currentVWheel.data) == 0 {
				vm.throwError("Cannot move on empty VWheel", &inst)
			}
			if currentVWheel.dir == 1 {
				currentVWheel.cursor = mod(currentVWheel.cursor+moveSteps, len(currentVWheel.data))
			} else {
				currentVWheel.cursor = mod(currentVWheel.cursor-moveSteps, len(currentVWheel.data))
			}
		case "JIZ":
			if !currentVWheel.CMPFLAG {
				moveSteps := inst.Argument
				if vm.C.dir == 1 {
					vm.C.cursor = mod(vm.C.cursor+moveSteps, len(vm.C.data))
				} else {
					vm.C.cursor = mod(vm.C.cursor-moveSteps, len(vm.C.data))
				}
				continue
			}
		case "DBGPRINTV":
			vm.printDebug()
		case "DBGPRINTC":
			vm.printDebugC()
		}

		vm.C.cursor++
	}
}

func (vm *VM) throwError(message string, inst *Instruction) {
	fmt.Printf("%s @ Line %d, instruction %s , argument %d", message, vm.C.cursor, inst.Mnemonic, inst.Argument)
	panic("^")
}

func (vm *VM) printDebugC() {
	n := len(vm.C.data)
	if n == 0 {
		fmt.Println("no instructions")
		return
	}
	radiusY := float64(n) * 0.8
	if radiusY < 4 {
		radiusY = 4
	}
	radiusX := radiusY * 2.0

	height := int(2*radiusY) + 1
	width := int(2*radiusX) + 1

	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	centerX := radiusX
	centerY := radiusY

	angleStep := 2 * math.Pi / float64(n)
	shades := []rune{'░', '▓'}

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			x := float64(j) - centerX
			y := float64(i) - centerY

			if (x*x)/(radiusX*radiusX)+(y*y)/(radiusY*radiusY) <= 1 {
				angle := math.Atan2(y, x)
				if angle < 0 {
					angle += 2 * math.Pi
				}

				shiftedAngle := angle + angleStep/2
				if shiftedAngle >= 2*math.Pi {
					shiftedAngle -= 2 * math.Pi
				}

				sector := int(math.Floor(shiftedAngle / angleStep))
				canvas[i][j] = shades[sector%2]
			}
		}
	}

	for i, item := range vm.C.data {
		angle := float64(i) * angleStep

		x := int(radiusX*math.Cos(angle) + centerX)
		y := int(radiusY*math.Sin(angle) + centerY)

		s := fmt.Sprintf("%s %d %s", item.Mnemonic, item.Argument, item.ArgumentStr)
		if i == vm.C.cursor {
			s = fmt.Sprintf("[%s]", s)
		}

		strLen := len(s)
		startPos := x - strLen/2
		for k, char := range s {
			if y >= 0 && y < height && startPos+k >= 0 && startPos+k < width {
				canvas[y][startPos+k] = char
			}
		}
	}

	for _, row := range canvas {
		fmt.Println(string(row))
	}
}

func (vm *VM) printDebug() {
	n := len(vm.dataStack[len(vm.dataStack)-1].data)
	if n == 0 {
		fmt.Println("no variables")
		return
	}
	radiusY := float64(n) * 0.8
	if radiusY < 4 {
		radiusY = 4
	}
	radiusX := radiusY * 2.0

	height := int(2*radiusY) + 1
	width := int(2*radiusX) + 1

	canvas := make([][]rune, height)
	for i := range canvas {
		canvas[i] = make([]rune, width)
		for j := range canvas[i] {
			canvas[i][j] = ' '
		}
	}

	centerX := radiusX
	centerY := radiusY

	angleStep := 2 * math.Pi / float64(n)
	shades := []rune{'░', '▓'}

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			x := float64(j) - centerX
			y := float64(i) - centerY

			if (x*x)/(radiusX*radiusX)+(y*y)/(radiusY*radiusY) <= 1 {
				angle := math.Atan2(y, x)
				if angle < 0 {
					angle += 2 * math.Pi
				}

				shiftedAngle := angle + angleStep/2
				if shiftedAngle >= 2*math.Pi {
					shiftedAngle -= 2 * math.Pi
				}

				sector := int(math.Floor(shiftedAngle / angleStep))
				canvas[i][j] = shades[sector%2]
			}
		}
	}

	for i, item := range vm.dataStack[len(vm.dataStack)-1].data {
		angle := float64(i) * angleStep

		x := int(radiusX*math.Cos(angle) + centerX)
		y := int(radiusY*math.Sin(angle) + centerY)

		s := fmt.Sprintf("%v", item)
		if i == vm.dataStack[len(vm.dataStack)-1].cursor {
			s = fmt.Sprintf("[%v]", item)
		}

		strLen := len(s)
		startPos := x - strLen/2
		for k, char := range s {
			if y >= 0 && y < height && startPos+k >= 0 && startPos+k < width {
				canvas[y][startPos+k] = char
			}
		}
	}

	for _, row := range canvas {
		fmt.Println(string(row))
	}
	fmt.Println(vm.dataStack[len(vm.dataStack)-1].data)
}

func pop_args_and_return(number int, args []interface{}) ([]interface{}, []interface{}) {
	if len(args) < number {
		panic("not enough arguments on the stack")
	}
	popped := args[:number]
	remaining := args[number:]
	return popped, remaining
}
