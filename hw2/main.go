package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"strings"
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
	mCompannies := make(map[string][]WorkingCompany)
	dec := json.NewDecoder(strings.NewReader(string(data)))

	// Reading opening bracket
	_, err = dec.Token()
	if err != nil {
		fmt.Println("err")
		return
	}

	// while array contains values
	for dec.More() {
		var c ParsingCompany
		err = dec.Decode(&c)
		if err != nil {
			fmt.Println("err")
			return
		}
		company, flag := CheckValid(c)
		if flag {
			mCompannies[company.WCompany] = append(mCompannies[company.WCompany], company)
		} else {
			continue
		}
	}

	// Reading closing bracket
	_, err = dec.Token()
	if err != nil {
		fmt.Println("err")
		return
	}

	// Sorting by date
	for s := range mCompannies {
		sort.Slice(mCompannies[s], func(i, j int) bool {
			t1, _ := time.Parse(time.RFC3339, mCompannies[s][i].WCreatedAt)
			t2, _ := time.Parse(time.RFC3339, mCompannies[s][j].WCreatedAt)
			return t1.Before(t2)
		})
	}

	// Fill in information about companies
	sResults := make([]Result, 0, len(mCompannies))
	for s := range mCompannies {
		r := Result{Company: s}
		Fill(&r, mCompannies[s])
		sResults = append(sResults, r)
	}

	// Sorting by name of company
	sort.Slice(sResults, func(i, j int) bool {
		return sResults[i].Company < sResults[j].Company
	})

	// Unmarshaling
	newData, err := json.MarshalIndent(sResults, "", "\t")
	if err != nil {
		fmt.Println("Error in marshaling")
		return
	}

	// Writing to a file
	out, err := os.Create("out.json")
	if err != nil {
		fmt.Println("Can't open out.json")
		return
	}
	if _, err = out.Write(newData); err != nil {
		fmt.Println("Can't write to out.json")
		return
	}
}
