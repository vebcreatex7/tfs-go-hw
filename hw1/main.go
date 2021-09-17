package main

import (
	"fmt"
)

type TFunc func(*map[string]int)

func size(s int) TFunc {
	return func(args *map[string]int) {
		(*args)["size"] = s
	}
}
func char(c int) TFunc {
	return func(args *map[string]int) {
		(*args)["char"] = c
	}
}
func color(c int) TFunc {
	return func(args *map[string]int) {
		(*args)["color"] = c
	}
}

func sandglass(params ...TFunc) {
	// default values
	args := map[string]int{
		"size":  10,
		"char":  'X',
		"color": 0,
	}

	for _, arg := range params {
		arg(&args)
	}

	var size int
	var char rune
	var color int

	size = args["size"]
	char = rune(args["char"])
	color = args["color"]

	// checking the conditions
	if !(0 <= char && char <= 127) {
		fmt.Print("Error. Wrong symbol. Expected ASCII.\n")
		return
	}

	if color != 0 && !(30 <= color && color <= 37) {
		fmt.Print("Error. Wrong color.\n")
		return
	}

	// Drawing

	// The upper cover
	for i := 0; i < size; i++ {
		fmt.Printf("\033[%dm%c\033[0m", color, char)
	}
	fmt.Printf("\n")

	// The upper part
	hight := size / 2
	for i := 1; i < hight; i++ {
		for j := 0; j < size; j++ {
			switch {
			case j == i:
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			case j == size-i-1:
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			default:
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}

	// The middle
	if size%2 == 1 {
		for i := 0; i < size/2; i++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\033[%dm%c\033[0m", color, char)
		for i := size/2 + 1; i < size; i++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
	}

	// The lower part
	for i := hight + 1; i < size-1; i++ {
		for j := 0; j < size; j++ {
			switch {
			case j == i:
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			case j == size-i-1:
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			default:
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}

	// The lower cover
	for i := 0; i < size; i++ {
		fmt.Printf("\033[%dm%c\033[0m", color, char)
	}
	fmt.Printf("\n")
}

func main() {
	sandglass(color(36), size(15), char('#'))
}
