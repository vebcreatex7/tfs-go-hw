package main

import (
	"strconv"
	"time"
)

// This struct will be used after testing on validity
type WorkingBill struct {
	WName      string
	WType      string
	WValue     int
	WId        interface{}
	WCreatedAt string
	IsValid    bool
}

// This structs are using for parsing json
type Bill struct {
	Operation
	Op Operation `json:"operation,omitempty"`
}

type Operation struct {
	Name      interface{} `json:"company,omitempty"`
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	ID        interface{} `json:"id,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
}

func (c Operation) NameChecking() bool {
	if s, ok := c.Name.(string); !ok || s == "" {
		return false
	}
	return true
}

// Checking the type
func (c Operation) TypeChecking() bool {
	if c.Type != "income" && c.Type != "outcome" && c.Type != "+" && c.Type != "-" {
		return false
	}
	return true
}

// Checking the value
func (c Operation) ValueChecking() bool {
	switch v := c.Value.(type) {
	case float64:
		if v != float64(int(v)) {
			return false
		}
		return true
	case string:
		_, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		return true
	default:
		return false
	}
}

// Checking the id
func (c Operation) IDChecking() bool {
	switch v := c.ID.(type) {
	case float64:
		if v != float64(int(v)) {
			return false
		}
		return true
	case string:
		return true
	default:
		return false
	}
}

// Checking the date
func (c Operation) CreatedAtChecking() bool {
	if _, err := time.Parse(time.RFC3339, c.CreatedAt); err != nil {
		return false
	}
	return true
}

// This struct contains info of each company
type Result struct {
	Name     string        `json:"company"`
	Count    int           `json:"valid_operations_count"`
	Balance  int           `json:"balance"`
	Invalids []interface{} `json:"invalid_operations,omitempty"`
}

//
func setValue(w *WorkingBill, i interface{}) {
	switch v := i.(type) {
	case float64:
		w.WValue = int(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		w.WValue = int(f)
	}
}

// The function tests for validity and marks it as valid
func Check(b Bill) (WorkingBill, bool) {
	var ans WorkingBill
	ans.IsValid = true

	// Checking the name of a company
	if ok := b.NameChecking(); !ok {
		return WorkingBill{}, false
	}
	ans.WName = b.Name.(string)

	// Checking Created_at
	if ok := b.CreatedAtChecking(); !ok {
		if ok := b.Op.CreatedAtChecking(); !ok {
			return WorkingBill{}, false
		}
		ans.WCreatedAt = b.Op.CreatedAt
	} else {
		ans.WCreatedAt = b.CreatedAt
	}

	// Checking  id
	if ok := b.IDChecking(); !ok {
		if ok := b.Op.IDChecking(); !ok {
			return WorkingBill{}, false
		}
		ans.WId = b.Op.ID
	} else {
		ans.WId = b.ID
	}

	// Checking the type
	if ok := b.TypeChecking(); !ok {
		if ok := b.Op.TypeChecking(); !ok {
			ans.IsValid = false
		} else {
			ans.WType = b.Op.Type
		}
	} else {
		ans.WType = b.Type
	}

	// Chrcking the value
	if ok := b.ValueChecking(); !ok {
		if ok := b.Op.ValueChecking(); !ok {
			ans.IsValid = false
		} else {
			setValue(&ans, b.Op.Value)
		}
	} else {
		setValue(&ans, b.Value)
	}

	return ans, true
}

func Fill(r *Result, c []WorkingBill) {
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
