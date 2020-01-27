/*
	ESScriptGO - A golang implementation of Extrasklep's scripting language (https://github.com/extrasklep/lang)
	Copyright (C) 2019, 2020 Rph

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
	
	Changelog: v1.0.1
		- Fixed comments
 */


package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var dbgL int
var cvars int64
var vars int64

type Sign int

const (
	SignWrite    Sign = 0
	SignIf       Sign = 1
	SignAdd      Sign = 2
	SignSubtract Sign = 3
	SignMultiply Sign = 4
	SignDivide   Sign = 5
)

type SideType int

const (
	SideSTDInput    SideType = 0
	SideSTDOutput   SideType = 1
	SideOutputRaw   SideType = 2
	SideVar         SideType = 3
	SideCVar        SideType = 4
	SideCharacter   SideType = 5
	SideNumber      SideType = 6
	SideLineNumber  SideType = 7
	SideNewLineChar SideType = 8
	SideNestedVar	SideType = 9
	SideNestedCVar 	SideType = 10
)

func isInput(sideType SideType) bool {
	if sideType == SideSTDInput ||
		sideType == SideVar ||
		sideType == SideCVar ||
		sideType == SideCharacter ||
		sideType == SideNumber ||
		sideType == SideLineNumber ||
		sideType == SideNewLineChar ||
		sideType == SideNestedVar ||
		sideType == SideNestedCVar {
		return true
	} else {
		return false
	}
}

func isOutput(sideType SideType) bool {
	if sideType == SideSTDOutput ||
		sideType == SideOutputRaw ||
		sideType == SideVar ||
		sideType == SideCVar ||
		sideType == SideNestedVar ||
		sideType == SideNestedCVar {
		return true
	} else {
		return false
	}
}

func isNumber(content string) bool {
	for i := 0; i < len(content); i++ {
		firstChar := content[i]

		if (firstChar >= 48 && firstChar <= 57) || (firstChar == 45 && i == 0) {

		} else {
			return false
		}
	}
	return true
}

func parseNumber(content string) int64 {
	isNegative := false
	var currentNumber int64
	currentNumber = 0

	for i := 0; i < len(content); i++ {
		char := content[i]
		if char == 45 && i == 0 {
			isNegative = true
		} else {
			currentNumber = currentNumber * 10
			currentNumber = currentNumber + (int64(char) - 48)
		}
	}

	if isNegative {
		currentNumber = currentNumber * -1
	}

	return currentNumber
}

func checkNestedValidity(content string) bool {
	individualChars := strings.Split(content, "")
	reachedNumber := false
	for i := 0; i < len(individualChars); i++ {
		if (individualChars[i] == "v" || individualChars[i] == "c") && !reachedNumber {

		} else {
			if isNumber(individualChars[i]) {
				reachedNumber = true
			} else {
				return false
			}
		}
	}
	return true
}

