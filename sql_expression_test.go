package gojson2sql

import (
	"testing"
)

func TestGetSqlExpression(t *testing.T) {

	if GetSqlExpression("=", "STRING", false, []byte(`"value"`)) != "= JQL_VALUE:'value':END_JQL_VALUE" {
		t.Error("Expected \"= JQL_VALUE:'value':END_JQL_VALUE\", got", GetSqlExpression("=", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("<>", "STRING", false, []byte(`"value"`)) != "<> JQL_VALUE:'value':END_JQL_VALUE" {
		t.Error("Expected \"<> JQL_VALUE:'value':END_JQL_VALUE\", got", GetSqlExpression("<>", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("<", "NUMBER", false, []byte(`1`)) != "< JQL_VALUE:1:END_JQL_VALUE" {
		t.Error("Expected \"< JQL_VALUE:1:END_JQL_VALUE\", got", GetSqlExpression("<", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression("<=", "NUMBER", false, []byte(`1`)) != "<= JQL_VALUE:1:END_JQL_VALUE" {
		t.Error("Expected \"<= JQL_VALUE:1:END_JQL_VALUE\", got", GetSqlExpression("<=", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression(">", "NUMBER", false, []byte(`1`)) != "> JQL_VALUE:1:END_JQL_VALUE" {
		t.Error("Expected \"> JQL_VALUE:1:END_JQL_VALUE\", got", GetSqlExpression(">", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression(">=", "NUMBER", false, []byte(`1`)) != ">= JQL_VALUE:1:END_JQL_VALUE" {
		t.Error("Expected \">= JQL_VALUE:1:END_JQL_VALUE\", got", GetSqlExpression(">=", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression("LIKE", "STRING", false, []byte(`"value"`)) != "LIKE JQL_VALUE:'value':END_JQL_VALUE" {
		t.Error("Expected \"LIKE JQL_VALUE:'value':END_JQL_VALUE\", got", GetSqlExpression("LIKE", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("BETWEEN", "NUMBER", false, []byte(`{"from": 1, "to": 2}`)) != "BETWEEN JQL_VALUE:1:END_JQL_VALUE AND JQL_VALUE:2:END_JQL_VALUE" {
		t.Error("Expected \"BETWEEN JQL_VALUE:1:END_JQL_VALUE AND JQL_VALUE:2:END_JQL_VALUE\", got", GetSqlExpression("BETWEEN", "NUMBER", false, []byte(`{"from": 1, "to": 2}`)))
	}
	if GetSqlExpression("NOT LIKE", "STRING", false, []byte(`"value"`)) != "NOT LIKE JQL_VALUE:'value':END_JQL_VALUE" {
		t.Error("Expected \"NOT LIKE JQL_VALUE:'value':END_JQL_VALUE\", got", GetSqlExpression("NOT LIKE", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("IN", "ARRAY", false, []byte("[\"value1\", \"value2\"]")) != "IN (JQL_VALUE:'value1':END_JQL_VALUE, JQL_VALUE:'value2':END_JQL_VALUE)" {
		t.Error("Expected \"IN (JQL_VALUE:'value1':END_JQL_VALUE, JQL_VALUE:'value2':END_JQL_VALUE)\", got", GetSqlExpression("IN", "ARRAY", false, []byte("[\"value1\", \"value2\"]")))
	}
	if GetSqlExpression("NOT IN", "ARRAY", false, []byte(`[1, 2]`)) != "NOT IN (JQL_VALUE:1:END_JQL_VALUE, JQL_VALUE:2:END_JQL_VALUE)" {
		t.Error("Expected \"NOT IN (JQL_VALUE:1:END_JQL_VALUE, JQL_VALUE:2:END_JQL_VALUE)\", got", GetSqlExpression("NOT IN", "ARRAY", false, []byte(`[1, 2]`)))
	}
	if GetSqlExpression("IS NULL", "STRING", false) != "IS NULL" {
		t.Error("Expected \"IS NULL\", got", GetSqlExpression("IS NULL", "STRING", false))
	}
	if GetSqlExpression("IS NOT NULL", "STRING", false) != "IS NOT NULL" {
		t.Error("Expected \"IS NOT NULL\", got", GetSqlExpression("IS NOT NULL", "STRING", false))
	}
	if GetSqlExpression("invalid", "STRING", false) != "" {
		t.Error("Expected \"\", got", GetSqlExpression("invalid", "STRING", false))
	}

}
