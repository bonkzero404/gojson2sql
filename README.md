# GoJSON2SQL

[![Go Report Card](https://goreportcard.com/badge/github.com/bonkzero404/gojson2sql)](https://goreportcard.com/report/github.com/bonkzero404/gojson2sql)
[![codecov](https://codecov.io/gh/bonkzero404/gojson2sql/branch/main/graphs/badge.svg?branch=main)](https://codecov.io/gh/bonkzero404/gojson2sql)
[![build-status](https://github.com/bonkzero404/gojson2sql/workflows/Go/badge.svg)](https://github.com/bonkzero404/gojson2sql/actions)

GoJson2SQL is a library for composing SQL queries using JSON. A JSON file is transformed into an SQL string. This library facilitates the process of generating SQL statements by utilizing a structured JSON format, enhancing the readability and simplicity of SQL query construction.

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

```
SQL: SELECT a, b FROM table_1 WHERE a = ? LIMIT
Param: [1]
```

You can use this raw SQL with either the sql package or GORM. Here's an example with sql package:

```go
jql, _ := gojson2sql.NewJson2Sql([]byte(sqlJson))
sql, param, _ := jql.Generate()

db.Query(sql, param)
```

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

## JSON Format

In general, the structure of the JSON format used is as follows:

- **_table_**: Used to describe the table name, e.g. table_name (string)

- **\*selectFields\*\***:
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

  > **_NOTE:_** If you are using a subquery, you do not need to describe the data type

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
  **orderBy, limit, offset**:
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

```
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

```
SQL:
SELECT table_1.a, table_1.b AS foo_bar, (SELECT * FROM table_4 WHERE a = ? LIMIT 1) AS baz, table_2.b, table_3.a, table_3.b FROM table_1 JOIN table_2 ON table_2.a = table_1.a LEFT JOIN table_3 ON table_3.a = table_2.a WHERE table_1.a = ? AND table_1.b = ? AND table_2.a > sum(?) AND table_2.b = (SELECT * FROM table_4 WHERE a = ? LIMIT 1) OR (table_3.a BETWEEN ? AND ? AND table_3.b = ?) GROUP BY table_1.a HAVING COUNT(table_2.a) > ? ORDER BY table_1.a, table_2.a ASC LIMIT 1 OFFSET 0

Param:
[1 foo true 100 1 2020-01-01 2023-01-01 2 10]
```
