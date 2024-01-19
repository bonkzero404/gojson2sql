# GoJSON2SQL

[![Go Report Card](https://goreportcard.com/badge/github.com/bonkzero404/gojson2sql)](https://goreportcard.com/report/github.com/bonkzero404/gojson2sql)
[![codecov](https://codecov.io/gh/bonkzero404/gojson2sql/branch/main/graphs/badge.svg?branch=main)](https://codecov.io/gh/bonkzero404/gojson2sql)
[![build-status](https://github.com/bonkzero404/gojson2sql/workflows/Go/badge.svg)](https://github.com/bonkzero404/gojson2sql/actions)

GoJson2SQL is a library for composing SQL queries using JSON. A JSON file is transformed into an SQL string. This library facilitates the process of generating SQL statements by utilizing a structured JSON format, enhancing the readability and simplicity of SQL query construction.

## Limitations

Currently, it can only perform select queries.

## TODO:

- More Queries
- Validate SQL Syntax
- ?

## Installation

```
go get github.com/bonkzero404/gojson2sql
```

## Simple Example

```go
package main

import (
	"fmt"

	"github.com/bonkzero404/gojson2sql"
)

func main() {
	sqlJson := `
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
		}
	`
	jql, err := gojson2sql.NewJson2Sql([]byte(sqlJson))
	if err != nil {
		panic(err)
	}

	sql, param, _ := jql.Generate()

	fmt.Println("SQL:", sql)
	fmt.Println("Param:", param)
}
```

Output:

```sql
SQL: SELECT a, b FROM table_1 WHERE a = ? LIMIT
Param: [1]
```

You can use this raw SQL with either the sql package or GORM. Here's an example with sql package:

```go
jql, _ := gojson2sql.NewJson2Sql([]byte(sqlJson))
sql, param, _ := jql.Generate()

db.Query(sql, param)
```

## Example Union Query Operation

```json
[
  {
    "table": "table_1",
    "selectFields": ["a", "b"],
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
    "selectFields": ["a", "b"],
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
```

Output:

```sql
SELECT a, b FROM table_1 WHERE a = 1 LIMIT 1 UNION SELECT a, b FROM table_2 WHERE a = 1 LIMIT 1
```

You can see the difference between a union query and a standard select. In a union, you must use a JSON array with the standard JSON format as before.

## Operator Lists

```go
const (
	Equal        SQLOperatorEnum = "="
	NotEqual     SQLOperatorEnum = "<>"
	LessThan     SQLOperatorEnum = "<"
	LessEqual    SQLOperatorEnum = "<="
	GreaterThan  SQLOperatorEnum = ">"
	GreaterEqual SQLOperatorEnum = ">="
	Like         SQLOperatorEnum = "LIKE"
	Ilike        SQLOperatorEnum = "ILIKE"
	Between      SQLOperatorEnum = "BETWEEN"
	NotLike      SQLOperatorEnum = "NOT LIKE"
	In           SQLOperatorEnum = "IN"
	NotIn        SQLOperatorEnum = "NOT IN"
	IsNull       SQLOperatorEnum = "IS NULL"
	IsNotNull    SQLOperatorEnum = "IS NOT NULL"
)
```

## Datatype Lists

```go
const (
	Boolean  SQLDataTypeEnum = "BOOLEAN"
	String   SQLDataTypeEnum = "STRING"
	Number   SQLDataTypeEnum = "NUMBER"
	Raw      SQLDataTypeEnum = "RAW"
	Function SQLDataTypeEnum = "FUNCTION"
	Array    SQLDataTypeEnum = "ARRAY"
)
```

## isStatic / isField Properties

- isStatic: (boolean)
  isStatic is used to set the value of the clause. If true, the value will not be parsed to parameters. However, if false, the opposite will occur. This is typically used in where statements.

- isField: (boolean)
  isField is used to set a field so that it does not use single quotes. For example, if you describe a function and do not use isField, it will look like this: COUNT('field'). However, if you use isField, it will look like this: COUNT(field).

## JSON Format

In general, the structure of the JSON format used is as follows:

- **_table_**: Used to describe the table name, e.g. table_name (string)

- **selectFields**:
  Used to select fields from a table, this property uses the **_Array_** type, you can combine **_Array of String_**, and **_Array of Json_**, the example is as follows:

  ```json
  {
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
    ]
  }
  ```

  There you can see there is a subquery, you can use a subquery with the same format as its parent, you can also describe a field with an alias in the selection field.

  > **_NOTE:_** If you are using a subquery, you don't need to describe the datatype property

- **_join_**:
  You can use join to combine multiple tables, an example is as follows:

  ```json
  {
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
    ]
  }
  ```

- **conditions**:
  Conditions are used for SQL Where clauses. The structure of these conditions is dynamic; you can use a function, subquery, or composite. Consider the following example:
  ```json
  {
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
    ]
  }
  ```
- **groupBy**:

  ```json
  {
    "groupBy": {
      "fields": ["table_1.a"]
    }
  }
  ```

- **having**
  ```json
  {
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
    ]
  }
  ```
- **orderBy, limit, offset**:
  ```json
  {
    "orderBy": {
      "fields": ["table_1.a", "table_2.a"],
      "sort": "asc"
    },
    "limit": 1,
    "offset": 0
  }
  ```
  Or you can describe the limit and offset like this
  ```json
  {
    "offset": {
      "isStatic": true,
      "value": 10
    },
    "offset": {
      "value": 10
    }
  }
  ```

## Convert to Raw Query

You can also convert to raw query without parameters.

```go
jql, err := gojson2sql.NewJson2Sql([]byte(sqlJson))
if err != nil {
  panic(err)
}

sql := jql.Build()

fmt.Println("SQL:", sql)
```

Output:

```sql
SQL: SELECT a, b FROM table_1 WHERE a = 1 LIMIT 1
```

## Full Example Advance Query

```go
sqlJson := `
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
jql, err := gojson2sql.NewJson2Sql([]byte(sqlJson))
if err != nil {
  panic(err)
}

sql, param, _ := jql.Generate()

fmt.Println("SQL:", sql)
fmt.Println("Param:", param)
```

output:

```sql
SQL:
SELECT
  table_1.a,
  table_1.b AS foo_bar,
  (SELECT * FROM table_4 WHERE a = ? LIMIT 1) AS baz,
  table_2.b,
  table_3.a,
  table_3.b
FROM table_1
  JOIN table_2 ON table_2.a = table_1.a
  LEFT JOIN table_3 ON table_3.a = table_2.a
WHERE
  table_1.a = ? AND
  table_1.b = ? AND
  table_2.a > sum(?) AND
  table_2.b = (SELECT * FROM table_4 WHERE a = ? LIMIT 1) OR
  (table_3.a BETWEEN ? AND ? AND table_3.b = ?)
GROUP BY table_1.a
HAVING COUNT(table_2.a) > ?
ORDER BY table_1.a, table_2.a ASC
LIMIT 1
OFFSET 0

Param:
[1 foo true 100 1 2020-01-01 2023-01-01 2 10]
```

## Complete Example JSON

```json
{
  "table": "table_1",
  "selectFields": [
    {
      "field": "table_1.a",
      "alias": "foo_bar"
    },
    {
      "alias": "foo_bar_baz",
      "addFunction": {
        "sqlFunc": {
          "name": "count",
          "isField": true,
          "params": ["table_1.b"]
        }
      }
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
    {
      "when": [
        {
          "clause": "table_2.b",
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
            "table": "table_3",
            "selectFields": ["*"],
            "conditions": [
              {
                "datatype": "NUMBER",
                "clause": "a",
                "operator": "=",
                "value": 1
              }
            ],
            "limit": 1
          }
        }
      },
      "alias": "field_alias"
    },
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
  "limit": {
    "isStatic": true,
    "value": 10
  },
  "offset": 0
}
```

output:

```sql
SELECT
  table_1.a AS foo_bar,
  COUNT(table_1.b) AS foo_bar_baz,
  (SELECT * FROM table_4 WHERE a = 1 LIMIT 1) AS baz,
  CASE
    WHEN table_2.b > 100 THEN true
    ELSE (SELECT * FROM table_3 WHERE a = 1 LIMIT 1)
  END AS field_alias,
  table_3.a,
  table_3.b
FROM table_1
  JOIN table_2 ON table_2.a = table_1.a
  LEFT JOIN table_3 ON table_3.a = table_2.a
WHERE
  table_1.a = 'foo' AND
  table_1.b = true AND
  table_2.a > sum(100) AND
  table_2.b = (SELECT * FROM table_4 WHERE a = 1 LIMIT 1) OR
  (table_3.a BETWEEN '2020-01-01' AND '2023-01-01' AND table_3.b = '2')
GROUP BY table_1.a
HAVING
  COUNT(table_2.a) > 10
ORDER BY table_1.a, table_2.a ASC
LIMIT 10
OFFSET 0
```

## Testing

```go
> go test -v -cover ./...
=== RUN   TestConstructor
--- PASS: TestConstructor (0.00s)
=== RUN   TestConstructor_Fail
--- PASS: TestConstructor_Fail (0.00s)
=== RUN   TestConstructor_Fail_Union
--- PASS: TestConstructor_Fail_Union (0.00s)
=== RUN   TestRawJson_OK
--- PASS: TestRawJson_OK (0.00s)
=== RUN   TestRawJson_Error
--- PASS: TestRawJson_Error (0.00s)
=== RUN   TestMaskedQueryValue
--- PASS: TestMaskedQueryValue (0.00s)
=== RUN   TestGenerateSelectFrom
--- PASS: TestGenerateSelectFrom (0.00s)
=== RUN   TestGenerateSelectFrom_Selection
--- PASS: TestGenerateSelectFrom_Selection (0.00s)
=== RUN   TestGenerateSelectFrom_CaseWhenThen
--- PASS: TestGenerateSelectFrom_CaseWhenThen (0.00s)
=== RUN   TestGenerateSelectFrom_CaseDefaultValueSub
--- PASS: TestGenerateSelectFrom_CaseDefaultValueSub (0.00s)
=== RUN   TestGenerateSelectFrom_SqlFunc
--- PASS: TestGenerateSelectFrom_SqlFunc (0.00s)
=== RUN   TestSqlLikeAndBlankDatatype
--- PASS: TestSqlLikeAndBlankDatatype (0.00s)
=== RUN   TestSqlLikeWithOperand
--- PASS: TestSqlLikeWithOperand (0.00s)
=== RUN   TestBetweenWithOperand
--- PASS: TestBetweenWithOperand (0.00s)
=== RUN   TestCompositeWithoutOperand
--- PASS: TestCompositeWithoutOperand (0.00s)
=== RUN   TestGenerateOrderBy
--- PASS: TestGenerateOrderBy (0.00s)
=== RUN   TestGenerateOrderBy_WithSort
--- PASS: TestGenerateOrderBy_WithSort (0.00s)
=== RUN   TestGenerateGroupBy
--- PASS: TestGenerateGroupBy (0.00s)
=== RUN   TestGenerateJoin_JOIN
--- PASS: TestGenerateJoin_JOIN (0.00s)
=== RUN   TestGenerateJoin_INNER_JOIN
--- PASS: TestGenerateJoin_INNER_JOIN (0.00s)
=== RUN   TestGenerateJoin_LEFT_JOIN
--- PASS: TestGenerateJoin_LEFT_JOIN (0.00s)
=== RUN   TestGenerateJoin_RIGHT_JOIN
--- PASS: TestGenerateJoin_RIGHT_JOIN (0.00s)
=== RUN   TestGenerateHaving
--- PASS: TestGenerateHaving (0.00s)
=== RUN   TestGenerateWhere
--- PASS: TestGenerateWhere (0.00s)
=== RUN   TestGenerateConditions
--- PASS: TestGenerateConditions (0.00s)
=== RUN   TestGenerateConditions_SubQuery
--- PASS: TestGenerateConditions_SubQuery (0.00s)
=== RUN   TestLimit_Static
--- PASS: TestLimit_Static (0.00s)
=== RUN   TestLimit_ToParam
--- PASS: TestLimit_ToParam (0.00s)
=== RUN   TestOffset_Static
--- PASS: TestOffset_Static (0.00s)
=== RUN   TestOffset_ToParam
--- PASS: TestOffset_ToParam (0.00s)
=== RUN   TestBuildJsonToSql
--- PASS: TestBuildJsonToSql (0.00s)
=== RUN   TestGenerateJsonToSql
--- PASS: TestGenerateJsonToSql (0.00s)
=== RUN   TestBuildRawUnion
--- PASS: TestBuildRawUnion (0.00s)
=== RUN   TestGenerateUnion
--- PASS: TestGenerateUnion (0.00s)
=== RUN   TestIsValidDataType
--- PASS: TestIsValidDataType (0.00s)
=== RUN   TestGetValueFromDataType
--- PASS: TestGetValueFromDataType (0.00s)
=== RUN   TestCheckArrayType
--- PASS: TestCheckArrayType (0.00s)
=== RUN   TestArrayConversionToStringExpression
--- PASS: TestArrayConversionToStringExpression (0.00s)
=== RUN   TestExtractValueByDataType
--- PASS: TestExtractValueByDataType (0.00s)
=== RUN   TestGetSqlExpression
--- PASS: TestGetSqlExpression (0.00s)
=== RUN   TestIsValidOperator
--- PASS: TestIsValidOperator (0.00s)
=== RUN   TestGetValueFromOperator
--- PASS: TestGetValueFromOperator (0.00s)
PASS
coverage: 100.0% of statements
```

## Benchmarking

Specs:

- MacBook Pro M1 (2020)
- 8-Cores (arm64)
- 8GB of RAM

```go
> go test -bench=. -benchmem
goos: darwin
goarch: arm64
pkg: github.com/bonkzero404/gojson2sql
BenchmarkJson2Sql_BuildRaw-8               12708             93796 ns/op           32254 B/op        557 allocs/op
BenchmarkJson2Sql_Generate-8                9804            121693 ns/op           43773 B/op        656 allocs/op
BenchmarkJson2Sql_Union_BuildRaw-8          6015            196128 ns/op           70700 B/op       1106 allocs/op
BenchmarkJson2Sql_Union_Generate-8          4876            240539 ns/op           80684 B/op       1232 allocs/op
```
