package query

import (
	"fmt"
	"testing"
)

func TestBuild(t *testing.T) {
	// Test with empty conds
	conds := []Conds{}
	query, values := Build(conds)
	if query != "" || values != nil {
		t.Errorf("Build() with empty conds returned non-empty query or values")
	}

	// Test with skip condition
	conds = []Conds{
		{
			Key:    "name = ?",
			Values: []interface{}{"John"},
			Skip:   true,
		},
		{
			Key:    "age = ?",
			Values: []interface{}{30},
			Skip:   false,
		},
	}
	query, values = Build(conds)
	if query != "age = $1" || fmt.Sprintf("%v", values) != "[30]" {
		t.Errorf("Build() with skip condition returned incorrect query or values")
	}

	// Test with non-empty values
	conds = []Conds{
		{
			Key:    "name = ?",
			Values: []interface{}{"John"},
			Skip:   false,
		},
		{
			Key:    "age = ?",
			Values: []interface{}{30},
			Skip:   false,
		},
	}
	query, values = Build(conds)
	if query != "name = $1 AND age = $2" || fmt.Sprintf("%v", values) != "[John 30]" {
		t.Errorf("Build() with non-empty values returned incorrect query or values")
	}

	// Test with empty values
	conds = []Conds{
		{
			Key:    "name = ?",
			Values: []interface{}{},
			Skip:   false,
		},
		{
			Key:    "age = ?",
			Values: []interface{}{},
			Skip:   false,
		},
	}
	query, values = Build(conds)
	if query != "" || values != nil {
		t.Errorf("Build() with empty values returned non-empty query or values")
	}

	// Test with LIMIT condition
	conds = []Conds{
		{
			Key:    "name = ?",
			Values: []interface{}{"John"},
			Skip:   false,
		},
		{
			Key:    "age = ?",
			Values: []interface{}{30},
			Skip:   false,
		},
		{
			Key:    "LIMIT ? OFFSET ?",
			Values: []interface{}{10, 0},
			Skip:   false,
		},
	}
	query, values = Build(conds)
	if query != "name = $1 AND age = $2 LIMIT $3 OFFSET $4" || fmt.Sprintf("%v", values) != "[John 30 10 0]" {
		t.Errorf("Build() with LIMIT condition returned incorrect query or values")
	}
}
func TestReplacePlaceholder(t *testing.T) {
	// Test with single question mark
	input := "SELECT * FROM table WHERE id = ?"
	expected := "SELECT * FROM table WHERE id = $1"
	output := ReplacePlaceholder(input)
	if output != expected {
		t.Errorf("ReplaceQuestionMarkToDollarSign() failed, expected: %s, got: %s", expected, output)
	}

	// Test with multiple question marks
	input = "INSERT INTO table (name, age) VALUES (?, ?)"
	expected = "INSERT INTO table (name, age) VALUES ($1, $2)"
	output = ReplacePlaceholder(input)
	if output != expected {
		t.Errorf("ReplaceQuestionMarkToDollarSign() failed, expected: %s, got: %s", expected, output)
	}

	// Test with no question mark
	input = "SELECT * FROM table"
	expected = "SELECT * FROM table"
	output = ReplacePlaceholder(input)
	if output != expected {
		t.Errorf("ReplaceQuestionMarkToDollarSign() failed, expected: %s, got: %s", expected, output)
	}
}
