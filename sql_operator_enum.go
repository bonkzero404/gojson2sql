package gojson2sql

type SQLOperatorEnum string

const (
	Equal        SQLOperatorEnum = "="
	NotEqual     SQLOperatorEnum = "<>"
	LessThan     SQLOperatorEnum = "<"
	LessEqual    SQLOperatorEnum = "<="
	GreaterThan  SQLOperatorEnum = ">"
	GreaterEqual SQLOperatorEnum = ">="
	Like         SQLOperatorEnum = "LIKE"
	Ilike        SQLOperatorEnum = "ILIKE"
	Between      SQLOperatorEnum = "BETWEEN"
	NotLike      SQLOperatorEnum = "NOT LIKE"
	In           SQLOperatorEnum = "IN"
	NotIn        SQLOperatorEnum = "NOT IN"
	IsNull       SQLOperatorEnum = "IS NULL"
	IsNotNull    SQLOperatorEnum = "IS NOT NULL"
)
