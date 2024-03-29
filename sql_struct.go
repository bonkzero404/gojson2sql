package gojson2sql

import "github.com/goccy/go-json"

type Condition struct {
	Operand     *string           `json:"operand"`
	Clause      json.RawMessage   `json:"clause"`
	Operator    SQLOperatorEnum   `json:"operator"`
	Value       json.RawMessage   `json:"value"`
	IsStatic    *bool             `json:"isStatic"`
	Datatype    *SQLDataTypeEnum  `json:"datatype"`
	Composite   *[]Condition      `json:"composite"`
	Expectation *ExpectationField `json:"expectation"`
}

type SQLJson struct {
	Table        string             `json:"table"`
	SelectFields *[]json.RawMessage `json:"selectFields"`
	Join         *[]Join            `json:"join"`
	Where        *Condition         `json:"where"`
	Conditions   *[]Condition       `json:"conditions"`
	GroupBy      *struct {
		Fields []string `json:"fields"`
	} `json:"groupBy"`
	Having  *[]Condition `json:"having"`
	OrderBy *struct {
		Fields []string `json:"fields"`
		Sort   *string  `json:"sort"`
	} `json:"orderBy"`
	Limit  *json.RawMessage `json:"limit"`
	Offset *json.RawMessage `json:"offset"`
}

type Join struct {
	Table *string           `json:"table"`
	Type  *string           `json:"type"`
	On    map[string]string `json:"on"`
}

type ValueRange struct {
	From json.RawMessage `json:"from"`
	To   json.RawMessage `json:"to"`
}

type SqlFunc struct {
	SqlFunc struct {
		Name    string          `json:"name"`
		IsField *bool           `json:"isField"`
		Params  json.RawMessage `json:"params"`
	} `json:"sqlFunc"`
}

type SelectionFields struct {
	Field       string   `json:"field"`
	Alias       *string  `json:"alias"`
	SubQuery    *SQLJson `json:"subquery"`
	AddFunction *SqlFunc `json:"addFunction"`
}

type Case struct {
	When         *[]Condition    `json:"when"`
	DefaultValue json.RawMessage `json:"defaultValue"`
	Alias        *string         `json:"alias"`
}

type ValueAdjacent struct {
	Value    json.RawMessage  `json:"value"`
	Datatype *SQLDataTypeEnum `json:"datatype"`
	IsStatic *bool            `json:"isStatic"`
}

type LimitOffsetValue struct {
	IsStatic bool `json:"isStatic"`
	Value    int  `json:"value"`
}

type ExpectationField ValueAdjacent
type CaseDefauleValue ValueAdjacent
