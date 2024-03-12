package gojson2sql

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
)

func TestConstructor(t *testing.T) {
	var njql, _ = NewJson2Sql([]byte(`{"table":"test"}`), &Json2SqlConf{})
	assert.Equal(t, "test", njql.sqlJson.Table)
}

func TestConstructor_Fail(t *testing.T) {
	var _, err = NewJson2Sql([]byte(`{"table":"test"`), &Json2SqlConf{})
	assert.NotNil(t, err)
}

func TestConstructor_Fail_Union(t *testing.T) {
	var _, err = NewJson2Sql([]byte(`[{"table":"test"`), &Json2SqlConf{WithUnion: true})
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
	strTest := `SELECT * FROM test WHERE v1 = JQL_VALUE:'A':END_JQL_VALUE AND v2 = JQL_VALUE:'B':END_JQL_VALUE AND v3 = JQL_VALUE:true:END_JQL_VALUE AND v4 = JQL_VALUE:CURRENT_TIME:END_JQL_VALUE AND v5 = JQL_VALUE:123:END_JQL_VALUE`

	strExpected := `SELECT * FROM test WHERE v1 = ? AND v2 = ? AND v3 = ? AND v4 = ? AND v5 = ?`
	arrStrLengthExpected := 5

	jql := Json2Sql{}
	str, astr := jql.MaskedQueryValue(strTest)

	assert.Equal(t, strExpected, str)
	assert.Equal(t, arrStrLengthExpected, len(astr))

	// Assert Datatype
	assert.Equal(t, "string", reflect.TypeOf(astr[0]).String())
	assert.Equal(t, "string", reflect.TypeOf(astr[1]).String())
	assert.Equal(t, "bool", reflect.TypeOf(astr[2]).String())
	assert.Equal(t, "string", reflect.TypeOf(astr[3]).String())
	assert.Equal(t, "float64", reflect.TypeOf(astr[4]).String())
}