func parseSide(content string, lineNum int64, side int) (SideType, int64, bool) {
	individualChars := strings.Split(content, "")

	switch individualChars[0] {
	case "v":
		// Expect the remainder to be numbers
		if isNumber(content[1:]) {
			return SideVar, parseNumber(content[1:]), false
		} else {
			if checkNestedValidity(content[1:]) {
				return SideNestedVar, 0, true
			}
			log.Panic("Line: ", lineNum, " side: ", side, " variable requires number, got: ", content[1:])
		}
	case "c":
		if isNumber(content[1:]) {
			return SideCVar, parseNumber(content[1:]), false
		} else {
			if checkNestedValidity(content[1:]) {
				return SideNestedCVar, 0, true
			}
			log.Panic("Line: ", lineNum, " side: ", side, " character variable requires number, got: ", content[1:])
		}
	case "i":
		if len(content) == 1 {
			return SideSTDInput, 0, false
		} else {
			log.Panic("Line: ", lineNum, " side: ", side, " input opcode must be exactly 1 character!")
		}
	case "o":
		if len(content) == 1 {
			return SideSTDOutput, 0, false
		} else {
			log.Panic("Line: ", lineNum, " side: ", side, " output opcode must be exactly 1 character!")
		}
	case "r":
		if len(content) == 1 {
			return SideOutputRaw, 0, false
		} else {
			log.Panic("Line: ", lineNum, " side: ", side, " raw output opcode must be exactly 1 character!")
		}
	case "l":
		if len(content) == 1 {
			return SideLineNumber, 0, false
		} else {
			log.Panic("Line: ", lineNum, " side: ", side, " line number opcode must be exactly 1 character!")
		}
	case "n":
		if len(content) == 1 {
			return SideNewLineChar, 0, false
		} else {
			log.Panic("Line: ", lineNum, " side: ", side, " new line char opcode must be exactly 1 character!")
		}
	case "\\":
		if len(content) == 2 {
			return SideCharacter, int64(content[1]), false
		} else {
			log.Panic("Line: ", lineNum, " side: ", side, " char opcode must be exactly 2 characters!")
		}
	default:
		// we're assuming its just a number
		if isNumber(content) {
			return SideNumber, parseNumber(content), false
		} else {
			log.Panic("Line: ", lineNum, " side: ", side, " no opcode, number expected, didn't get number")
		}
	}
	return SideNumber, 0, false
}

type Line struct {
	left      	SideType
	leftParam 	int64

	sign 		Sign

	right      	SideType
	rightParam 	int64

	hasCode 	bool

	lineNum 	int64

	leftNest   	string
	rightNest	string
}

var parsedLines []Line

var memVar []int64
var memCVar []byte

func dbgLog(s string) {
	if dbgL > 1 {
		fmt.Println(s)
	}
}

func resolveNumberFromNest(nest string) int64 {
	var lastNumber int64
	lastNumber = 0
	// Locate where number starts
	tokens := strings.Split(nest, "")
	otherTokens := tokens


	for i := 0; i < len(tokens); i++ {
		if tokens[i] != "c" && tokens[i] != "v" {
			tokens = tokens[0:i]
			otherTokens = otherTokens[i:]
			lastNumber = parseNumber(strings.Join(otherTokens, ""))
			break
		}
	}

	for i := len(tokens) - 1; i >= 0; i-- {
		t := tokens[i]
		if t == "c" {
			if lastNumber < 0 || lastNumber > cvars - 1 {
				log.Panic("NESTED VALUE PARSER: UNCLAMPED ACCESS!")
			}
			lastNumber = int64(memCVar[lastNumber])
		}
		if t == "v" {
			if lastNumber < 0 || lastNumber > vars - 1 {
				log.Panic("NESTED VALUE PARSER: UNCLAMPED ACCESS!")
			}
			lastNumber = memVar[lastNumber]
		}
	}

	return lastNumber
}

func readInput(side SideType, param int64, nest string, lineNum int64) int64 {
	switch side {
	case SideSTDInput:
		fmt.Print("< ")
		var inp int64
		_, _ = fmt.Scan(&inp)
		return inp
	case SideVar:
		if param < 0 || param > vars {
			log.Panic(lineNum, " unclamped variable access")
		}
		return memVar[param]
	case SideCVar:
		if param < 0 || param > cvars {
			log.Panic(lineNum, " unclamped variable access")
		}
		return int64(memCVar[param])
	case SideCharacter:
		return param
	case SideNumber:
		return param
	case SideLineNumber:
		return lineNum
	case SideNewLineChar:
		return int64("\n"[0])
	case SideNestedVar:
		index := resolveNumberFromNest(nest)
		return memVar[index]
	case SideNestedCVar:
		index := resolveNumberFromNest(nest)
		return int64(memCVar[index])
	}
	return 0
}

