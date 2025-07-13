package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type Instruction struct {
	Mnemonic    string
	Argument    int
	ArgumentF   float64
	ArgumentStr string
	Args        bool
}

// test
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
		case "DEL":
			if inst.Args {
				numericArgs, err := getNumericArgs(&args, 1)
				if err != nil {
					vm.throwError(fmt.Sprintf("%s: %v", ARITHMETIC_ERROR, err), &inst)
				}
				if len(numericArgs) == 0 {
					vm.throwError(NOT_ENOUGH_ARGS_ERROR, &inst)
				}
				delay := numericArgs[0]
				if delay < 0 {
					delay = 0
				}
				time.Sleep(time.Duration(delay) * time.Millisecond)
			} else {
				time.Sleep(time.Duration(inst.Argument) * time.Millisecond)
			}
		case "DEF":
			searchCursor := vm.C.cursor + 1
			for searchCursor < len(vm.C.data) && vm.C.data[searchCursor].Mnemonic != "RET" {
				searchCursor++
			}
			if searchCursor == len(vm.C.data) {
				vm.throwError(INCORRECT_TERMINATION_ERROR, &inst)
			}
			vm.C.cursor = searchCursor
		case "ARGVIEW":
			for _, item := range args {
				fmt.Printf("%v ", item)
			}
			println("\n")
		case "JMP":
			moveSteps := inst.Argument
			if vm.C.dir == 1 {
				vm.C.cursor = mod(vm.C.cursor+moveSteps, len(vm.C.data))
			} else {
				vm.C.cursor = mod(vm.C.cursor-moveSteps, len(vm.C.data))
			}
			//println(vm.C.data[vm.C.cursor].Mnemonic)
			continue

		case "CALL":
			funcName := inst.ArgumentStr
			if startAddr, found := functions[funcName]; found {
				vm.callStack = append(vm.callStack, vm.C.cursor+1)
				var popped_args []interface{}
				if inst.Argument > 0 {
					popped_args, args = pop_args_and_return(inst.Argument, args)
				} else if inst.Args {
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
				vm.throwError(fmt.Sprintf("%s '%s'", UNDEFINED_FUNCTION_ERROR, funcName), &inst)
			}
		case "RET":
			if len(vm.callStack) > 0 {
				// Pop the function's VWheel if it's not the last one
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
				vm.throwError(BAD_ARGUMENT_ERROR, &inst)
			}
			currentVWheel.dir = inst.Argument
		case "WHLDIRC":
			if inst.Argument != 1 && inst.Argument != -1 {
				vm.throwError(BAD_ARGUMENT_ERROR, &inst)
			}
			vm.C.dir = inst.Argument
		case "ADDARG":
			args = append(args, currentVWheel.data[currentVWheel.cursor])
		case "CMP":
			if inst.Args {
				var popped_args []interface{}
				popped_args, _ = pop_args_and_return(1, args)
				cursor_data := currentVWheel.data[currentVWheel.cursor]
				switch cursor_data.(type) {
				case int:
					currentVWheel.CMPFLAG = cursor_data.(int) > popped_args[0].(int)
				case string:
					currentVWheel.CMPFLAG = cursor_data.(string) == popped_args[0].(string)
				}
			} else {
				cursor_data := currentVWheel.data[currentVWheel.cursor]
				if len(inst.ArgumentStr) > 0 {
					switch val := cursor_data.(type) {
					case int:
						currentVWheel.CMPFLAG = strconv.Itoa(val) == inst.ArgumentStr
					case string:
						currentVWheel.CMPFLAG = val == inst.ArgumentStr
					default:
						currentVWheel.CMPFLAG = false
					}
				} else {
					currentVWheel.CMPFLAG = cursor_data.(int) > inst.Argument
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
				fmt.Println(inst.ArgumentStr)
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
				vm.throwError(EMPTY_VWHEEL_ERROR, &inst)
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
				//println(vm.C.data[vm.C.cursor].Mnemonic)
				continue
			}
		case "DBGPRINTV":
			vm.printDebug()
		case "ADD":
			if inst.Args {
				numArgs := inst.Argument
				numericArgs, err := getNumericArgs(&args, numArgs)
				if err != nil {
					vm.throwError(fmt.Sprintf("%s: %v", ARITHMETIC_ERROR, err), &inst)
				}
				result := 0
				for _, val := range numericArgs {
					result += val
				}
				if len(currentVWheel.data) == 0 {
					vm.throwError(EMPTY_VWHEEL_ERROR, &inst)
				}

				currentVWheel.data[currentVWheel.cursor] = result
			} else if inst.Argument != 0 {
				result := inst.Argument + currentVWheel.data[currentVWheel.cursor].(int)
				currentVWheel.data[currentVWheel.cursor] = result
			} else {
				if len(currentVWheel.data) == 0 {
					vm.throwError(EMPTY_VWHEEL_ERROR, &inst)
				}
				result := 0
				for _, item := range currentVWheel.data {
					if val, ok := item.(int); ok {
						result += val
					} else {
						vm.throwError(NUMERIC_DATA_ERROR, &inst)
					}
				}
				currentVWheel.data[currentVWheel.cursor] = result
			}
		case "SUB":
			if inst.Argument > 0 {
				numArgs := inst.Argument
				numericArgs, err := getNumericArgs(&args, numArgs)
				if err != nil {
					vm.throwError(fmt.Sprintf("%s: %v", ARITHMETIC_ERROR, err), &inst)
				}
				if len(numericArgs) == 0 {
					vm.throwError(NOT_ENOUGH_ARGS_ERROR, &inst)
				}
				result := numericArgs[0]
				for i := 1; i < len(numericArgs); i++ {
					result -= numericArgs[i]
				}
				if len(currentVWheel.data) == 0 {
					vm.throwError(EMPTY_VWHEEL_ERROR, &inst)
				}
				currentVWheel.data[currentVWheel.cursor] = result
			} else {
				if len(currentVWheel.data) < 1 {
					vm.throwError(NOT_ENOUGH_ARGS_ERROR, &inst)
				}
				var numericData []int
				for _, item := range currentVWheel.data {
					if val, ok := item.(int); ok {
						numericData = append(numericData, val)
					} else {
						vm.throwError(NUMERIC_DATA_ERROR, &inst)
					}
				}
				result := numericData[0]
				for i := 1; i < len(numericData); i++ {
					result -= numericData[i]
				}
				currentVWheel.data[currentVWheel.cursor] = result
			}
		case "MUL":
			if inst.Args {
				numArgs := inst.Argument
				numericArgs, err := getNumericArgs(&args, numArgs)
				if err != nil {
					vm.throwError(fmt.Sprintf("%s: %v", ARITHMETIC_ERROR, err), &inst)
				}
				if len(numericArgs) == 0 {
					vm.throwError(NOT_ENOUGH_ARGS_ERROR, &inst)
				}
				result := 1
				for _, val := range numericArgs {
					result *= val
				}
				if len(currentVWheel.data) == 0 {
					vm.throwError(EMPTY_VWHEEL_ERROR, &inst)
				}
				currentVWheel.data[currentVWheel.cursor] = result
			} else if inst.Argument > 0 {
				result := inst.Argument * currentVWheel.data[currentVWheel.cursor].(int)
				currentVWheel.data[currentVWheel.cursor] = result
			} else {
				if len(currentVWheel.data) == 0 {
					vm.throwError(EMPTY_VWHEEL_ERROR, &inst)
				}
				result := 1
				for _, item := range currentVWheel.data {
					if val, ok := item.(int); ok {
						result *= val
					} else {
						vm.throwError(NUMERIC_DATA_ERROR, &inst)
					}
				}
				currentVWheel.data[currentVWheel.cursor] = result
			}
		case "DIV":
			if inst.Argument > 0 {
				numArgs := inst.Argument
				numericArgs, err := getNumericArgs(&args, numArgs)
				if err != nil {
					vm.throwError(fmt.Sprintf("%s: %v", ARITHMETIC_ERROR, err), &inst)
				}
				if len(numericArgs) == 0 {
					vm.throwError(NOT_ENOUGH_ARGS_ERROR, &inst)
				}
				result := numericArgs[0]
				for i := 1; i < len(numericArgs); i++ {
					if numericArgs[i] == 0 {
						vm.throwError(DIVISION_BY_ZERO_ERROR, &inst)
					}
					result /= numericArgs[i]
				}
				if len(currentVWheel.data) == 0 {
					vm.throwError(EMPTY_VWHEEL_ERROR, &inst)
				}
				currentVWheel.data[currentVWheel.cursor] = result
			} else {
				if len(currentVWheel.data) < 1 {
					vm.throwError(NOT_ENOUGH_ARGS_ERROR, &inst)
				}
				var numericData []int
				for _, item := range currentVWheel.data {
					if val, ok := item.(int); ok {
						numericData = append(numericData, val)
					} else {
						vm.throwError(NUMERIC_DATA_ERROR, &inst)
					}
				}
				var result float64
				result = float64(numericData[0])
				for i := 1; i < len(numericData); i++ {
					if numericData[i] == 0 {
						vm.throwError(DIVISION_BY_ZERO_ERROR, &inst)
					}
					result /= float64(numericData[i])
				}
				currentVWheel.data[currentVWheel.cursor] = result
			}
		case "DBGPRINTC":
			vm.printDebugC()
		}

		vm.C.cursor++
	}
}

