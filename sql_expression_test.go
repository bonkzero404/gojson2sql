package gojson2sql

import (
	"testing"
)

func TestGetSqlExpression(t *testing.T) {

	if GetSqlExpression("=", "STRING", false, []byte(`"value"`)) != "= JQL_VALUE:'value'" {
		t.Error("Expected \"= JQL_VALUE:'value'\", got", GetSqlExpression("=", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("<>", "STRING", false, []byte(`"value"`)) != "<> JQL_VALUE:'value'" {
		t.Error("Expected \"<> JQL_VALUE:'value'\", got", GetSqlExpression("<>", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("<", "NUMBER", false, []byte(`1`)) != "< JQL_VALUE:1" {
		t.Error("Expected \"< JQL_VALUE:1\", got", GetSqlExpression("<", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression("<=", "NUMBER", false, []byte(`1`)) != "<= JQL_VALUE:1" {
		t.Error("Expected \"<= JQL_VALUE:1\", got", GetSqlExpression("<=", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression(">", "NUMBER", false, []byte(`1`)) != "> JQL_VALUE:1" {
		t.Error("Expected \"> JQL_VALUE:1\", got", GetSqlExpression(">", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression(">=", "NUMBER", false, []byte(`1`)) != ">= JQL_VALUE:1" {
		t.Error("Expected \">= JQL_VALUE:1\", got", GetSqlExpression(">=", "NUMBER", false, []byte(`1`)))
	}
	if GetSqlExpression("LIKE", "STRING", false, []byte(`"value"`)) != "LIKE JQL_VALUE:'value'" {
		t.Error("Expected \"LIKE JQL_VALUE:'value'\", got", GetSqlExpression("LIKE", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("BETWEEN", "NUMBER", false, []byte(`{"from": 1, "to": 2}`)) != "BETWEEN JQL_VALUE:1 AND JQL_VALUE:2" {
		t.Error("Expected \"BETWEEN JQL_VALUE:1 AND JQL_VALUE:2\", got", GetSqlExpression("BETWEEN", "NUMBER", false, []byte(`{"from": 1, "to": 2}`)))
	}
	if GetSqlExpression("NOT LIKE", "STRING", false, []byte(`"value"`)) != "NOT LIKE JQL_VALUE:'value'" {
		t.Error("Expected \"NOT LIKE JQL_VALUE:'value'\", got", GetSqlExpression("NOT LIKE", "STRING", false, []byte(`"value"`)))
	}
	if GetSqlExpression("IN", "ARRAY", false, []byte("[\"value1\", \"value2\"]")) != "IN (JQL_VALUE:'value1', JQL_VALUE:'value2')" {
		t.Error("Expected \"IN (JQL_VALUE:'value1', JQL_VALUE:'value2')\", got", GetSqlExpression("IN", "ARRAY", false, []byte("[\"value1\", \"value2\"]")))
	}
	if GetSqlExpression("NOT IN", "ARRAY", false, []byte(`[1, 2]`)) != "NOT IN (JQL_VALUE:1, JQL_VALUE:2)" {
		t.Error("Expected \"NOT IN (JQL_VALUE:1, JQL_VALUE:2)\", got", GetSqlExpression("NOT IN", "ARRAY", false, []byte(`[1, 2]`)))
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
