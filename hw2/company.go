package main

import (
	"strconv"
	"time"
)

type Working struct {
	WorkingType      string
	WorkingValue     int
	WorkingID        interface{}
	WorkingCreatedAt string
}

type Operation struct {
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	ID        interface{} `json:"id,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
}

type Company struct {
	Company   interface{} `json:"company,omitempty"`
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	ID        interface{} `json:"id,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
	Operation `json:"operation,omitempty"`

	IsSkipped bool
	IsValid   bool

	// This struct will be used after testing on validity
	// To make it esier to search in struct
	Working `json:"working"`
}

type Result struct {
	Company  string        `json:"company"`
	Count    int           `json:"valid_operations_count"`
	Balance  int           `json:"balance"`
	Invalids []interface{} `json:"invalid_operations,omitempty"`
}

//
// The function tests for validity and marks it as valid
func SetValid(c *Company) {
	// Checking the name of a company
	if s, ok := c.Company.(string); !ok || s == "" {
		c.IsSkipped = true
		return
	}

	// Checking Created_at
	if _, err := time.Parse(time.RFC3339, c.CreatedAt); err != nil {
		if _, err := time.Parse(time.RFC3339, c.Operation.CreatedAt); err != nil {
			c.IsSkipped = true
			return
		}
		c.WorkingCreatedAt = c.CreatedAt
	} else {
		c.WorkingCreatedAt = c.CreatedAt
	}

	// Checking  id
	switch v := c.ID.(type) {
	case float64:
		if v != float64(int(v)) {
			c.IsSkipped = true
			return
		}
		c.WorkingID = int(v)
	case string:
		// Ok
		c.WorkingID = v

	// Also for Operation
	default:
		switch v := c.Operation.ID.(type) {
		case float64:
			if v != float64(int(v)) {
				c.IsSkipped = true
				return
			}
			c.WorkingID = int(v)
		case string:
			// Ok
			c.WorkingID = v
		default:
			c.IsSkipped = true

			return
		}
	}

	// Checking the type in Company
	if c.Type != "income" && c.Type != "outcome" && c.Type != "+" && c.Type != "-" {
		// Also for Operation
		if c.Operation.Type != "income" && c.Operation.Type != "outcome" && c.Operation.Type != "+" && c.Operation.Type != "-" {
			c.IsValid = false
			return
		}
		c.WorkingType = c.Operation.Type
	} else {
		c.WorkingType = c.Type
	}

	//
	// Checking the value
	switch v := c.Value.(type) {
	case float64:
		if v != float64(int(v)) {
			c.IsValid = false
			return
		}
		c.WorkingValue = int(v)
	case string:

		f, err2 := strconv.ParseFloat(v, 64)
		if err2 != nil {
			c.IsValid = false
			return
		}
		if f != float64(int(f)) {
			c.IsValid = false
			return
		}
		c.WorkingValue = int(f)

	default:
		// Also for Operation
		switch v := c.Operation.Value.(type) {
		case float64:
			if v != float64(int(v)) {
				c.IsValid = false
				return
			}
			c.WorkingValue = int(v)
			c.Value = int(v)
		case string:
			f, err2 := strconv.ParseFloat(v, 64)
			if err2 != nil {
				c.IsValid = false
				return
			}
			if f != float64(int(f)) {
				c.IsValid = false
				return
			}
			c.WorkingValue = int(f)

		default:
			c.IsValid = false
			return
		}
	}

	c.IsValid = true
}

func Fill(r *Result, c []Company) {
	for _, val := range c {
		if val.IsValid {
			r.Count++
			if val.WorkingType == "income" || val.WorkingType == "+" {
				r.Balance += val.WorkingValue
			} else {
				r.Balance -= val.WorkingValue
			}
		} else {
			r.Invalids = append(r.Invalids, val.WorkingID)
		}
	}
}
