package gojson2sql

import (
	"errors"
)

func IsValidOperator(operator string) bool {
	switch SQLOperatorEnum(operator) {
	case Equal, NotEqual, LessThan, LessEqual, GreaterThan, GreaterEqual,
		Like, Between, NotLike, In, NotIn, IsNull, IsNotNull:
		return true
	default:
		return false
	}
}

func GetValueFromOperator(operator string) (SQLOperatorEnum, error) {
	if IsValidOperator(operator) {
		return SQLOperatorEnum(operator), nil
	}
	return "", errors.New("invalid SQL operator")
}
