package main

import (
	"strconv"
	"time"
)

// This struct will be used after testing on validity
type WorkingCompany struct {
	WCompany   string
	WType      string
	WValue     int
	WId        interface{}
	WCreatedAt string
	IsValid    bool
}

// This structs are using for parsing json
type ParsingCompany struct {
	Company   interface{} `json:"company,omitempty"`
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	ID        interface{} `json:"id,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
	Operation `json:"operation,omitempty"`
}
type Operation struct {
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	ID        interface{} `json:"id,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
}

// This struct contains info of each company
type Result struct {
	Company  string        `json:"company"`
	Count    int           `json:"valid_operations_count"`
	Balance  int           `json:"balance"`
	Invalids []interface{} `json:"invalid_operations,omitempty"`
}

//
// The function tests for validity and marks it as valid
func CheckValid(c ParsingCompany) (WorkingCompany, bool) {
	var ans WorkingCompany
	ans.IsValid = true
	// Checking the name of a company
	if s, ok := c.Company.(string); !ok || s == "" {
		return WorkingCompany{}, false
	}
	ans.WCompany = c.Company.(string)

	// Checking Created_at
	if _, err := time.Parse(time.RFC3339, c.CreatedAt); err != nil {
		if _, err := time.Parse(time.RFC3339, c.Operation.CreatedAt); err != nil {
			return WorkingCompany{}, false
		}
		ans.WCreatedAt = c.CreatedAt
	} else {
		ans.WCreatedAt = c.CreatedAt
	}

	// Checking  id
	switch v := c.ID.(type) {
	case float64:
		if v != float64(int(v)) {
			return WorkingCompany{}, false
		}
		ans.WId = int(v)
	case string:
		// Ok
		ans.WId = v

	// Also for Operation
	default:
		switch v := c.Operation.ID.(type) {
		case float64:
			if v != float64(int(v)) {
				return WorkingCompany{}, false
			}
			ans.WId = int(v)
		case string:
			// Ok
			ans.WId = v
		default:
			return WorkingCompany{}, false
		}
	}

	// Checking the type in Company
	if c.Type != "income" && c.Type != "outcome" && c.Type != "+" && c.Type != "-" {
		// Also for Operation
		if c.Operation.Type != "income" && c.Operation.Type != "outcome" && c.Operation.Type != "+" && c.Operation.Type != "-" {
			ans.IsValid = false
		}
		ans.WType = c.Operation.Type
	} else {
		ans.WType = c.Type
	}

	// Checking the value
	switch v := c.Value.(type) {
	case float64:
		if v != float64(int(v)) {
			ans.IsValid = false
		}
		ans.WValue = int(v)
	case string:
		f, err2 := strconv.ParseFloat(v, 64)
		if err2 != nil {
			ans.IsValid = false
		}
		if f != float64(int(f)) {
			ans.IsValid = false
		}
		ans.WValue = int(f)

	default:
		// Also for Operation
		switch v := c.Operation.Value.(type) {
		case float64:
			if v != float64(int(v)) {
				ans.IsValid = false
			}
			ans.WValue = int(v)
		case string:
			f, err2 := strconv.ParseFloat(v, 64)
			if err2 != nil {
				ans.IsValid = false
			}
			if f != float64(int(f)) {
				ans.IsValid = false
			}
			ans.WValue = int(f)

		default:
			ans.IsValid = false
		}
	}

	return ans, true
}

func Fill(r *Result, c []WorkingCompany) {
	for _, val := range c {
		if val.IsValid {
			r.Count++
			if val.WType == "income" || val.WType == "+" {
				r.Balance += val.WValue
			} else {
				r.Balance -= val.WValue
			}
		} else {
			r.Invalids = append(r.Invalids, val.WId)
		}
	}
}
