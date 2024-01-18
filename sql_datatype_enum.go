package gojson2sql

type SQLDataTypeEnum string

const (
	Boolean  SQLDataTypeEnum = "BOOLEAN"
	String   SQLDataTypeEnum = "STRING"
	Number   SQLDataTypeEnum = "NUMBER"
	Raw      SQLDataTypeEnum = "RAW"
	Function SQLDataTypeEnum = "FUNCTION"
	Array    SQLDataTypeEnum = "ARRAY"
)
