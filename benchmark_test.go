package gojson2sql

import (
	"testing"
)

func BenchmarkJson2Sql_Build(b *testing.B) {
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

	for i := 0; i < b.N; i++ {
		jql, _ := NewJson2Sql([]byte(jsonData))
		jql.Build()
	}
}

func BenchmarkJson2Sql_Generate(b *testing.B) {
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

	for i := 0; i < b.N; i++ {
		jql, _ := NewJson2Sql([]byte(jsonData))
		jql.Generate()
	}
}
