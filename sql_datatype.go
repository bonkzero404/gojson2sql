package gojson2sql

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

var JQL_FLAG_OPEN = "JQL_VALUE:"
var JQL_FLAG_CLOSE = ":END_JQL_VALUE"

func IsValidDataType(datatype string) bool {
	switch SQLDataTypeEnum(datatype) {
	case String, Boolean, Number, Raw, Function:
		return true
	default:
		return false
	}
}

func GetValueFromDataType(datatype string) (SQLDataTypeEnum, error) {
	if IsValidDataType(datatype) {
		return SQLDataTypeEnum(datatype), nil
	}
	return "", errors.New("invalid SQL datatype")
}

func checkArrayType(raw json.RawMessage) (string, error) {
	var arrayData []interface{}

	err := json.Unmarshal(raw, &arrayData)

	for _, item := range arrayData {
		switch item.(type) {
		case string:
			return "ArrayString", nil
		case float64:
			return "ArrayNumber", nil
		}
	}

	return "Unknown array type", err
}

func ArrayConversionToStringExpression(value json.RawMessage, isStatic bool, isField ...bool) string {
	valCheckArrayType, _ := checkArrayType(value)

	jqlFlagOpen := ""
	jqlFlagClose := ""

	if !isStatic {
		jqlFlagOpen = JQL_FLAG_OPEN
		jqlFlagClose = JQL_FLAG_CLOSE
	}

	if isField != nil && isField[0] {
		var valueArray []string
		json.Unmarshal(value, &valueArray)
		return strings.Join(valueArray, ", ")
	}

	switch valCheckArrayType {
	case "ArrayString":
		var valueArrayString []string

		json.Unmarshal(value, &valueArrayString)

		var tmpValueArrayString []string
		for _, item := range valueArrayString {
			tmpValueArrayString = append(tmpValueArrayString, jqlFlagOpen+"'"+item+"'"+jqlFlagClose)
		}

		return strings.Join(tmpValueArrayString, ", ")
	case "ArrayNumber":
		var valueArrayNumber []float64
		json.Unmarshal(value, &valueArrayNumber)
		var valueArrayString []string
		for _, item := range valueArrayNumber {
			valueArrayString = append(valueArrayString, jqlFlagOpen+strconv.FormatFloat(item, 'f', -1, 64)+jqlFlagClose)
		}
		return strings.Join(valueArrayString, ", ")

	default:
		return ""
	}
}

func ExtractValueByDataType(datatype SQLDataTypeEnum, value json.RawMessage, isStatic bool) string {
	var valueString string
	jqlFlagOpen := ""
	jqlFlagClose := ""

	if !isStatic {
		jqlFlagOpen = JQL_FLAG_OPEN
		jqlFlagClose = JQL_FLAG_CLOSE
	}

	switch datatype {
	case String:
		json.Unmarshal(value, &valueString)
		return jqlFlagOpen + "'" + valueString + "'" + jqlFlagClose
	case Boolean:
		var valueBool bool

		json.Unmarshal(value, &valueBool)
		return jqlFlagOpen + string(value) + jqlFlagClose
	case Number:
		var valueNumber float64

		json.Unmarshal(value, &valueNumber)
		return jqlFlagOpen + string(value) + jqlFlagClose
	case Raw:
		return jqlFlagOpen + string(value) + jqlFlagClose
	case Array:
		return ArrayConversionToStringExpression(value, isStatic)
	case Function:
		var valueFunction SqlFunc
		json.Unmarshal(value, &valueFunction)

		if valueFunction.SqlFunc.IsField != nil && *valueFunction.SqlFunc.IsField {
			newValue := ArrayConversionToStringExpression(valueFunction.SqlFunc.Params, isStatic, *valueFunction.SqlFunc.IsField)
			return valueFunction.SqlFunc.Name + "(" + newValue + ")"
		}

		newValue := ArrayConversionToStringExpression(valueFunction.SqlFunc.Params, isStatic)
		return valueFunction.SqlFunc.Name + "(" + newValue + ")"
	default:
		return ""
	}
}
