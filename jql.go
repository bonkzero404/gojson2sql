package gojson2sql

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Json2Sql struct {
	sqlJson *SQLJson
}

func NewJson2Sql(jsonData json.RawMessage) *Json2Sql {
	var sqlJson *SQLJson

	err := json.Unmarshal(jsonData, &sqlJson)

	if err != nil {
		return nil
	}

	return &Json2Sql{sqlJson}
}

func cleanSpaces(input string) string {
	words := strings.Fields(input)
	result := strings.Join(words, " ")
	return result
}

func (jql *Json2Sql) JsonRawString(raw json.RawMessage) (string, bool) {
	var str string
	err := json.Unmarshal(raw, &str)
	return str, err == nil
}

func (jql *Json2Sql) JsonRawSqlFunc(raw json.RawMessage) (SqlFunc, bool) {
	var fn SqlFunc
	err := json.Unmarshal(raw, &fn)
	return fn, err == nil
}

func (jql *Json2Sql) JsonRawSelectDetail(raw json.RawMessage) (SelectionFields, bool) {
	var v SelectionFields
	err := json.Unmarshal(raw, &v)
	return v, err == nil
}

func (jql *Json2Sql) JsonRawSelectCase(raw json.RawMessage) (Case, bool) {
	var v Case
	err := json.Unmarshal(raw, &v)
	return v, err == nil
}

func (jql *Json2Sql) JsonRawCaseDefauleValue(raw json.RawMessage) (CaseDefauleValue, bool) {
	var v CaseDefauleValue
	err := json.Unmarshal(raw, &v)
	return v, err == nil
}

func (jql *Json2Sql) MaskedQueryValue(query string) (string, []string) {
	re := regexp.MustCompile(`JQL_VALUE:(\d+)|JQL_VALUE:'([^']+)'`)
	matches := re.FindAllStringSubmatch(query, -1)

	var values []string
	for _, match := range matches {
		if len(match) == 3 {
			if match[1] != "" {
				values = append(values, match[1])
			} else {
				values = append(values, match[2])
			}
		}
	}

	replacedQuery := re.ReplaceAllStringFunc(query, func(s string) string {
		return "?"
	})

	return replacedQuery, values
}

func (jql *Json2Sql) rawValueExtractor(query string) string {
	replacedQuery := strings.ReplaceAll(query, "JQL_VALUE:", "")
	return replacedQuery
}

func (jql *Json2Sql) GenerateSelectFrom(selection ...json.RawMessage) string {
	sql := "SELECT"

	if jql.sqlJson.SelectFields == nil {
		sql += fmt.Sprintf(" * FROM %s ", jql.sqlJson.Table)
		return sql
	}

	if selection == nil {
		selection = *jql.sqlJson.SelectFields
	}

	if selection != nil {
		if len(selection) > 0 {
			var selectFields []string
			for _, selectField := range selection {
				var field string

				field, isStringField := jql.JsonRawString(selectField)
				if !isStringField {

					sqlSelectDetail, isSqlSelectDetailField := jql.JsonRawSelectDetail(selectField)
					if isSqlSelectDetailField {
						if sqlSelectDetail.Alias != nil {
							field = fmt.Sprintf("%s AS %s", sqlSelectDetail.Field, *sqlSelectDetail.Alias)
						} else {
							field = sqlSelectDetail.Field
						}

						if sqlSelectDetail.SubQuery != nil {
							jsonBytes, _ := json.Marshal(*sqlSelectDetail.SubQuery)
							jql := NewJson2Sql(jsonBytes)
							field = fmt.Sprintf("(%s) AS %s", jql.Build(), *sqlSelectDetail.Alias)
						}
					}

					sqlSelectCase, isSqlSelectCaseField := jql.JsonRawSelectCase(selectField)
					if isSqlSelectCaseField {
						if sqlSelectCase.When != nil {
							var alias = ""

							if sqlSelectCase.Alias != nil {
								alias = fmt.Sprintf("AS %s", *sqlSelectDetail.Alias)
							}

							var defaultValue = ""

							sqlDefaultValue, isSqlDefaultValue := jql.JsonRawCaseDefauleValue(sqlSelectCase.DefaultValue)
							if isSqlDefaultValue {
								if sqlDefaultValue.Datatype != nil {
									expectDataType := SQLDataTypeEnum(strings.ToUpper(string(*sqlDefaultValue.Datatype)))
									defaultValue = ExtractValueByDataType(expectDataType, sqlDefaultValue.Value, *sqlDefaultValue.IsStatic)
								} else {
									selectExpect, isSelectExpect := jql.JsonRawSelectDetail(sqlDefaultValue.Value)
									if isSelectExpect {
										if selectExpect.SubQuery != nil {
											jsonBytes, _ := json.Marshal(*selectExpect.SubQuery)
											jql := NewJson2Sql(jsonBytes)
											defaultValue = fmt.Sprintf("(%s)", jql.Build())
										}
									}
								}
							}

							field = "CASE " + jql.GenerateConditions(*sqlSelectCase.When...) + " ELSE " + defaultValue + " END " + alias
						}
					}
				}

				selectFields = append(selectFields, field)
			}
			sql += fmt.Sprintf(" %s FROM %s ", strings.Join(selectFields, ", "), jql.sqlJson.Table)
		} else {
			sql += fmt.Sprintf(" * FROM %s ", jql.sqlJson.Table)
		}
	}

	return sql
}

