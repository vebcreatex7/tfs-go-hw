package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"time"
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
	mBills := make(map[string][]WorkingBill)

	var sBills []Bill

	if err = json.Unmarshal(data, &sBills); err != nil {
		fmt.Println("err")
		return
	}

	for _, bill := range sBills {
		company, ok := Check(bill)
		if ok {
			mBills[company.WName] = append(mBills[company.WName], company)
		}
	}

	// Sorting by date
	for s := range mBills {
		sort.Slice(mBills[s], func(i, j int) bool {
			t1, _ := time.Parse(time.RFC3339, mBills[s][i].WCreatedAt)
			t2, _ := time.Parse(time.RFC3339, mBills[s][j].WCreatedAt)
			return t1.Before(t2)
		})
	}

	// Fill in information about companies
	sResults := make([]Result, 0, len(mBills))
	for s := range mBills {
		r := Result{Name: s}
		Fill(&r, mBills[s])
		sResults = append(sResults, r)
	}

	// Sorting by name of company
	sort.Slice(sResults, func(i, j int) bool {
		return sResults[i].Name < sResults[j].Name
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
