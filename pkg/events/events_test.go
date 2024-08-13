package events

import (
	"testing"

	"github.com/nats-io/nats.go"
)

func TestParseJson(t *testing.T) {
	// Test with valid JSON
	data := []byte(`{"name": "John", "age": 30}`)
	var v map[string]interface{}
	msg := &nats.Msg{Data: data}
	err := ParseJson(msg, &v)
	if err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}

	// Test with invalid JSON
	invalidData := []byte(`{"name": "John", "age": 30`)
	invalidMsg := &nats.Msg{Data: invalidData}
	err = ParseJson(invalidMsg, &v)
	if err == nil {
		t.Errorf("Expected error, but got nil")
	}

	// Test with struct
	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var p Person
	err = ParseJson(msg, &p)
	if err != nil {
		t.Errorf("Failed to parse JSON into struct: %v", err)
	}
}
