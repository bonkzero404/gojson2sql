package gojson2sql

import (
	"testing"
)

func TestIsValidDataType(t *testing.T) {
	if !IsValidDataType("BOOLEAN") {
		t.Error("Expected true, got false")
	}
	if !IsValidDataType("STRING") {
		t.Error("Expected true, got false")
	}
	if !IsValidDataType("NUMBER") {
		t.Error("Expected true, got false")
	}
	if !IsValidDataType("RAW") {
		t.Error("Expected true, got false")
	}
	if !IsValidDataType("FUNCTION") {
		t.Error("Expected true, got false")
	}
	if IsValidDataType("invalid") {
		t.Error("Expected false, got true")
	}
}

func TestGetValueFromDataType(t *testing.T) {
	if _, err := GetValueFromDataType("BOOLEAN"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromDataType("STRING"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromDataType("NUMBER"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromDataType("RAW"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromDataType("FUNCTION"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromDataType("invalid"); err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCheckArrayType(t *testing.T) {
	if _, err := checkArrayType([]byte("[\"1\", \"2\"]")); err != nil {
		t.Error("Expected true, got false")
	}
	if _, err := checkArrayType([]byte("[1, 2]")); err != nil {
		t.Error("Expected true, got false")
	}
	if _, err := checkArrayType([]byte("1")); err == nil {
		t.Error("Expected false, got true")
	}
}

func TestArrayConversionToStringExpression(t *testing.T) {
	if ArrayConversionToStringExpression([]byte("[\"1\", \"2\"]"), false) != "JQL_VALUE:'1', JQL_VALUE:'2'" {
		t.Error("Expected \"JQL_VALUE:'1', JQL_VALUE:'2'\", got", ArrayConversionToStringExpression([]byte("[\"1\", \"2\"]"), false))
	}
	if ArrayConversionToStringExpression([]byte("[1, 2]"), false) != "JQL_VALUE:1, JQL_VALUE:2" {
		t.Error("Expected \"JQL_VALUE:1, JQL_VALUE:2\", got", ArrayConversionToStringExpression([]byte("[1, 2]"), false))
	}
	if ArrayConversionToStringExpression([]byte("1"), false) != "" {
		t.Error("Expected \"\", got", ArrayConversionToStringExpression([]byte("1"), false))
	}
}

func TestExtractValueByDataType(t *testing.T) {
	if ExtractValueByDataType("BOOLEAN", []byte("true"), false) != "JQL_VALUE:true" {
		t.Error("Expected true, got false", ExtractValueByDataType("BOOLEAN", []byte("true"), false))
	}
	if ExtractValueByDataType("STRING", []byte("\"string\""), false) != "JQL_VALUE:'string'" {
		t.Error("Expected 'string', got", ExtractValueByDataType("STRING", []byte("\"string\""), false))
	}
	if ExtractValueByDataType("NUMBER", []byte("1"), false) != "JQL_VALUE:1" {
		t.Error("Expected 1, got", ExtractValueByDataType("NUMBER", []byte("1"), false))
	}
	if ExtractValueByDataType("FUNCTION", []byte("{\"sqlFunc\": {\"name\": \"func\", \"params\": [\"param1\", \"param2\"]}}"), false) != "func(JQL_VALUE:'param1', JQL_VALUE:'param2')" {
		t.Error("Expected \"func(JQL_VALUE:'param1', JQL_VALUE:'param2')\", got", ExtractValueByDataType("FUNCTION", []byte("{\"sqlFunc\": {\"name\": \"func\", \"params\": [\"param1\", \"param2\"]}}"), false))
	}
	if ExtractValueByDataType("FUNCTION", []byte("{\"sqlFunc\": {\"name\": \"func\", \"isField\": true, \"params\": [\"param1\"]}}"), false) != "func(param1)" {
		t.Error("Expected \"func(param1)\", got", ExtractValueByDataType("FUNCTION", []byte("{\"sqlFunc\": {\"name\": \"func\", \"isField\": true, \"params\": [\"param1\"]}}"), false))
	}
	if ExtractValueByDataType("RAW", []byte("raw"), false) != "JQL_VALUE:raw" {
		t.Error("Expected \"JQL_VALUE:raw\", got", ExtractValueByDataType("RAW", []byte("raw"), false))
	}

	if ExtractValueByDataType("ARRAY", []byte("[\"1\", \"2\"]"), false) != "JQL_VALUE:'1', JQL_VALUE:'2'" {
		t.Error("Expected \"JQL_VALUE:'1', JQL_VALUE:'2'\", got", ExtractValueByDataType("ARRAY", []byte("[\"1\", \"2\"]"), false))
	}

	if ExtractValueByDataType("ARRAY", []byte("[1, 2]"), false) != "JQL_VALUE:1, JQL_VALUE:2" {
		t.Error("Expected \"JQL_VALUE:1, JQL_VALUE:2\", got", ExtractValueByDataType("ARRAY", []byte("[JQL_VALUE:1, JQL_VALUE:2]"), false))
	}
}
