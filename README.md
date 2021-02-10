# Simpler - a simpler database story for Go

A small wrapper around [ozzo-dbx](https://github.com/go-ozzo/ozzo-dbx)

Inspired by [HugSQL](https://www.hugsql.org/)

* Embrace SQL
* Load queries from files
* Be happy

## Usage

Sample `users.sql` file in `data/sql` directory of our project.

```sql
-- name: all-users
SELECT * FROM users

-- name: delete-user
DELETE FROM users WHERE id = {:id}
```


```go
package main

import (
    "coderats.dev/simpler"
    _ "github.com/jackc/pgx/stdlib"
)

func main() {
    // Load all *.sql files from a directory
    registry, err := simpler.NewRegistry("data/sql")
    if err != nil {
        panic(err)
    }

    // you can use any adapter that ozzo-dbx support
    // for example with the following import
    // import _ "github.com/go-sql-driver/mysql"
    // you can use "mysql" as an adapter string (first arg to Connect)
    err = registry.Connect("pgx", "postgres://root:root@localhost:5432/mydb_dev")
    if err != nil {
        panic(err)
    }

    // q is just a *dbx.Query from ozzo-dbx package
    // same result as using dbx.NewQuery call
    q := registry.Query("users/all-users")

    var users []User
    err := q.All(&users)

    // Load delete-user query
    q = registry.Query("users/delete-user")
    // Pass named parameters to a query
    q.Bind(dbx.Params{"id": 3})
    // And execute it
    _, err = q.Execute()

    // Or you can access *dbx.DB directly
    user := User{
        Name: "example",
        Email: "test@example.com",
    }
    err = registry.DB().Model(&user).Insert()

    // Example database/sql pool configuration
    registry.DB().DB().SetMaxOpenConns(100)
    registry.DB().DB().SetMaxIdleConns(10)
    registry.DB().DB().SetConnMaxLifetime(0)
}
```
