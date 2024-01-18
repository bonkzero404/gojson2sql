package gojson2sql

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

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

	jqlFlag := ""

	if !isStatic {
		jqlFlag = "JQL_VALUE:"
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
			tmpValueArrayString = append(tmpValueArrayString, jqlFlag+"'"+item+"'")
		}

		return strings.Join(tmpValueArrayString, ", ")
	case "ArrayNumber":
		var valueArrayNumber []float64
		json.Unmarshal(value, &valueArrayNumber)
		var valueArrayString []string
		for _, item := range valueArrayNumber {
			valueArrayString = append(valueArrayString, jqlFlag+strconv.FormatFloat(item, 'f', -1, 64))
		}
		return strings.Join(valueArrayString, ", ")

	default:
		return ""
	}
}

func ExtractValueByDataType(datatype SQLDataTypeEnum, value json.RawMessage, isStatic bool) string {
	switch datatype {
	case String:
		var valueString string
		jqlFlag := ""

		if !isStatic {
			jqlFlag = "JQL_VALUE:"
		}
		json.Unmarshal(value, &valueString)
		return jqlFlag + "'" + valueString + "'"
	case Boolean:
		var valueBool bool
		jqlFlag := ""

		if !isStatic {
			jqlFlag = "JQL_VALUE:"
		}

		json.Unmarshal(value, &valueBool)
		return jqlFlag + string(value)
	case Number:
		var valueNumber float64
		jqlFlag := ""

		if !isStatic {
			jqlFlag = "JQL_VALUE:"
		}

		json.Unmarshal(value, &valueNumber)
		return jqlFlag + string(value)
	case Raw:
		jqlFlag := ""

		if !isStatic {
			jqlFlag = "JQL_VALUE:"
		}

		return jqlFlag + string(value)
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
