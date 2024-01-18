package gojson2sql

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstructor(t *testing.T) {
	var njql, _ = NewJson2Sql([]byte(`{"table":"test"}`))
	assert.Equal(t, "test", njql.sqlJson.Table)
}

func TestConstructor_Fail(t *testing.T) {
	var _, err = NewJson2Sql([]byte(`{"table":"test"`))
	fmt.Println(err)
	assert.NotNil(t, err)
}

func TestRawJson_OK(t *testing.T) {
	var expected = "hello"

	raw := json.RawMessage(`"hello"`)

	jql := Json2Sql{}
	res, _ := jql.JsonRawString(raw)

	assert.Equal(t, expected, res)
}

func TestRawJson_Error(t *testing.T) {
	raw := json.RawMessage(`{"hello":"world"}`)

	jql := Json2Sql{}
	res, _ := jql.JsonRawString(raw)

	assert.Empty(t, res)
}

func TestMaskedQueryValue(t *testing.T) {
	strTest := `SELECT * FROM test WHERE v1 = JQL_VALUE:'A' AND v2 = JQL_VALUE:'B'`

	strExpected := `SELECT * FROM test WHERE v1 = ? AND v2 = ?`
	arrStrLengthExpected := 2

	jql := Json2Sql{}
	str, astr := jql.MaskedQueryValue(strTest)

	assert.Equal(t, strExpected, str)
	assert.Equal(t, arrStrLengthExpected, len(astr))
}

func TestGenerateSelectFrom(t *testing.T) {
	strTest := `{"table":"test"}`
	strExpected := `SELECT * FROM test `

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateSelectFrom()

	assert.Equal(t, strExpected, str)
}