func (jql *Json2Sql) GenerateWhere() string {
	var sql = ""

	if jql.sqlJson.Conditions != nil {
		sql += " WHERE " + jql.GenerateConditions(*jql.sqlJson.Conditions...)
	}

	return sql
}

func (jql *Json2Sql) GenerateOrderBy() string {
	var sql = ""

	if jql.sqlJson.OrderBy != nil {
		if jql.sqlJson.OrderBy.Sort != nil {
			sql += fmt.Sprintf(" ORDER BY %s %s", strings.Join(jql.sqlJson.OrderBy.Fields, ", "), strings.ToUpper(*jql.sqlJson.OrderBy.Sort))
		} else {
			sql += fmt.Sprintf(" ORDER BY %s", strings.Join(jql.sqlJson.OrderBy.Fields, ", "))
		}
	}

	return sql
}

func (jql *Json2Sql) GenerateGroupBy() string {
	var sql = ""

	if jql.sqlJson.GroupBy != nil {
		sql += fmt.Sprintf(" GROUP BY %s", strings.Join(jql.sqlJson.GroupBy.Fields, ", "))
	}

	return sql
}

func (jql *Json2Sql) GenerateJoin() string {
	var joinStr []string

	if jql.sqlJson.Join != nil {

		for _, joinCondition := range *jql.sqlJson.Join {
			for left, right := range joinCondition.On {
				if joinCondition.Type != nil && strings.ToUpper(*joinCondition.Type) == "LEFT" {
					joinStr = append(joinStr, fmt.Sprintf("%s %s ON %s = %s", " LEFT JOIN", *joinCondition.Table, left, right))
				} else if joinCondition.Type != nil && strings.ToUpper(*joinCondition.Type) == "RIGHT" {
					joinStr = append(joinStr, fmt.Sprintf("%s %s ON %s = %s", " RIGHT JOIN", *joinCondition.Table, left, right))
				} else if joinCondition.Type != nil && strings.ToUpper(*joinCondition.Type) == "INNER" {
					joinStr = append(joinStr, fmt.Sprintf("%s %s ON %s = %s", " INNER JOIN", *joinCondition.Table, left, right))
				} else {
					joinStr = append(joinStr, fmt.Sprintf("%s %s ON %s = %s", " JOIN", *joinCondition.Table, left, right))
				}
			}
		}
	}
	return strings.Join(joinStr, " ")
}

func (jql *Json2Sql) GenerateHaving() string {
	var sql = ""

	if jql.sqlJson.Having != nil {
		sql += " HAVING " + jql.GenerateConditions(*jql.sqlJson.Having...)
	}

	return sql
}