func writeOutput(side SideType, param int64, nest string, value int64, lineNum int64) {
	switch side {
	case SideSTDOutput:
		// Number to stdout
		fmt.Println(">", value)
	case SideOutputRaw:
		// byte to stdout
		var buff [1]byte
		buff[0] = byte(value)
		_, _ = os.Stdout.Write(buff[:])
	case SideVar:
		if param < 0 || param > vars {
			log.Panic(lineNum, " unclamped variable access")
		}
		memVar[param] = value
	case SideCVar:
		if param < 0 || param > cvars {
			log.Panic(lineNum, " unclamped variable access")
		}
		memCVar[param] = byte(value)
	case SideNestedVar:
		index := resolveNumberFromNest(nest)
		memVar[index] = value
	case SideNestedCVar:
		index := resolveNumberFromNest(nest)
		memCVar[index] = byte(value)
	}
}

func execute(parsedLines []Line) {
	var currentLine int64
	currentLine = 0

	dbgLog("Beginning execution")

	for {
		if currentLine >= int64(len(parsedLines)) || currentLine < 1{
			dbgLog("exit: Ran out of lines or line underflow")
			break
		}
		normalJump := true
		workingLine := parsedLines[currentLine]

		if workingLine.hasCode {
			switch workingLine.sign {
			case SignWrite:
				a := readInput(workingLine.left, workingLine.leftParam, workingLine.leftNest, workingLine.lineNum)
				writeOutput(workingLine.right, workingLine.rightParam, workingLine.rightNest, a, workingLine.lineNum)
			case SignIf:
				a := readInput(workingLine.left, workingLine.leftParam, workingLine.leftNest, workingLine.lineNum)
				b := readInput(workingLine.right, workingLine.rightParam, workingLine.rightNest, workingLine.lineNum)
				if a > 0 {
					currentLine = b - 1
					normalJump = false
				}
			case SignAdd:
				a := readInput(workingLine.left, workingLine.leftParam, workingLine.leftNest, workingLine.lineNum)
				b := readInput(workingLine.right, workingLine.rightParam, workingLine.rightNest, workingLine.lineNum)
				writeOutput(workingLine.right, workingLine.rightParam, workingLine.rightNest, a+b, workingLine.lineNum)
			case SignSubtract:
				a := readInput(workingLine.left, workingLine.leftParam, workingLine.leftNest, workingLine.lineNum)
				b := readInput(workingLine.right, workingLine.rightParam, workingLine.rightNest, workingLine.lineNum)
				writeOutput(workingLine.right, workingLine.rightParam, workingLine.rightNest, b-a, workingLine.lineNum)
			case SignMultiply:
				a := readInput(workingLine.left, workingLine.leftParam, workingLine.leftNest, workingLine.lineNum)
				b := readInput(workingLine.right, workingLine.rightParam, workingLine.rightNest, workingLine.lineNum)
				writeOutput(workingLine.right, workingLine.rightParam, workingLine.rightNest, b*a, workingLine.lineNum)
			case SignDivide:
				a := readInput(workingLine.left, workingLine.leftParam, workingLine.leftNest, workingLine.lineNum)
				b := readInput(workingLine.right, workingLine.rightParam, workingLine.rightNest, workingLine.lineNum)
				writeOutput(workingLine.right, workingLine.rightParam, workingLine.rightNest, b/a, workingLine.lineNum)
			}
		}

		if normalJump {
			currentLine++
		}
	}
}

