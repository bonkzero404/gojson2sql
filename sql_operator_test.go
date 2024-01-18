package gojson2sql

import "testing"

func TestIsValidOperator(t *testing.T) {
	if !IsValidOperator("=") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("<>") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("<") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("<=") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator(">") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator(">=") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("LIKE") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("BETWEEN") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("NOT LIKE") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("IN") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("NOT IN") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("IS NULL") {
		t.Error("Expected true, got false")
	}
	if !IsValidOperator("IS NOT NULL") {
		t.Error("Expected true, got false")
	}
	if IsValidOperator("invalid") {
		t.Error("Expected false, got true")
	}
}

func TestGetValueFromOperator(t *testing.T) {
	if _, err := GetValueFromOperator("="); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("<>"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("<"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("<="); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator(">"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator(">="); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("LIKE"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("BETWEEN"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("NOT LIKE"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("IN"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("NOT IN"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("IS NULL"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("IS NOT NULL"); err != nil {
		t.Error("Expected nil, got", err)
	}
	if _, err := GetValueFromOperator("invalid"); err == nil {
		t.Error("Expected error, got nil")
	}
}
