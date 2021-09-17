package main

import "fmt"

type TFunc func(*[]int)

func size(s int) TFunc {
	return func(slice *[]int) {
		*slice = append(*slice, s)
	}
}

func char(c int) TFunc {
	return func(slice *[]int) {
		*slice = append(*slice, c)
	}
}

func color(c int) TFunc {
	return func(slice *[]int) {
		*slice = append(*slice, c)
	}
}

func sandglass(params ...TFunc) {
	args := new([]int)

	for _, arg := range params {
		arg(args)
	}

	n := len(*args)

	var size int
	var char rune
	var color int

	char = 'X' // default value
	color = 0  // default value

	switch {
	case n == 0:
		fmt.Println("Error. Not enough arguments")
		return
	case n == 1:
		size = (*args)[0]
	case n == 2:
		size = (*args)[0]
		char = rune((*args)[1])
	case n == 3:
		size = (*args)[0]
		char = rune((*args)[1])
		color = (*args)[2]
	default:
		fmt.Println("Error. To many arguments")
		return
	}

	// checking the conditions
	if size <= 0 {
		fmt.Print("Error. Wrong size\n")
		return
	}

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
	sandglass(size(5), char('c'), color(32))
}