func TestGenerateSelectFrom(t *testing.T) {
	strTest := `{"table":"test"}`
	strExpected := `SELECT * FROM test `

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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
	strExpected := `SELECT a, b, c, d AS e, (SELECT id, name FROM users WHERE id = JQL_VALUE:1:END_JQL_VALUE LIMIT 1) AS user, (SELECT * FROM users WHERE id = JQL_VALUE:1:END_JQL_VALUE LIMIT 1) AS user FROM test`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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
	strExpected := `SELECT CASE WHEN a > 100 THEN true WHEN b > 200 THEN 10 WHEN c > count(a.field) THEN sum(1000) WHEN c > 'A' THEN (SELECT * FROM users WHERE id = JQL_VALUE:1:END_JQL_VALUE LIMIT 1) ELSE 'NICE' END AS field_alias FROM test`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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
	strExpected := `SELECT CASE WHEN a > 100 THEN true ELSE (SELECT * FROM users WHERE id = JQL_VALUE:1:END_JQL_VALUE LIMIT 1) END AS field_alias FROM test`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateSelectFrom()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestGenerateSelectFrom_SqlFunc(t *testing.T) {
	strTest := `{
		"table": "test",
		"selectFields": [
			{
				"alias": "a_test",
				"addFunction": {
					"sqlFunc": {
						"name": "count",
						"isField": true,
						"params": ["a"]
					}
				}
			}
		]
	}`
	strExpected := `SELECT COUNT(a) AS a_test FROM test`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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

	strExpected := `WHERE a LIKE JQL_VALUE:'%lorem ipsum%':END_JQL_VALUE b =`
	jql, _ := NewJson2Sql([]byte(sqlTest), &Json2SqlConf{})
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

	strExpected := `WHERE b = JQL_VALUE:'lorem ipsum':END_JQL_VALUE AND a ILIKE JQL_VALUE:'%lorem ipsum%':END_JQL_VALUE`
	jql, _ := NewJson2Sql([]byte(sqlTest), &Json2SqlConf{})
	str := jql.GenerateWhere()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestBetweenWithOperand(t *testing.T) {
	var sqlTest = `
		{
			"conditions": [
				{
					"clause": "a",
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

	strExpected := `WHERE a BETWEEN JQL_VALUE:1:END_JQL_VALUE AND JQL_VALUE:2:END_JQL_VALUE`
	jql, _ := NewJson2Sql([]byte(sqlTest), &Json2SqlConf{})
	str := jql.GenerateWhere()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestWhereIs(t *testing.T) {
	var sqlTest = `
		{
			"conditions": [
				{
          "operand": "or",
					"clause": "a",
					"datatype": "STRING",
					"operator": "is not null",
					"value": null
				},
				{
          "operand": "and",
					"clause": "b",
					"datatype": "STRING",
					"operator": "is null",
					"value": null
				}
			]
		}
	`

	strExpected := `WHERE a IS NOT NULL AND b IS NULL`
	jql, _ := NewJson2Sql([]byte(sqlTest), &Json2SqlConf{})
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

	strExpected := `WHERE (a = JQL_VALUE:'b':END_JQL_VALUE)`
	jql, _ := NewJson2Sql([]byte(sqlTest), &Json2SqlConf{})
	str := jql.GenerateWhere()

	assert.Equal(t, strings.TrimSpace(strExpected), strings.TrimSpace(str))
}

func TestGenerateOrderBy(t *testing.T) {
	strTest := `{"orderBy": {"fields":["a.b"]}}`
	strExpected := `ORDER BY a.b`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateOrderBy()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateOrderBy_WithSort(t *testing.T) {
	strTest := `{"orderBy": {"fields":["a.b"],"sort":"ASC"}}`
	strExpected := `ORDER BY a.b ASC`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateOrderBy()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestGenerateGroupBy(t *testing.T) {
	strTest := `{"groupBy": {"fields":["a"]}}`
	strExpected := `GROUP BY a`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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
	strExpected := `HAVING COUNT(a) > JQL_VALUE:100:END_JQL_VALUE AND SUM(a) > JQL_VALUE:100:END_JQL_VALUE`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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
	strExpected := `WHERE a = JQL_VALUE:'b':END_JQL_VALUE users.birthdate BETWEEN JQL_VALUE:'2015-01-01':END_JQL_VALUE AND JQL_VALUE:'2021-01-01':END_JQL_VALUE AND c = JQL_VALUE:'d':END_JQL_VALUE`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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
	strExpected := `users.id = JQL_VALUE:1:END_JQL_VALUE AND gender.gender_name = JQL_VALUE:'female':END_JQL_VALUE AND transaction.total > sum(JQL_VALUE:100:END_JQL_VALUE) OR (users.birthdate BETWEEN JQL_VALUE:'2015-01-01':END_JQL_VALUE AND JQL_VALUE:'2021-01-01':END_JQL_VALUE AND users.status = JQL_VALUE:'active':END_JQL_VALUE)`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
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
	strExpected := `users.id = (SELECT * FROM users WHERE id = JQL_VALUE:1:END_JQL_VALUE LIMIT 1)`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateConditions()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestLimit_Static(t *testing.T) {
	strTest := `{
		"limit": {
			"isStatic": true,
			"value": 10
		}
	}`
	strExpected := `LIMIT 10`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateLimit()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestLimit_ToParam(t *testing.T) {
	strTest := `{
		"limit": {
			"value": 10
		}
	}`
	strExpected := `LIMIT JQL_VALUE:10:END_JQL_VALUE`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateLimit()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestOffset_Static(t *testing.T) {
	strTest := `{
		"offset": {
			"isStatic": true,
			"value": 10
		}
	}`
	strExpected := `OFFSET 10`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateOffset()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestOffset_ToParam(t *testing.T) {
	strTest := `{
		"offset": {
			"value": 10
		}
	}`
	strExpected := `OFFSET JQL_VALUE:10:END_JQL_VALUE`

	jql, _ := NewJson2Sql([]byte(strTest), &Json2SqlConf{})
	str := jql.GenerateOffset()

	assert.Equal(t, strExpected, strings.TrimSpace(str))
}

func TestBuildJsonToSql(t *testing.T) {
	jsonData := `
		{
			"table": "table_1",
			"selectFields": [
				"table_1.a",
				{
					"field": "table_1.b",
					"alias": "foo_bar"
				},
				{
					"field": "table_2.a",
					"alias": "baz",
					"subquery": {
						"table": "table_4",
						"selectFields": ["*"],
						"conditions": [
							{
								"datatype": "number",
								"clause": "a",
								"operator": "=",
								"value": 1
							}
						],
						"limit": 1
					}
				},
				"table_2.b",
				"table_3.a",
				"table_3.b"
			],
			"join": [
				{
					"table": "table_2",
					"type": "join",
					"on": {
						"table_2.a": "table_1.a"
					}
				},
				{
					"table": "table_3",
					"type": "left",
					"on": {
						"table_3.a": "table_2.a"
					}
				}
			],
			"conditions": [
				{
					"datatype": "string",
					"clause": "table_1.a",
					"operator": "=",
					"value": "foo"
				},
				{
					"operand": "and",
					"datatype": "boolean",
					"clause": "table_1.b",
					"operator": "=",
					"value": true
				},
				{
					"operand": "and",
					"datatype": "function",
					"clause": "table_2.a",
					"operator": ">",
					"value": {
						"sqlFunc": {
							"name": "sum",
							"params": [100]
						}
					}
				},
				{
					"operand": "and",
					"clause": "table_2.b",
					"operator": "=",
					"value": {
						"subquery": {
							"table": "table_4",
							"selectFields": ["*"],
							"conditions": [
								{
									"datatype": "number",
									"clause": "a",
									"operator": "=",
									"value": 1
								}
							],
							"limit": 1
						}
					}
				},
				{
					"operand": "or",
					"composite": [
						{
							"clause": "table_3.a",
							"datatype": "string",
							"operator": "between",
							"value": {
								"from": "2020-01-01",
								"to": "2023-01-01"
							}
						},
						{
							"operand": "and",
							"datatype": "number",
							"clause": "table_3.b",
							"operator": "=",
							"value": 2
						}
					]
				}
			],
			"groupBy": {
				"fields": ["table_1.a"]
			},
			"having": [
				{
					"clause": {
						"sqlFunc": {
							"name": "count",
							"isField": true,
							"params": ["table_2.a"]
						}
					},
					"datatype": "number",
					"operator": ">",
					"value": 10
				}
			],
			"orderBy": {
				"fields": ["table_1.a", "table_2.a"],
				"sort": "asc"
			},
			"limit": 1,
			"offset": 0
		}
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{})
	sql := jql.Build()

	strExpectation := "SELECT table_1.a, table_1.b AS foo_bar, (SELECT * FROM table_4 WHERE a = 1 LIMIT 1) AS baz, table_2.b, table_3.a, table_3.b FROM table_1 JOIN table_2 ON table_2.a = table_1.a LEFT JOIN table_3 ON table_3.a = table_2.a WHERE table_1.a = 'foo' AND table_1.b = true AND table_2.a > sum(100) AND table_2.b = (SELECT * FROM table_4 WHERE a = 1 LIMIT 1) OR (table_3.a BETWEEN '2020-01-01' AND '2023-01-01' AND table_3.b = 2) GROUP BY table_1.a HAVING COUNT(table_2.a) > 10 ORDER BY table_1.a, table_2.a ASC LIMIT 1 OFFSET 0"
	assert.Equal(t, strExpectation, sql)
}

func TestGenerateJsonToSql(t *testing.T) {
	jsonData := `
		{
			"table": "table_1",
			"selectFields": [
				"table_1.a",
				{
					"field": "table_1.b",
					"alias": "foo_bar"
				},
				{
					"field": "table_2.a",
					"alias": "baz",
					"subquery": {
						"table": "table_4",
						"selectFields": ["*"],
						"conditions": [
							{
								"datatype": "number",
								"clause": "a",
								"operator": "=",
								"value": 1
							}
						],
						"limit": 1
					}
				},
				"table_2.b",
				"table_3.a",
				"table_3.b"
			],
			"join": [
				{
					"table": "table_2",
					"type": "join",
					"on": {
						"table_2.a": "table_1.a"
					}
				},
				{
					"table": "table_3",
					"type": "left",
					"on": {
						"table_3.a": "table_2.a"
					}
				}
			],
			"conditions": [
				{
					"datatype": "string",
					"clause": "table_1.a",
					"operator": "=",
					"value": "foo"
				},
				{
					"operand": "and",
					"datatype": "boolean",
					"clause": "table_1.b",
					"operator": "=",
					"value": true
				},
				{
					"operand": "and",
					"datatype": "function",
					"clause": "table_2.a",
					"operator": ">",
					"value": {
						"sqlFunc": {
							"name": "sum",
							"params": [100]
						}
					}
				},
				{
					"operand": "and",
					"clause": "table_2.b",
					"operator": "=",
					"value": {
						"subquery": {
							"table": "table_4",
							"selectFields": ["*"],
							"conditions": [
								{
									"datatype": "number",
									"clause": "a",
									"operator": "=",
									"value": 1
								}
							],
							"limit": 1
						}
					}
				},
				{
					"operand": "or",
					"composite": [
						{
							"clause": "table_3.a",
							"datatype": "string",
							"operator": "between",
							"value": {
								"from": "2020-01-01",
								"to": "2023-01-01"
							}
						},
						{
							"operand": "and",
							"datatype": "string",
							"clause": "table_3.b",
							"operator": "=",
							"value": "2"
						}
					]
				}
			],
			"groupBy": {
				"fields": ["table_1.a"]
			},
			"having": [
				{
					"clause": {
						"sqlFunc": {
							"name": "count",
							"isField": true,
							"params": ["table_2.a"]
						}
					},
					"datatype": "number",
					"operator": ">",
					"value": 10
				}
			],
			"orderBy": {
				"fields": ["table_1.a", "table_2.a"],
				"sort": "asc"
			},
			"limit": 1,
			"offset": 0
		}
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{})
	sql, filter, _ := jql.Generate()

	strExpectation := "SELECT table_1.a, table_1.b AS foo_bar, (SELECT * FROM table_4 WHERE a = ? LIMIT 1) AS baz, table_2.b, table_3.a, table_3.b FROM table_1 JOIN table_2 ON table_2.a = table_1.a LEFT JOIN table_3 ON table_3.a = table_2.a WHERE table_1.a = ? AND table_1.b = ? AND table_2.a > sum(?) AND table_2.b = (SELECT * FROM table_4 WHERE a = ? LIMIT 1) OR (table_3.a BETWEEN ? AND ? AND table_3.b = ?) GROUP BY table_1.a HAVING COUNT(table_2.a) > ? ORDER BY table_1.a, table_2.a ASC LIMIT 1 OFFSET 0"
	assert.Equal(t, strExpectation, sql)
	assert.Equal(t, []interface{}{float64(1), "foo", true, float64(100), float64(1), "2020-01-01", "2023-01-01", "2", float64(10)}, filter)
}

func TestGenerateWithStaticCond(t *testing.T) {
	jsonData := `
		{
      "table": "v_transaction_redemptions",
      "selectFields": [
        "v_transaction_redemptions.transaction_number",
        "v_transaction_redemptions.customer_id",
        "v_transaction_redemptions.customer_name",
        "v_transaction_redemptions.amounts"
      ],
      "conditions": [
        {
            "operand": "and",
            "clause": "v_transaction_redemptions.transaction_date",
            "datatype": "STRING",
            "isStatic": false,
            "operator": "=",
            "value": "2024-03-08 23:59:59"
        },
        {
            "operand": "and",
            "clause": "v_transaction_redemptions.customer_id",
            "datatype": "STRING",
            "isStatic": true,
            "operator": "=",
            "value": "15"
        }
    ]
  }
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{})
	sql, _, _ := jql.Generate()

	fmt.Println(sql)

	// strExpectation := "SELECT table_1.a, table_1.b AS foo_bar, (SELECT * FROM table_4 WHERE a = ? LIMIT 1) AS baz, table_2.b, table_3.a, table_3.b FROM table_1 JOIN table_2 ON table_2.a = table_1.a LEFT JOIN table_3 ON table_3.a = table_2.a WHERE table_1.a = ? AND table_1.b = ? AND table_2.a > sum(?) AND table_2.b = (SELECT * FROM table_4 WHERE a = ? LIMIT 1) OR (table_3.a BETWEEN ? AND ? AND table_3.b = ?) GROUP BY table_1.a HAVING COUNT(table_2.a) > ? ORDER BY table_1.a, table_2.a ASC LIMIT 1 OFFSET 0"
	// assert.Equal(t, strExpectation, sql)
	// assert.Equal(t, []interface{}{float64(1), "foo", true, float64(100), float64(1), "2020-01-01", "2023-01-01", "2", float64(10)}, filter)
}

