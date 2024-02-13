package gojson2sql

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/goccy/go-json"
)

type Json2SqlConf struct {
	WithUnion              bool
	WithSanitizedInjection bool
}
type Json2Sql struct {
	sqlJson            *SQLJson
	sqlJsonSelectUnion *[]SQLJson
	config             *Json2SqlConf
}

func NewJson2Sql(jsonData []byte, conf *Json2SqlConf) (*Json2Sql, error) {
	var sqlJson *SQLJson
	var sqlJsonUnion *[]SQLJson

	if conf != nil && conf.WithUnion {

		err := json.Unmarshal(jsonData, &sqlJsonUnion)
		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
	} else {

		err := json.Unmarshal(jsonData, &sqlJson)

		if err != nil {
			return nil, fmt.Errorf("error: %s", err)
		}
	}

	return &Json2Sql{
		sqlJson:            sqlJson,
		sqlJsonSelectUnion: sqlJsonUnion,
		config:             conf,
	}, nil
}

func cleanSpaces(input string) string {
	words := strings.Fields(input)
	result := strings.Join(words, " ")
	return result
}

func cleanWhereCond(input string) string {
	re := regexp.MustCompile(`(?i)where\s*and|where or`)
	cleanedInput := re.ReplaceAllString(input, "WHERE")
	return cleanedInput
}

func sanitizeInjection(input string) string {
	re := regexp.MustCompile(`(?i)[;]|--|drop\s*table|@@\s*version|insert\s*into|if\s*\(|sleep\s*\(|"|\/\*|\*\/|\\0|\\'|\\"|\\b|\\n|\\r|\\t|\\Z|\\\\|\\%|\\_`)
	cleanedInput := re.ReplaceAllString(input, "")
	return cleanedInput
}

func isValidSQL(input string) bool {
	return sanitizeInjection(input) == input
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

func (jql *Json2Sql) JsonRawLimitOffsetValue(raw json.RawMessage) (LimitOffsetValue, bool) {
	var v LimitOffsetValue
	err := json.Unmarshal(raw, &v)
	return v, err == nil
}

func (jql *Json2Sql) isStringNumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func (jql *Json2Sql) MaskedQueryValue(query string) (string, []interface{}) {
	sRegex := fmt.Sprintf(`%s'([^\s]+)'%s|%s([^\s]+)%s`, JQL_FLAG_OPEN, JQL_FLAG_CLOSE, JQL_FLAG_OPEN, JQL_FLAG_CLOSE)
	re := regexp.MustCompile(sRegex)
	matches := re.FindAllStringSubmatch(query, -1)

	var values []interface{}
	for _, match := range matches {
		for i := 1; i < len(match); i += 2 {
			if match[i] != "" {
				values = append(values, match[i])
			} else if match[i+1] != "" {
				var m = match[i+1]
				if strings.ToLower(m) == "true" || strings.ToLower(m) == "false" {
					booleanValue, _ := strconv.ParseBool(m)
					values = append(values, booleanValue)
				} else if jql.isStringNumeric(m) {
					intValue, _ := strconv.ParseFloat(m, 64)
					values = append(values, intValue)
				} else {
					values = append(values, m)
				}

			}
		}
	}

	replacedQuery := re.ReplaceAllStringFunc(query, func(s string) string {
		return "?"
	})

	return replacedQuery, values
}

func (jql *Json2Sql) rawValueExtractor(query string) string {
	replacedQuery := strings.ReplaceAll(query, JQL_FLAG_OPEN, "")
	replacedQuery = strings.ReplaceAll(replacedQuery, JQL_FLAG_CLOSE, "")

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
				// var field string

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

							jql, _ := NewJson2Sql(jsonBytes, &Json2SqlConf{WithSanitizedInjection: jql.config.WithSanitizedInjection})

							field = fmt.Sprintf("(%s) AS %s", jql.rawBuild(), *sqlSelectDetail.Alias)

						}

						if sqlSelectDetail.AddFunction != nil {
							paramFunc := ArrayConversionToStringExpression(sqlSelectDetail.AddFunction.SqlFunc.Params, *sqlSelectDetail.AddFunction.SqlFunc.IsField, *sqlSelectDetail.AddFunction.SqlFunc.IsField)
							field = fmt.Sprintf("%s(%s) AS %s", strings.ToUpper(sqlSelectDetail.AddFunction.SqlFunc.Name), paramFunc, *sqlSelectDetail.Alias)
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
											jql, _ := NewJson2Sql(jsonBytes, &Json2SqlConf{WithSanitizedInjection: jql.config.WithSanitizedInjection})
											defaultValue = fmt.Sprintf("(%s)", jql.rawBuild())
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
		sql += cleanWhereCond(" WHERE " + jql.GenerateConditions(*jql.sqlJson.Conditions...))
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
						jql, _ := NewJson2Sql(jsonBytes, &Json2SqlConf{WithSanitizedInjection: jql.config.WithSanitizedInjection})
						expression = string(condition.Operator) + " " + fmt.Sprintf("(%s)", jql.rawBuild())
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
							jql, _ := NewJson2Sql(jsonBytes, &Json2SqlConf{WithSanitizedInjection: jql.config.WithSanitizedInjection})
							expect = fmt.Sprintf("(%s)", jql.rawBuild())
						}
					}
				}

				conditionsStr = append(conditionsStr, fmt.Sprintf("WHEN %s %s THEN %s", clause, expression, expect))
			}
		}
	}

	return strings.Join(conditionsStr, " ")
}

