package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

func asEnv() (string, bool) {
	var file = "FILE"
	name, ok := os.LookupEnv(file)
	return name, ok && name == "billing.json"
}

func asFlag() (string, bool) {
	var name = flag.String("file", "", "--file=<file>\n")
	flag.Parse()
	return *name, *name == "billing.json"
}

func main() {

	var data []byte

	// Selecting a file
	var file *os.File
	var err error

	// Open a file using flags
	if name, ok := asFlag(); ok {
		file, err = os.Open(name)
		if err != nil {
			fmt.Println("Can't open a file")
			file.Close()
			return
		}

		// Open a file using an env variable
	} else if name, ok := asEnv(); ok {
		file, err = os.Open(name)
		if err != nil {
			fmt.Println("Can't open a file")
			file.Close()
			return
		}

		// Read from stdin
	} else {
		file = os.Stdin
	}

	// Reading from a file to a buffer
	data, err = ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("Can't read from a file")
		file.Close()
		return
	}
	file.Close()

	// Reading from a JSON to a slice of company
	var Companies []Company
	if err := json.Unmarshal(data, &Companies); err != nil {
		fmt.Println("Error in unmarshaling")
		return
	}

	for i, val := range Companies {
		SetValid(&val)
		Companies[i] = val
	}
	// Split into valid and invalid

	newData, _ := json.MarshalIndent(Companies, "", "\t")
	fmt.Println(string(newData))

}