func TestRawFunction(t *testing.T) {
	jsonData := `
	{
    "selectFields": [
      {
        "alias": "amounts",
        "addFunction": {
          "sqlFunc": {
            "name": "sum",
            "isField": true,
            "params": [
              "v_transaction_redemptions.amounts"
            ]
          }
        }
      }
    ],
    "conditions": [
      {
        "isStatic": false,
        "datatype": "raw",
        "clause": "v_transaction_redemptions.transaction_date",
        "operator": ">=",
        "operand": "AND",
        "value": "CURRENT_DATE - INTERVAL '3 months'"
      },
      {
        "isStatic": false,
        "datatype": "number",
        "clause": "v_transaction_redemptions.customer_id",
        "operator": "=",
        "operand": "AND",
        "value": 15
      }
    ],
    "table": "v_transaction_redemptions"
  }
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{WithSanitizedInjection: true})
	sql, _, _ := jql.Generate()

	fmt.Println(sql)

	// strExpectation := "SELECT table_1.a, table_1.b AS foo_bar, (SELECT * FROM table_4 WHERE a = ? LIMIT 1) AS baz, table_2.b, table_3.a, table_3.b FROM table_1 JOIN table_2 ON table_2.a = table_1.a LEFT JOIN table_3 ON table_3.a = table_2.a WHERE table_1.a = ? AND table_1.b = ? AND table_2.a > sum(?) AND table_2.b = (SELECT * FROM table_4 WHERE a = ? LIMIT 1) OR (table_3.a BETWEEN ? AND ? AND table_3.b = ?) GROUP BY table_1.a HAVING COUNT(table_2.a) > ? ORDER BY table_1.a, table_2.a ASC LIMIT 1 OFFSET 0"
	// assert.Equal(t, strExpectation, sql)
	// assert.Equal(t, []interface{}{float64(1), "foo", true, float64(100), float64(1), "2020-01-01", "2023-01-01", "2", float64(10)}, filter)
}

func TestBuildRawUnion(t *testing.T) {
	jsonData := `
		[
			{
				"table": "table_1",
				"selectFields": [
					"a",
					"b",
          {
            "field": "table_1.b",
            "alias": "foo_bar"
          },
          {
            "field": "table_2.a",
            "alias": "baz",
            "subquery": {
              "table": "table_4",
              "selectFields": ["a","b"],
              "conditions": [
                {
                  "datatype": "number",
                  "clause": "a",
                  "operator": "=",
                  "value": 1
                }
              ],
              "limit": 1
            }
          }
				],
				"conditions": [
					{
						"datatype": "number",
						"clause": "a",
						"operator": "=",
						"value": 1
					}
				],
				"limit": 1
			},
			{
				"table": "table_2",
				"selectFields": [
					"a",
					"b",
          {
            "field": "table_1.b",
            "alias": "foo_bar"
          },
          {
            "field": "table_2.a",
            "alias": "baz",
            "subquery": {
              "table": "table_4",
              "selectFields": ["a","b"],
              "conditions": [
                {
                  "datatype": "number",
                  "clause": "a",
                  "operator": "=",
                  "value": 1
                }
              ],
              "limit": 1
            }
          }
				],
				"conditions": [
					{
						"datatype": "number",
						"clause": "a",
						"operator": "=",
						"value": 1
					}
				],
				"limit": 1
			}
		]
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{WithUnion: true})
	sql := jql.BuildUnion()

	strExpectation := "SELECT a, b, table_1.b AS foo_bar, (SELECT a, b FROM table_4 WHERE a = 1 LIMIT 1) AS baz FROM table_1 WHERE a = 1 LIMIT 1 UNION SELECT a, b, table_1.b AS foo_bar, (SELECT a, b FROM table_4 WHERE a = 1 LIMIT 1) AS baz FROM table_2 WHERE a = 1 LIMIT 1"

	assert.Equal(t, strExpectation, sql)
}