func (jql *Json2Sql) GenerateLimit() string {
	var sql = ""
	if jql.sqlJson.Limit != nil {
		v, b := jql.JsonRawLimitOffsetValue(*jql.sqlJson.Limit)
		if b {
			if v.IsStatic {
				sql += fmt.Sprintf(" LIMIT %s", strconv.Itoa(v.Value))
			} else {
				sql += fmt.Sprintf(" LIMIT %s%s%s", JQL_FLAG_OPEN, strconv.Itoa(v.Value), JQL_FLAG_CLOSE)
			}
		} else {
			sql += fmt.Sprintf(" LIMIT %s", *jql.sqlJson.Limit)
		}
	}

	return sql
}

func (jql *Json2Sql) GenerateOffset() string {
	var sql = ""
	if jql.sqlJson.Offset != nil {
		v, b := jql.JsonRawLimitOffsetValue(*jql.sqlJson.Offset)
		if b {
			if v.IsStatic {
				sql += fmt.Sprintf(" OFFSET %s", strconv.Itoa(v.Value))
			} else {
				sql += fmt.Sprintf(" OFFSET %s%s%s", JQL_FLAG_OPEN, strconv.Itoa(v.Value), JQL_FLAG_CLOSE)
			}
		} else {
			sql += fmt.Sprintf(" OFFSET %s", *jql.sqlJson.Offset)
		}
	}

	return sql
}

func (jql *Json2Sql) concateQueryString() string {
	return jql.GenerateSelectFrom() + jql.GenerateJoin() + jql.GenerateWhere() + jql.GenerateGroupBy() + jql.GenerateHaving() + jql.GenerateOrderBy() + jql.GenerateLimit() + jql.GenerateOffset()
}

func (jql *Json2Sql) rawBuild() string {
	return cleanSpaces(jql.concateQueryString())
}

func (jql *Json2Sql) Build() string {
	sqlCleanValue := jql.rawValueExtractor(jql.concateQueryString())

	if jql.config != nil && jql.config.WithSanitizedInjection && !isValidSQL(sqlCleanValue) {
		return "Invalid sql string you've got sanitized SQL string"
	}

	return cleanSpaces(sqlCleanValue)
}

func (jql *Json2Sql) Generate() (string, []interface{}, error) {
	sql := jql.rawBuild()

	if jql.config != nil && jql.config.WithSanitizedInjection && !isValidSQL(sql) {
		return "", nil, fmt.Errorf("error: %s", "Invalid sql string you've got sanitized SQL string")
	}

	newQuery, values := jql.MaskedQueryValue(sql)
	return newQuery, values, nil
}

func (jql *Json2Sql) buildRawUnion() string {
	var sql string
	var sqlUnion []string

	if jql.sqlJsonSelectUnion != nil {
		for _, v := range *jql.sqlJsonSelectUnion {
			currentV := v
			jql.sqlJson = &currentV
			strBuild := jql.rawBuild()
			sqlUnion = append(sqlUnion, strBuild)
		}
	}

	// jql.sqlJson = nil
	sql = strings.Join(sqlUnion, " UNION ")

	return sql
}

func (jql *Json2Sql) BuildUnion() string {
	sqlCleanValue := jql.rawValueExtractor(jql.buildRawUnion())

	if jql.config != nil && jql.config.WithSanitizedInjection && !isValidSQL(sqlCleanValue) {
		return "Invalid sql string you've got sanitized SQL string"
	}

	return cleanSpaces(sqlCleanValue)
}

func (jql *Json2Sql) GenerateUnion() (string, []interface{}, error) {
	sql := jql.buildRawUnion()

	if jql.config != nil && jql.config.WithSanitizedInjection && !isValidSQL(sql) {
		return "", nil, fmt.Errorf("error: %s", "Invalid sql string you've got sanitized SQL string")
	}

	newQuery, values := jql.MaskedQueryValue(sql)

	return newQuery, values, nil
}
