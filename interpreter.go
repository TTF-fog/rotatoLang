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
	Args        bool //zis eez por le loading of zi args
}

type VWheel struct {
	cursor   int
	data     []interface{}
	dir      int
	argStack []interface{}
	CMPFLAG  bool
}

type CWheel struct {
	cursor int
	data   []Instruction
	dir    int
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
	var args []interface{}
	vm.V.dir = 1
	for vm.C.cursor < len(vm.C.data) {
		inst := vm.C.data[vm.C.cursor]
		switch inst.Mnemonic {
		case "NEWV":
			if len(inst.ArgumentStr) > 0 {
				vm.V.data = append(vm.V.data, inst.ArgumentStr)
			} else {
				vm.V.data = append(vm.V.data, inst.Argument)
			}
		case "WHLDIRV":
			if inst.Argument != 1 && inst.Argument != -1 {
				vm.throwError("Invalid argument", &inst)
			}
			vm.V.dir = inst.Argument
		case "WHLDIRC":
			if inst.Argument != 1 && inst.Argument != -1 {
				vm.throwError("Invalid argument", &inst)
			}
			vm.C.dir = inst.Argument
		case "ADDARG":
			args = append(args, vm.V.data[vm.V.cursor]) //idk if we shud destroy the variable
		case "CMP":
			if inst.Args {
				var popped_args []interface{}
				popped_args, args = pop_args_and_return(1, args)
				cursor_data := vm.V.data[vm.V.cursor]
				switch cursor_data.(type) {
				case int:
					vm.V.CMPFLAG = cursor_data.(int) > popped_args[0].(int)
				case string:
					vm.V.CMPFLAG = cursor_data.(string) == popped_args[0].(string)
				}
			} else {

				cursor_data := vm.V.data[vm.V.cursor]
				switch cursor_data.(type) {
				case int:
					vm.V.CMPFLAG = cursor_data.(int) > inst.Argument
				case string:
					vm.V.CMPFLAG = cursor_data.(string) == inst.ArgumentStr
				}
			}

		case "OUT":
			if len(inst.ArgumentStr) > 0 {
				println(inst.ArgumentStr)
			} else {
				fmt.Printf("%v \n", vm.V.data[vm.V.cursor])
			}
		case "INP":
			if len(inst.ArgumentStr) > 0 {
				println(inst.ArgumentStr)
			}
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			if val, err := strconv.Atoi(input); err == nil {
				vm.V.data[vm.V.cursor] = val
			} else {
				vm.V.data[vm.V.cursor] = input
			}
		case "MOVVW":

			moveSteps := inst.Argument
			if vm.V.dir == 1 {
				vm.V.cursor = mod(vm.V.cursor+moveSteps, len(vm.V.data))
			} else {
				vm.V.cursor = mod(vm.V.cursor-moveSteps, len(vm.V.data))
			}
		case "JIZ":
			if !vm.V.CMPFLAG {
				moveSteps := inst.Argument
				if vm.C.dir == 1 {
					vm.C.cursor = mod(vm.C.cursor+moveSteps, len(vm.C.data))
				} else {
					fmt.Println(vm.C.data[vm.C.cursor].Mnemonic)
				}
				continue
			}
		case "DBGPRINTV":
			vm.printDebug()
		case "DBGPRINTC":
			vm.printDebugC()
		case "RET":
			os.Exit(0)
		}

		vm.C.cursor++
	}
}

func (vm *VM) throwError(message string, inst *Instruction) {
	fmt.Printf("%s @ Line %d, instruction %s , argument %d \n", message, vm.C.cursor, inst.Mnemonic, inst.Argument)
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
	n := len(vm.V.data)
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

	for i, item := range vm.V.data {
		angle := float64(i) * angleStep

		x := int(radiusX*math.Cos(angle) + centerX)
		y := int(radiusY*math.Sin(angle) + centerY)

		s := fmt.Sprintf("%v", item)
		if i == vm.V.cursor {
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
	fmt.Println(vm.V.data)
}

func pop_args_and_return(number int, args []interface{}) ([]interface{}, []interface{}) {
	//returns a certain number of args from the front of the list
	if len(args) < number {
		panic("not enough arguments on the stack")
	}
	popped := args[:number]
	remaining := args[number:]
	return popped, remaining
}