func TestGenerateUnion(t *testing.T) {
	jsonData := `
		[
			{
				"table": "table_1",
				"selectFields": [
					"a",
					"b"
				],
				"conditions": [
					{
						"datatype": "number",
						"clause": "a",
						"operator": "=",
						"value": 1
					}
				],
				"limit": 1
			},
			{
				"table": "table_2",
				"selectFields": [
					"a",
					"b"
				],
				"conditions": [
					{
						"datatype": "number",
						"clause": "a",
						"operator": "=",
						"value": 1
					}
				],
				"limit": 1
			}
		]
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{WithUnion: true})
	sql, filter, _ := jql.GenerateUnion()

	strExpectation := "SELECT a, b FROM table_1 WHERE a = ? LIMIT 1 UNION SELECT a, b FROM table_2 WHERE a = ? LIMIT 1"

	assert.Equal(t, strExpectation, sql)
	assert.Equal(t, []interface{}{float64(1), float64(1)}, filter)
}

func TestGenerateBuild_PreventInjection(t *testing.T) {
	jsonData := `
    {
      "table": "table_2",
      "selectFields": [
        "a",
        "b;drop table table_2 --"
      ],
      "conditions": [
        {
          "datatype": "number",
          "clause": "a",
          "operator": "=",
          "value": 1
        }
      ],
      "limit": 1
    }
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{WithSanitizedInjection: true})
	sql := jql.Build()

	strExpectation := "Invalid sql string you've got sanitized SQL string"

	assert.Equal(t, strExpectation, sql)
}