func (jql *Json2Sql) GenerateConditions(conditions ...Condition) string {
	var conditionsStr []string

	if conditions == nil {
		conditions = *jql.sqlJson.Conditions
	}

	// var valueRange ValueRange
	// var value string
	var clause string

	for _, condition := range conditions {
		var isStatic bool = false

		if condition.IsStatic != nil && *condition.IsStatic {
			isStatic = true
		}

		if condition.Clause != nil {
			strClause, isStringClause := jql.JsonRawString(condition.Clause)
			if isStringClause {
				clause = strClause
			}

			fnClause, isSqlFuncClause := jql.JsonRawSqlFunc(condition.Clause)
			if isSqlFuncClause {
				params := ArrayConversionToStringExpression(fnClause.SqlFunc.Params, isStatic, *fnClause.SqlFunc.IsField)
				clause = fmt.Sprintf("%s(%s)", strings.ToUpper(fnClause.SqlFunc.Name), params)
			}
		}

		// Check composite
		if condition.Composite != nil {
			compositeStr := jql.GenerateConditions(*condition.Composite...)
			if condition.Operand != nil {
				conditionsStr = append(conditionsStr, fmt.Sprintf("%s (%s)", strings.ToUpper(*condition.Operand), fmt.Sprint(compositeStr)))
			} else {
				conditionsStr = append(conditionsStr, fmt.Sprintf("(%s)", fmt.Sprint(compositeStr)))
			}
		} else {
			var expression = ""
			if condition.Datatype != nil {
				expression = GetSqlExpression(condition.Operator, *condition.Datatype, isStatic, condition.Value)
			} else {
				selectSub, isSelectSub := jql.JsonRawSelectDetail(condition.Value)
				if isSelectSub {
					if selectSub.SubQuery != nil {
						jsonBytes, _ := json.Marshal(*selectSub.SubQuery)
						jql := NewJson2Sql(jsonBytes)
						expression = string(condition.Operator) + " " + fmt.Sprintf("(%s)", jql.Build())
					}
				}
			}

			if condition.Expectation == nil {
				if condition.Operand != nil {
					conditionsStr = append(conditionsStr, fmt.Sprintf("%s %s %s", strings.ToUpper(*condition.Operand), clause, expression))
				} else {
					conditionsStr = append(conditionsStr, fmt.Sprintf("%s %s", clause, expression))
				}
			} else {
				// If condition selection case then
				var expect string

				if condition.Expectation.Datatype != nil {
					expectDataType := SQLDataTypeEnum(strings.ToUpper(string(*condition.Expectation.Datatype)))
					expect = ExtractValueByDataType(expectDataType, condition.Expectation.Value, *condition.Expectation.IsStatic)
				} else {
					selectExpect, isSelectExpect := jql.JsonRawSelectDetail(condition.Expectation.Value)
					if isSelectExpect {
						if selectExpect.SubQuery != nil {
							jsonBytes, _ := json.Marshal(*selectExpect.SubQuery)
							jql := NewJson2Sql(jsonBytes)
							expect = fmt.Sprintf("(%s)", jql.Build())
						}
					}
				}

				conditionsStr = append(conditionsStr, fmt.Sprintf("WHEN %s %s THEN %s", clause, expression, expect))
			}
		}
	}

	return strings.Join(conditionsStr, " ")
}

func (jql *Json2Sql) rawBuild() string {
	sql := jql.GenerateSelectFrom() + jql.GenerateJoin() + jql.GenerateWhere() + jql.GenerateGroupBy() + jql.GenerateHaving() + jql.GenerateOrderBy()

	return cleanSpaces(sql)
}

func (jql *Json2Sql) Build() string {
	sql := jql.GenerateSelectFrom() + jql.GenerateJoin() + jql.GenerateWhere() + jql.GenerateGroupBy() + jql.GenerateHaving() + jql.GenerateOrderBy()
	sqlCleanValue := jql.rawValueExtractor(sql)

	return cleanSpaces(sqlCleanValue)
}

func (jql *Json2Sql) Generate() (string, []string, error) {
	sql := jql.rawBuild()
	newQuery, values := jql.MaskedQueryValue(sql)

	return newQuery, values, nil
}
