package main

import (
	"strconv"
	"time"
)

type Working struct {
	WorkingType      string
	WorkingValue     int
	WorkingId        interface{}
	WorkingCreatedAt string
}

type Operation struct {
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	Id        interface{} `json:"id,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
}

type Company struct {
	Company   string      `json:"company,omitempty"`
	Type      string      `json:"type,omitempty"`
	Value     interface{} `json:"value,omitempty"`
	Id        interface{} `json:"id,omitempty"`
	CreatedAt string      `json:"created_at,omitempty"`
	Operation `json:"operation,omitempty"`

	// This variables will be used after testing on validity
	// To make it esier to search in struct
	Working `json:"working"`

	IsValid bool
}

//
// The function tests for validity and marks it as valid
func SetValid(c *Company) {

	var err error

	/*
		var err error
		_, err = time.Parse(time.RFC3339, c.CreatedAt)
		if c.CreatedAt != "" && err != nil {
			c.IsValid = false
			return
		}

		// Checking the date in Operation
		_, err = time.Parse(time.RFC3339, c.Operation.CreatedAt)
		if c.Operation.CreatedAt != "" && err != nil {
			c.IsValid = false
			return
		}
	*/

	// Checking the type in Company
	if c.Type != "income" && c.Type != "outcome" && c.Type != "+" && c.Type != "-" {
		// Also for Operation
		if c.Operation.Type != "income" && c.Operation.Type != "outcome" && c.Operation.Type != "+" && c.Operation.Type != "-" {
			c.IsValid = false
			return
		} else {
			c.WorkingType = c.Operation.Type
		}
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

		/*
			i, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
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
			}
			c.WorkingValue = int(i)
		*/
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

			/*
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					f, err2 := strconv.ParseFloat(v, 64)
					if err2 != nil {
						c.IsValid = false
						return
					}
					if f != float64(int(f)) {
						c.IsValid = false
						return
					}
					c.Value = int(f)
				}
				c.Value = int(i)
				c.WorkingValue = int(i)
			*/
		default:
			c.IsValid = false
			return
		}
	}

	//Checking the id
	switch v := c.Id.(type) {
	case float64:
		if v != float64(int(v)) {
			c.IsValid = false
			return
		}
		c.WorkingId = int(v)
	case string:
		// Ok
		c.WorkingId = v

	default:
		// Also for Operation
		switch v := c.Operation.Id.(type) {
		case float64:
			if v != float64(int(v)) {
				c.IsValid = false
				return
			}
			c.WorkingId = int(v)
		case string:
			// Ok
			c.WorkingId = v
		default:
			c.IsValid = false
			return
		}

	}

	// Checking the date in Company
	if c.CreatedAt != "" {
		if _, err = time.Parse(time.RFC3339, c.CreatedAt); err != nil {
			c.IsValid = false
			return
		} else {
			c.WorkingCreatedAt = c.CreatedAt
		}
	} else {
		if _, err = time.Parse(time.RFC3339, c.Operation.CreatedAt); err != nil {
			c.IsValid = false
			return
		} else {
			c.WorkingCreatedAt = c.Operation.CreatedAt
		}
	}

	c.IsValid = true
}