func TestGenerateSelectFrom_Selection(t *testing.T) {
	strTest := `{
		"table": "test",
		"selectFields":
		[
			"a",
			"b",
			{"field":"c"},
			{"field":"d","alias":"e"},
			{
				"alias":"user",
				"subquery": {
					"table": "users",
					"selectFields": [
						"id",
						"name"
					],
					"conditions": [
						{
							"datatype": "NUMBER",
							"clause": "id",
							"operator": "=",
							"value": 1
						}
					],
					"limit": 1
				}
			},
			{
				"alias":"user",
				"subquery": {
					"table": "users",
					"conditions": [
						{
							"datatype": "NUMBER",
							"clause": "id",
							"operator": "=",
							"value": 1
						}
					],
					"limit": 1
				}
			},
			{
				"alias":"user",
				"subquery": {
					"table": "users",
					"selectFields": [],
					"conditions": [
						{
							"datatype": "NUMBER",
							"clause": "id",
							"operator": "=",
							"value": 1
						}
					],
					"limit": 1
				}
			}
		]
	}`
	strExpected := `SELECT a, b, c, d AS e, (SELECT id, name FROM users WHERE id = 1) AS user, (SELECT * FROM users WHERE id = 1) AS user, (SELECT * FROM users WHERE id = 1) AS user FROM test`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateSelectFrom()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestGenerateSelectFrom_CaseWhenThen(t *testing.T) {
	strTest := `{
		"table": "test",
		"selectFields":
		[
			{
				"when": [
					{
						"clause": "a",
						"datatype": "number",
						"isStatic": true,
						"operator": ">",
						"value": 100,
						"expectation": {
							"datatype": "BOOLEAN",
							"isStatic": true,
							"value": true
						}
					},
					{
						"clause": "b",
						"datatype": "number",
						"isStatic": true,
						"operator": ">",
						"value": 200,
						"expectation": {
							"datatype": "NUMBER",
							"isStatic": true,
							"value": 10
						}
					},
					{
						"clause": "c",
						"datatype": "function",
						"isStatic": true,
						"operator": ">",
						"value": {
							"sqlFunc": {
								"name": "count",
								"isField": true,
								"params": ["a.field"]
							}
						},
						"expectation": {
							"datatype": "function",
							"isStatic": true,
							"value": {
								"sqlFunc": {
									"name": "sum",
									"params": [1000]
								}
							}
						}
					},
					{
						"clause": "c",
						"datatype": "string",
						"isStatic": true,
						"operator": ">",
						"value": "A",
						"expectation": {
							"value": {
								"subquery": {
									"table": "users",
									"selectFields": [],
									"conditions": [
										{
											"datatype": "NUMBER",
											"clause": "id",
											"operator": "=",
											"value": 1
										}
									],
									"limit": 1
								}
							}
						}
					}
				],
				"defaultValue": {
					"datatype": "STRING",
					"isStatic": true,
					"value": "NICE"
				},
				"alias": "field_alias"
			}
		]
	}`
	strExpected := `SELECT CASE WHEN a > 100 THEN true WHEN b > 200 THEN 10 WHEN c > count(a.field) THEN sum(1000) WHEN c > 'A' THEN (SELECT * FROM users WHERE id = 1) ELSE 'NICE' END AS field_alias FROM test`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateSelectFrom()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestGenerateSelectFrom_CaseDefaultValueSub(t *testing.T) {
	strTest := `{
		"table": "test",
		"selectFields":
		[
			{
				"when": [
					{
						"clause": "a",
						"datatype": "number",
						"isStatic": true,
						"operator": ">",
						"value": 100,
						"expectation": {
							"datatype": "BOOLEAN",
							"isStatic": true,
							"value": true
						}
					}
				],
				"defaultValue": {
					"isStatic": true,
					"value": {
						"subquery": {
							"table": "users",
							"selectFields": [],
							"conditions": [
								{
									"datatype": "NUMBER",
									"clause": "id",
									"operator": "=",
									"value": 1
								}
							],
							"limit": 1
						}
					}
				},
				"alias": "field_alias"
			}
		]
	}`
	strExpected := `SELECT CASE WHEN a > 100 THEN true ELSE (SELECT * FROM users WHERE id = 1) END AS field_alias FROM test`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateSelectFrom()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestSqlLikeAndBlankDatatype(t *testing.T) {
	var sqlTest = `
		{
			"conditions": [
				{
					"clause": "a",
					"datatype": "STRING",
					"operator": "LIKE",
					"value": "%lorem ipsum%"
				},
				{
					"clause": "b",
					"datatype": "",
					"operator": "=",
					"value": "lorem ipsum"
				}
			]
		}
	`

	strExpected := `WHERE a LIKE JQL_VALUE:'%lorem ipsum%' b =`
	jql, _ := NewJson2Sql([]byte(sqlTest))
	str := jql.GenerateWhere()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestSqlLikeWithOperand(t *testing.T) {
	var sqlTest = `
		{
			"conditions": [
				{
					"clause": "b",
					"datatype": "STRING",
					"operator": "=",
					"value": "lorem ipsum"
				},
				{
					"operand": "and",
					"clause": "a",
					"datatype": "STRING",
					"operator": "ILIKE",
					"value": "%lorem ipsum%"
				}
			]
		}
	`

	strExpected := `WHERE b = JQL_VALUE:'lorem ipsum' AND a ILIKE JQL_VALUE:'%lorem ipsum%'`
	jql, _ := NewJson2Sql([]byte(sqlTest))
	str := jql.GenerateWhere()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestBetweenWithOperand(t *testing.T) {
	var sqlTest = `
		{
			"conditions": [
				{
					"clause": "a",
					"operand": "and",
					"datatype": "NUMBER",
					"operator": "BETWEEN",
					"value": {
						"from": 1,
						"to": 2
					}
				}
			]
		}
	`

	strExpected := `WHERE AND a BETWEEN JQL_VALUE:1 AND JQL_VALUE:2`
	jql, _ := NewJson2Sql([]byte(sqlTest))
	str := jql.GenerateWhere()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestCompositeWithoutOperand(t *testing.T) {
	var sqlTest = `
		{
			"conditions": [
				{
					"composite": [
						{
							"clause": "a",
							"datatype": "string",
							"operator": "=",
							"value": "b"
						}
					]
				}
			]
		}
	`

	strExpected := `WHERE (a = JQL_VALUE:'b')`
	jql, _ := NewJson2Sql([]byte(sqlTest))
	str := jql.GenerateWhere()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestGenerateOrderBy(t *testing.T) {
	strTest := `{"orderBy": {"fields":["a.b"]}}`
	strExpected := `ORDER BY a.b`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateOrderBy()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateOrderBy_WithSort(t *testing.T) {
	strTest := `{"orderBy": {"fields":["a.b"],"sort":"ASC"}}`
	strExpected := `ORDER BY a.b ASC`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateOrderBy()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateGroupBy(t *testing.T) {
	strTest := `{"groupBy": {"fields":["a"]}}`
	strExpected := `GROUP BY a`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateGroupBy()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateJoin_JOIN(t *testing.T) {
	strTest := `{
		"join": [
			{
				"table":"t2",
				"type":"join",
				"on":{"a":"b"} 
			},
			{
				"table":"t1",
				"type":"join",
				"on":{"b":"a"} 
			}
		]
	}`
	strExpected := `JOIN t2 ON a = b  JOIN t1 ON b = a`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateJoin()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateJoin_INNER_JOIN(t *testing.T) {
	strTest := `{
		"join": [
			{
				"table":"t2",
				"type":"inner",
				"on":{"a":"b"} 
			},
			{
				"table":"t1",
				"type":"inner",
				"on":{"b":"a"} 
			}
		]
	}`
	strExpected := `INNER JOIN t2 ON a = b  INNER JOIN t1 ON b = a`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateJoin()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateJoin_LEFT_JOIN(t *testing.T) {
	strTest := `{
		"join": [
			{
				"table":"t2",
				"type":"left",
				"on":{"a":"b"} 
			},
			{
				"table":"t1",
				"type":"left",
				"on":{"b":"a"} 
			}
		]
	}`
	strExpected := `LEFT JOIN t2 ON a = b  LEFT JOIN t1 ON b = a`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateJoin()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateJoin_RIGHT_JOIN(t *testing.T) {
	strTest := `{
		"join": [
			{
				"table":"t2",
				"type":"right",
				"on":{"a":"b"} 
			},
			{
				"table":"t1",
				"type":"right",
				"on":{"b":"a"} 
			}
		]
	}`
	strExpected := `RIGHT JOIN t2 ON a = b  RIGHT JOIN t1 ON b = a`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateJoin()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateHaving(t *testing.T) {
	strTest := `
	{
		"having": [
			{
				"clause": {
					"sqlFunc": {
						"name": "count",
						"isField": true,
						"params": ["a"]
					}
				},
				"operator": ">",
				"datatype": "number",
				"value": 100
			},
			{
				"clause": {
					"sqlFunc": {
						"name": "sum",
						"isField": true,
						"params": ["a"]
					}
				},
				"operand": "AND",
				"operator": ">",
				"datatype": "number",
				"value": 100
			}
		]
	}`
	strExpected := `HAVING COUNT(a) > JQL_VALUE:100 AND SUM(a) > JQL_VALUE:100`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateHaving()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateWhere(t *testing.T) {
	strTest := `
	{
		"conditions": [
			{
				"clause": "a",
				"operator": "=",
				"datatype": "STRING",
				"value": "b"
			},
			{
				"clause": "users.birthdate",
				"operator": "between",
				"datatype": "STRING",
				"value": {
					"from": "2015-01-01",
					"to": "2021-01-01"
				}
			},
			{
				"operand": "and",
				"clause": "c",
				"operator": "=",
				"datatype": "STRING",
				"value": "d"
			}
		]
	}`
	strExpected := `WHERE a = JQL_VALUE:'b' users.birthdate BETWEEN JQL_VALUE:'2015-01-01' AND JQL_VALUE:'2021-01-01' AND c = JQL_VALUE:'d'`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateWhere()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateConditions(t *testing.T) {
	strTest := `
	{
		"conditions": [
			{
				"clause": "users.id",
				"operator": "=",
				"datatype": "NUMBER",
				"value": 1
			},
			{
				"operand": "and",
				"clause": "gender.gender_name",
				"datatype": "STRING",
				"operator": "=",
				"value": "female"
			},
			{
				"operand": "and",
				"datatype": "FUNCTION",
				"operator": ">",
				"clause": "transaction.total",
				"value": {
					"sqlFunc": {
						"name": "sum",
						"params": [100]
					}
				}
			},
			{
				"operand": "or",
				"composite": [
					{
						"clause": "users.birthdate",
						"operator": "between",
						"datatype": "STRING",
						"value": {
							"from": "2015-01-01",
							"to": "2021-01-01"
						}
					},
					{
						"operand": "and",
						"clause": "users.status",
						"datatype": "STRING",
						"operator": "=",
						"value": "active"
					}
				]
			}
		]
	}
	`
	strExpected := `users.id = JQL_VALUE:1 AND gender.gender_name = JQL_VALUE:'female' AND transaction.total > sum(JQL_VALUE:100) OR (users.birthdate BETWEEN JQL_VALUE:'2015-01-01' AND JQL_VALUE:'2021-01-01' AND users.status = JQL_VALUE:'active')`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateConditions()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateConditions_SubQuery(t *testing.T) {
	strTest := `
	{
		"conditions": [
			{
				"clause": "users.id",
				"operator": "=",
				"value": {
					"subquery": {
						"table": "users",
						"selectFields": [],
						"conditions": [
							{
								"datatype": "NUMBER",
								"clause": "id",
								"operator": "=",
								"value": 1
							}
						],
						"limit": 1
					}
				}
			}
		]
	}
	`
	strExpected := `users.id = (SELECT * FROM users WHERE id = 1)`

	jql, _ := NewJson2Sql([]byte(strTest))
	str := jql.GenerateConditions()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestBuildJsonToSql(t *testing.T) {
	jsonData := `
		{
			"table": "users",
			"selectFields": [
				"users.id",
				"users.name",
				"users.birthdate",
				"gender.gender_name"
			],
			"join": [
				{
					"table": "role",
					"type": "left",
					"on": {
						"users.id": "role.user_id"
					}
				},
				{
					"table": "gender",
					"type": "left",
					"on": {
						"users.gender_id": "gender.id"
					}
				}
			],
			"conditions": [
				{
					"clause": "users.id",
					"operator": "=",
					"datatype": "number",
					"value": 1
				},
				{
					"operand": "and",
					"datatype": "string",
					"clause": "gender.gender_name",
					"operator": "=",
					"value": "female"
				},
				{
					"operand": "and",
					"datatype": "function",
					"clause": "transaction.total",
					"operator": ">",
					"value": {
						"sqlFunc": {
							"name": "sum",
							"params": [100]
						}
					}
				},
				{
					"operand": "or",
					"composite": [
						{
							"clause": "users.birthdate",
							"datatype": "string",
							"operator": "between",
							"value": {
								"from": "2015-01-01",
								"to": "2021-01-01"
							}
						},
						{
							"operand": "and",
							"datatype": "string",
							"clause": "users.status",
							"operator": "=",
							"value": "active"
						}
					]
				}
			],
			"groupBy": {
				"fields": ["users.id"]
			},
			"having": [
				{
					"clause": {
						"sqlFunc": {
							"name": "count",
							"isField": true,
							"params": ["users.id"]
						}
					},
					"datatype": "number",
					"operator": ">",
					"value": 10
				}
			],
			"orderBy": {
				"fields": ["users.id", "gender.id"],
				"sort": "asc"
			},
			"limit": 1
		}
	`

	jql, _ := NewJson2Sql([]byte(jsonData))
	sql := jql.Build()

	strExpectation := "SELECT users.id, users.name, users.birthdate, gender.gender_name FROM users LEFT JOIN role ON users.id = role.user_id LEFT JOIN gender ON users.gender_id = gender.id WHERE users.id = 1 AND gender.gender_name = 'female' AND transaction.total > sum(100) OR (users.birthdate BETWEEN '2015-01-01' AND '2021-01-01' AND users.status = 'active') GROUP BY users.id HAVING COUNT(users.id) > 10 ORDER BY users.id, gender.id ASC"
	assert.Equal(t, strExpectation, sql)
}

func TestGenerateJsonToSql(t *testing.T) {
	jsonData := `
		{
			"table": "users",
			"selectFields": [
				"users.id",
				"users.name",
				"users.birthdate",
				"gender.gender_name"
			],
			"join": [
				{
					"table": "role",
					"type": "left",
					"on": {
						"users.id": "role.user_id"
					}
				},
				{
					"table": "gender",
					"type": "left",
					"on": {
						"users.gender_id": "gender.id"
					}
				}
			],
			"conditions": [
				{
					"datatype": "string",
					"clause": "gender.gender_name",
					"operator": "=",
					"value": "female"
				},
				{
					"clause": "users.id",
					"datatype": "number",
					"operator": "=",
					"value": 1
				},
				{
					"operand": "and",
					"datatype": "function",
					"clause": "transaction.total",
					"operator": ">",
					"value": {
						"sqlFunc": {
							"name": "sum",
							"params": [100]
						}
					}
				},
				{
					"operand": "or",
					"composite": [
						{
							"clause": "users.birthdate",
							"datatype": "string",
							"operator": "between",
							"value": {
								"from": "2015-01-01",
								"to": "2021-01-01"
							}
						},
						{
							"operand": "and",
							"datatype": "string",
							"clause": "users.status",
							"operator": "=",
							"value": "active"
						}
					]
				}
			],
			"groupBy": {
				"fields": ["users.id"]
			},
			"having": [
				{
					"clause": {
						"sqlFunc": {
							"name": "count",
							"isField": true,
							"params": ["users.id"]
						}
					},
					"datatype": "number",
					"operator": ">",
					"value": 10
				}
			],
			"orderBy": {
				"fields": ["users.id", "gender.id"],
				"sort": "asc"
			},
			"limit": 1
		}
	`

	jql, _ := NewJson2Sql([]byte(jsonData))
	sql, filter, _ := jql.Generate()

	strExpectation := "SELECT users.id, users.name, users.birthdate, gender.gender_name FROM users LEFT JOIN role ON users.id = role.user_id LEFT JOIN gender ON users.gender_id = gender.id WHERE gender.gender_name = ? users.id = ? AND transaction.total > sum(?) OR (users.birthdate BETWEEN ? AND ? AND users.status = ?) GROUP BY users.id HAVING COUNT(users.id) > ? ORDER BY users.id, gender.id ASC"
	assert.Equal(t, strExpectation, sql)
	assert.Equal(t, []string{"female", "1", "100", "2015-01-01", "2021-01-01", "active", "10"}, filter)
}