const (
	BAD_ARGUMENT_ERROR          = "Bad Argument"
	INCORRECT_TERMINATION_ERROR = "Incorrect Termination"
	EMPTY_VWHEEL_ERROR          = "Cannot move on empty VWheel"
	NUMERIC_DATA_ERROR          = "Numeric data required in VWheel"
	NOT_ENOUGH_ARGS_ERROR       = "Not enough arguments"
	DIVISION_BY_ZERO_ERROR      = "Division by zero"
	UNDEFINED_FUNCTION_ERROR    = "Call to undefined function"
	ARITHMETIC_ERROR            = "Arithmetic error"
)

func (vm *VM) throwError(message string, inst *Instruction) {
	nextInst := vm.C.data[vm.C.cursor+1]
	if nextInst.Mnemonic == "ERRH" {
		var moveSteps int
		moveSteps = nextInst.Argument
		//!! always fold this code or you will be blinded

		if len(nextInst.ArgumentStr) > 0 {
			arg := nextInst.ArgumentStr
			switch arg {
			//i apolgise for inflicting this code upon the world
			case "BAD_ARGUMENT_ERROR":
				if message == BAD_ARGUMENT_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			case "INCORRECT_TERMINATION_ERROR":
				if message == INCORRECT_TERMINATION_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			case "EMPTY_VWHEEL_ERROR":
				if message == EMPTY_VWHEEL_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			case "NUMERIC_DATA_ERROR":
				if message == NUMERIC_DATA_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			case "NOT_ENOUGH_ARGS_ERROR":
				if message == NOT_ENOUGH_ARGS_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			case "DIVISION_BY_ZERO_ERROR":
				if message == DIVISION_BY_ZERO_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			case "UNDEFINED_FUNCTION_ERROR":
				if message == UNDEFINED_FUNCTION_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			case "ARITHMETIC_ERROR":
				if message == ARITHMETIC_ERROR {
					moveSteps = nextInst.Argument
				} else {
					moveSteps = 0
				}
			default:
				panic("wtf bro :sob:")
			}
		}
		if vm.C.dir == 1 {
			vm.C.cursor = mod(vm.C.cursor+moveSteps, len(vm.C.data))
		} else {
			vm.C.cursor = mod(vm.C.cursor-moveSteps, len(vm.C.data))
		}
	} else {
		fmt.Printf("%s @ Line %d, instruction %s , argument %d", message, vm.C.cursor, inst.Mnemonic, inst.Argument)
		panic("^")
	}

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
		panic(NOT_ENOUGH_ARGS_ERROR)
	}
	popped := args[:number]
	remaining := args[number:]
	return popped, remaining
}

func getNumericArgs(args *[]interface{}, count int) ([]int, error) {
	poppedArgs, remainingArgs := pop_args_and_return(count, *args)
	*args = remainingArgs

	numericArgs := make([]int, count)
	for i, arg := range poppedArgs {
		if val, ok := arg.(int); ok {
			numericArgs[i] = val
		} else {
			return nil, fmt.Errorf("%s: %v", BAD_ARGUMENT_ERROR, arg)
		}
	}
	return numericArgs, nil
}