func main() {
	flag.IntVar(&dbgL, "debug", 0, "Debug level of the interpreter")
	flag.Int64Var(&cvars, "cvar", 32768, "Character variable amount")
	flag.Int64Var(&vars, "var", 256, "Variable amount")

	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("No script supplied!")
		fmt.Println("Usage: interpreter [--debug=debugLevel] [--cvar=CVarAmount] [--var=VarAmount] <script>")
		return
	}

	dbgLog("Allocating memory...")

	memVar = make([]int64, vars)
	memCVar = make([]byte, cvars)

	if flag.Arg(0) == "testMode" {
		//writeOutput(SideOutputRaw, 0, "", 13, 0)
		memVar[100] = 50
		memVar[50] = 25
		memVar[25] = 40
		fmt.Println(readInput(SideVar, 50, "", 1))

		return
	}

	dbgLog("Reading file")

	file, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	content := string(buf)

	lines := strings.Split(content, "\n")

	if dbgL > 2 {
		fmt.Println(lines)
	}

	if dbgL > 1 {
		fmt.Println("Line amount: ", len(lines))
	}

	dbgLog("Parsing lines...")

	parsedLines = make([]Line, len(lines))

	for index, element := range lines {
		if dbgL > 2 {
			fmt.Println(index, " ", element)
		}

		thisLine := Line{}
		thisLine.hasCode = false
		thisLine.lineNum = int64(index)

		// Step 1: Split into character slice
		chars := strings.Split(element, "")

		// Step 2: Step over all characters
		previousChar := ""
		currentSide := 0

		leftSide := ""
		command := ""
		rightSide := ""

		for _, character := range chars {
			// Start of line comments
			if character == "/" && previousChar == "/" {
				dbgLog("comment, break")
				if currentSide != 2 {
					thisLine.hasCode = false
					dbgLog("premature comment")
				}
				break
			}

			if currentSide == 2 {
				if character != ";" {
					rightSide = rightSide + character
				} else {
					thisLine.hasCode = true
					break
				}
			}

			if currentSide == 1 {
				command = character
				currentSide = 2
			}

			// Parse left side
			if currentSide == 0 {
				// Beginning of command
				if character == ">" && previousChar != "\\" {
					currentSide = 1
				} else {
					leftSide = leftSide + character
				}
			}

			previousChar = character
		}

		// Step 3: Detect command type
		if thisLine.hasCode {
			switch command {
			case ">":
				thisLine.sign = SignWrite
			case "?":
				thisLine.sign = SignIf
			case "+":
				thisLine.sign = SignAdd
			case "-":
				thisLine.sign = SignSubtract
			case "*":
				thisLine.sign = SignMultiply
			case "/":
				thisLine.sign = SignDivide
			default:
				log.Panic("Invalid command! (line: ", thisLine.lineNum + 1, ", command: ", command, " )")
			}
		}

		if thisLine.hasCode {
			// Step 4: Parse left side
			lside, lmeta, isNested := parseSide(leftSide, thisLine.lineNum, 0)
			if !isInput(lside) {
				log.Panic("Line: ", thisLine.lineNum + 1, " left side, input expected, got output only datatype.")
			}

			thisLine.left = lside
			thisLine.leftParam = lmeta

			if (isNested) {
				thisLine.leftNest = leftSide[1:]
			}

			// Step 5: Parse right side
			rside, rmeta, isNested := parseSide(rightSide, thisLine.lineNum, 0)
			if thisLine.sign == SignIf {
				if !isInput(rside) {
					log.Panic("Line: ", thisLine.lineNum + 1, " right side, input expected, got output only datatype.")
				}
			} else {
				if thisLine.sign == SignAdd || thisLine.sign == SignSubtract || thisLine.sign == SignMultiply || thisLine.sign == SignDivide {
					if isOutput(rside) && isInput(rside) {

					} else {
						log.Panic("Line: ", thisLine.lineNum + 1, " right side, input AND output expected, got only one.")
					}
				} else {
					if !isOutput(rside) {
						log.Panic("Line: ", thisLine.lineNum + 1, " right side, output expected, got input only datatype.")
					}
				}
			}

			thisLine.right = rside
			thisLine.rightParam = rmeta

			if isNested {
				thisLine.rightNest = rightSide[1:]
			}
		}
		parsedLines[index] = thisLine
		parsedLines[index].lineNum = parsedLines[index].lineNum + 1
	}

	execute(parsedLines)
}
