package gojson2sql

import (
	"encoding/json"
	"strings"
)

func GetSqlExpression(operator SQLOperatorEnum, datatype SQLDataTypeEnum, isStatic bool, value ...json.RawMessage) string {
	op := strings.ToUpper(string(operator))
	dt := strings.ToUpper(string(datatype))

	switch op {
	case string(Equal), string(NotEqual), string(LessThan), string(LessEqual), string(GreaterThan), string(GreaterEqual):
		return string(op) + " " + ExtractValueByDataType(SQLDataTypeEnum(dt), value[0], isStatic)
	case string(Like), string(Ilike), string(NotLike):
		return string(op) + " " + ExtractValueByDataType(SQLDataTypeEnum(dt), value[0], isStatic)
	case string(Between):
		var valueRange ValueRange
		json.Unmarshal(value[0], &valueRange)
		return string(op) + " " + ExtractValueByDataType(SQLDataTypeEnum(dt), json.RawMessage(valueRange.From), isStatic) + " AND " + ExtractValueByDataType(SQLDataTypeEnum(dt), json.RawMessage(valueRange.To), isStatic)
	case string(In), string(NotIn):
		var values = ExtractValueByDataType(SQLDataTypeEnum(dt), json.RawMessage(value[0]), isStatic)
		return string(op) + " (" + values + ")"
	case string(IsNull), string(IsNotNull):
		return string(op)
	default:
		return ""
	}
}
