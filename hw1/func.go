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
	var char rune = 'X' // default value
	var color int = 0   // default value

	if n == 0 {
		fmt.Println("Error. Not enough arguments")
		return
	} else if n == 1 {
		size = (*args)[0]
	} else if n == 2 {
		size = (*args)[0]
		char = rune((*args)[1])
	} else if n == 3 {
		size = (*args)[0]
		char = rune((*args)[1])
		color = (*args)[2]
	} else {
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
	for i := 0; i < size; i += 1 {
		fmt.Printf("\033[%dm%c\033[0m", color, char)
	}
	fmt.Printf("\n")

	//The upper part
	hight := int(size / 2)
	for i := 1; i < hight; i += 1 {
		for j := 0; j < size; j += 1 {
			if j == i {
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			} else if j == size-i-1 {
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}

	//The middle
	if size%2 == 1 {
		for i := 0; i < int(size/2); i += 1 {
			fmt.Printf(" ")
		}
		fmt.Printf("\033[%dm%c\033[0m", color, char)
		for i := int(size/2) + 1; i < size; i += 1 {
			fmt.Printf(" ")
		}
		fmt.Printf("\n")
	}

	//The lower part
	for i := hight + 1; i < size-1; i += 1 {
		for j := 0; j < size; j += 1 {
			if j == i {
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			} else if j == size-i-1 {
				fmt.Printf("\033[%dm%c\033[0m", color, char)
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Printf("\n")
	}

	//The lower cover
	for i := 0; i < size; i += 1 {
		fmt.Printf("\033[%dm%c\033[0m", color, char)
	}
	fmt.Printf("\n")

}

func main() {

	sandglass(size(11), char('A'), color(31))
}
