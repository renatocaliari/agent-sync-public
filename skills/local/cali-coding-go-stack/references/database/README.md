# Database Reference

SQLite database helper with connection pooling, migrations, and transactions.

## Files

| File | Description |
|------|-------------|
| `database.go` | Database wrapper with read/write pools and migrations |
| `user_crud_example.templ` | Complete CRUD example with users table |

## Setup

### 1. Add Dependencies

```bash
go get zombiezen.com/go/sqlite
go get zombiezen.com/go/sqlite/sqlitemigration
go get zombiezen.com/go/sqlite/sqlitex
```

### 2. Copy database.go

Copy `database.go` to your project:

```
yourproject/
└── db/
    └── database.go
```

### 3. Initialize Database

```go
import "yourproject/db"

func main() {
    ctx := context.Background()
    
    database, err := db.NewDatabase(ctx,
        db.DatabaseWithFilename("data/app.sqlite"),
        db.DatabaseWithMigrations([]string{
            `CREATE TABLE IF NOT EXISTS users (
                id TEXT PRIMARY KEY,
                name TEXT NOT NULL,
                email TEXT NOT NULL,
                created_at INTEGER
            )`,
        }),
    )
    if err != nil {
        panic(err)
    }
    defer database.Close()
}
```

## API

### Options

```go
db.DatabaseWithFilename("path/to/db.sqlite")
db.DatabaseWithMigrations([]string{"CREATE TABLE ...", "ALTER TABLE ..."})
db.DatabaseWithPragmas("foreign_keys = ON", "journal_mode = WAL")
db.DatabaseWithShouldClear(true)  // Clear existing data
```

### Read Transaction

```go
err := database.ReadTX(ctx, func(tx *sqlite.Conn) error {
    stmt := tx.Prep("SELECT id, name FROM users")
    defer stmt.Reset()

    for {
        hasRow, err := stmt.Step()
        if err != nil || !hasRow {
            return err
        }
        // Process row: stmt.GetText("id"), stmt.GetText("name")
    }
})
```

### Write Transaction

```go
err := database.WriteTX(ctx, func(tx *sqlite.Conn) error {
    stmt := tx.Prep("INSERT INTO users (id, name) VALUES ($id, $name)")
    defer stmt.Reset()

    stmt.SetText("$id", "123")
    stmt.SetText("$name", "John")

    _, err := stmt.Step()
    return err
})
```

## Statement Methods

```go
stmt.SetText("$param", "value")
stmt.SetInt64("$param", 123)
stmt.SetFloat("$param", 1.5)
stmt.SetZeroBlob("$param", 1024)  // BLOB

stmt.GetText("column")
stmt.GetInt64("column")
stmt.GetFloat("column")
stmt.GetLen("column")  // For BLOB
stmt.GetBytes("column", buf)
```

## Example: Full CRUD Service

See `user_crud_example.templ` for a complete example with:
- Service layer with Create, Read, Update, Delete
- HTTP handlers with Datastar SSE
- Templ templates for UI

### Migrations Pattern

```go
migrations := []string{
    // Migration 1
    `CREATE TABLE users (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        email TEXT UNIQUE,
        created_at INTEGER
    )`,
    
    // Migration 2
    `CREATE INDEX idx_users_email ON users(email)`,
    
    // Migration 3
    `ALTER TABLE users ADD COLUMN updated_at INTEGER`,
}

db.NewDatabase(ctx, db.DatabaseWithMigrations(migrations))
```

## Features

- **Connection Pooling**: Separate read and write pools
- **WAL Mode**: Write-Ahead Logging for better concurrency
- **Auto-Migrations**: Run migrations on startup
- **Transaction Support**: Read and write transactions
- **Pragmas**: Configure SQLite behavior