func TestGenerate_PreventInjection(t *testing.T) {
	jsonData := `
    {
      "table": "table_2",
      "selectFields": [
        "a",
        "b;drop table table_2 --"
      ],
      "conditions": [
        {
          "datatype": "number",
          "clause": "a",
          "operator": "=",
          "value": 1
        }
      ],
      "limit": 1
    }
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{WithSanitizedInjection: true})
	_, _, err := jql.Generate()

	assert.NotNil(t, err)
}

func TestGenerateBuildUnion_PreventInjection(t *testing.T) {
	jsonData := `
    [
      {
        "table": "table_2",
        "selectFields": [
          "a",
          "b;drop table table_2 --"
        ],
        "conditions": [
          {
            "datatype": "number",
            "clause": "a",
            "operator": "=",
            "value": 1
          }
        ],
        "limit": 1
      },
      {
        "table": "table_2",
        "selectFields": [
          "a",
          "b;drop table table_2 --"
        ],
        "conditions": [
          {
            "datatype": "number",
            "clause": "a",
            "operator": "=",
            "value": 1
          }
        ],
        "limit": 1
      }
    ]
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{WithSanitizedInjection: true, WithUnion: true})
	sql := jql.BuildUnion()

	strExpectation := "Invalid sql string you've got sanitized SQL string"

	assert.Equal(t, strExpectation, sql)
}

func TestGenerateUnion_PreventInjection(t *testing.T) {
	jsonData := `
    [
      {
        "table": "table_2",
        "selectFields": [
          "a",
          "b;drop table table_2 --"
        ],
        "conditions": [
          {
            "datatype": "number",
            "clause": "a",
            "operator": "=",
            "value": 1
          }
        ],
        "limit": 1
      },
      {
        "table": "table_2",
        "selectFields": [
          "a",
          "b;drop table table_2 --"
        ],
        "conditions": [
          {
            "datatype": "number",
            "clause": "a",
            "operator": "=",
            "value": 1
          }
        ],
        "limit": 1
      }
    ]
	`

	jql, _ := NewJson2Sql([]byte(jsonData), &Json2SqlConf{WithSanitizedInjection: true, WithUnion: true})
	_, _, err := jql.GenerateUnion()

	assert.NotNil(t, err)
}
