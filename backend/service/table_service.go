package service

import (
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sqlmanager/pkg/db"
	"strings"
	"time"
	"unicode/utf8"
)

// safeConvertValue converts a database value to a JSON-safe representation
// Binary data that is not valid UTF-8 is converted to hex string
func safeConvertValue(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	if b, ok := v.([]byte); ok {
		if len(b) == 0 {
			return ""
		}
		if utf8.Valid(b) {
			return string(b)
		}
		// Binary data - display as hex
		return "0x" + hex.EncodeToString(b)
	}
	return v
}

// TableInfo represents table metadata
type TableInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	RowCount  int64  `json:"rowCount"`
	Size      int64  `json:"size,omitempty"`
	Engine    string `json:"engine,omitempty"`
	Comment   string `json:"comment,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// ColumnInfo represents column metadata
type ColumnInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"defaultValue"`
	Comment      string `json:"comment,omitempty"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
	Extra        string `json:"extra,omitempty"`
}

// IndexInfo represents index metadata
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// TableSchema represents complete table schema
type TableSchema struct {
	Columns       []ColumnInfo `json:"columns"`
	Indexes       []IndexInfo  `json:"indexes"`
	PrimaryKey    []string     `json:"primaryKey"`
	AutoIncrement *int64       `json:"autoIncrement,omitempty"`
}

// TableData represents query result data
type TableData struct {
	Columns  []string        `json:"columns"`
	Rows     [][]interface{} `json:"rows"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
}

// QueryResult represents SQL execution result
type QueryResult struct {
	Success       bool       `json:"success"`
	Error         string     `json:"error,omitempty"`
	AffectedRows  int64      `json:"affectedRows,omitempty"`
	Data          *TableData `json:"data,omitempty"`
	ExecutionTime float64    `json:"executionTime"` // milliseconds
}

// TableService handles table operations
type TableService struct{}

// NewTableService creates a new table service
func NewTableService() *TableService {
	return &TableService{}
}

// ListTables returns a list of all tables in the database
func (ts *TableService) ListTables(conn db.DBConnection, dbName string) ([]TableInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Switch to the target database first (critical for PostgreSQL)
	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return nil, fmt.Errorf("failed to select database %s: %w", dbName, err)
		}
	}

	var query string
	var args []interface{}
	dbType := conn.GetDBType()

	switch dbType {
	case db.MySQL:
		query = `SELECT table_name, table_type, table_rows, data_length, engine, table_comment, create_time, update_time
				 FROM information_schema.tables 
				 WHERE table_schema = ? AND table_type IN ('BASE TABLE', 'VIEW')
				 ORDER BY CASE WHEN table_type = 'BASE TABLE' THEN 0 ELSE 1 END, table_name`
		args = []interface{}{dbName}
	case db.PostgreSQL:
		query = `SELECT table_name, table_type
		         FROM information_schema.tables
		         WHERE table_catalog = $1
		           AND table_schema NOT IN ('pg_catalog', 'information_schema')
		         ORDER BY CASE WHEN table_type = 'BASE TABLE' THEN 0 ELSE 1 END, table_schema, table_name`
		args = []interface{}{dbName}
	case db.SQLite:
		query = `SELECT name, type FROM sqlite_master WHERE type IN ('table', 'view') AND name NOT LIKE 'sqlite_%' ORDER BY CASE WHEN type='table' THEN 0 ELSE 1 END, name`
	}

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var table TableInfo

		switch dbType {
		case db.MySQL:
			var objectType string
			var rowCount sql.NullInt64
			var size sql.NullInt64
			var engine sql.NullString
			var comment sql.NullString
			var createdAt sql.NullTime
			var updatedAt sql.NullTime
			if err := rows.Scan(&table.Name, &objectType, &rowCount, &size, &engine, &comment, &createdAt, &updatedAt); err != nil {
				return nil, err
			}
			if objectType == "VIEW" {
				table.Type = "view"
			} else {
				table.Type = "table"
			}
			if rowCount.Valid {
				table.RowCount = rowCount.Int64
			}
			if size.Valid {
				table.Size = size.Int64
			}
			if engine.Valid {
				table.Engine = engine.String
			}
			if comment.Valid {
				table.Comment = comment.String
			}
			if createdAt.Valid {
				table.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			}
			if updatedAt.Valid {
				table.UpdatedAt = updatedAt.Time.Format("2006-01-02 15:04:05")
			}
		case db.PostgreSQL:
			var objectType string
			if err := rows.Scan(&table.Name, &objectType); err != nil {
				return nil, err
			}
			if objectType == "VIEW" {
				table.Type = "view"
			} else {
				table.Type = "table"
			}
			table.RowCount = 0
		case db.SQLite:
			var objectType string
			if err := rows.Scan(&table.Name, &objectType); err != nil {
				return nil, err
			}
			if objectType == "view" {
				table.Type = "view"
			} else {
				table.Type = "table"
			}
			table.RowCount = 0
		}

		tables = append(tables, table)
	}

	return tables, rows.Err()
}

// GetTableSchema returns the schema of a table
func (ts *TableService) GetTableSchema(conn db.DBConnection, dbName, tableName string) (*TableSchema, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	schema := &TableSchema{
		Columns:    []ColumnInfo{},
		Indexes:    []IndexInfo{},
		PrimaryKey: []string{},
	}

	dbType := conn.GetDBType()

	// Get columns
	var colQuery string
	var args []interface{}

	switch dbType {
	case db.MySQL:
		colQuery = `SELECT column_name, column_type, is_nullable, column_default, column_comment, column_key, extra
					FROM information_schema.columns
					WHERE table_schema = ? AND table_name = ?
					ORDER BY ordinal_position`
		args = []interface{}{dbName, tableName}
	case db.PostgreSQL:
		colQuery = `SELECT column_name, data_type, is_nullable, column_default
					FROM information_schema.columns
					WHERE table_catalog = $1 AND table_name = $2
					ORDER BY ordinal_position`
		args = []interface{}{dbName, tableName}
	case db.SQLite:
		colQuery = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	}

	rows, err := conn.Query(ctx, colQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get table schema: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo

		switch dbType {
		case db.MySQL:
			var nullable, columnKey, extra string
			var defaultVal, comment sql.NullString
			if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal, &comment, &columnKey, &extra); err != nil {
				return nil, err
			}
			col.Nullable = nullable == "YES"
			if defaultVal.Valid {
				col.DefaultValue = defaultVal.String
			}
			if comment.Valid {
				col.Comment = comment.String
			}
			col.Extra = extra
			col.IsPrimaryKey = columnKey == "PRI"
			if col.IsPrimaryKey {
				schema.PrimaryKey = append(schema.PrimaryKey, col.Name)
			}
		case db.PostgreSQL:
			var nullable string
			var defaultVal sql.NullString
			if err := rows.Scan(&col.Name, &col.Type, &nullable, &defaultVal); err != nil {
				return nil, err
			}
			col.Nullable = nullable == "YES"
			if defaultVal.Valid {
				col.DefaultValue = defaultVal.String
			}
		case db.SQLite:
			var cid int
			var notNull, pk int
			var defaultVal sql.NullString
			if err := rows.Scan(&cid, &col.Name, &col.Type, &notNull, &defaultVal, &pk); err != nil {
				return nil, err
			}
			col.Nullable = notNull == 0
			if defaultVal.Valid {
				col.DefaultValue = defaultVal.String
			}
			col.IsPrimaryKey = pk > 0
			if col.IsPrimaryKey {
				schema.PrimaryKey = append(schema.PrimaryKey, col.Name)
			}
		}

		schema.Columns = append(schema.Columns, col)
	}

	// Get primary key for PostgreSQL separately
	if dbType == db.PostgreSQL {
		pkQuery := `SELECT a.attname
					FROM pg_index i
					JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
					WHERE i.indrelid = $1::regclass AND i.indisprimary`
		pkRows, err := conn.Query(ctx, pkQuery, tableName)
		if err == nil {
			defer pkRows.Close()
			for pkRows.Next() {
				var colName string
				if err := pkRows.Scan(&colName); err == nil {
					schema.PrimaryKey = append(schema.PrimaryKey, colName)
					// Mark column as primary key
					for i := range schema.Columns {
						if schema.Columns[i].Name == colName {
							schema.Columns[i].IsPrimaryKey = true
						}
					}
				}
			}
		}
	}

	// Get auto-increment value
	switch dbType {
	case db.MySQL:
		var autoInc sql.NullInt64
		aiQuery := `SELECT AUTO_INCREMENT FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`
		if err := conn.QueryRow(ctx, aiQuery, dbName, tableName).Scan(&autoInc); err == nil && autoInc.Valid {
			schema.AutoIncrement = &autoInc.Int64
		}
	case db.PostgreSQL:
		// PostgreSQL: find sequence value for serial/identity columns
		for _, col := range schema.Columns {
			if strings.Contains(strings.ToLower(col.DefaultValue), "nextval") {
				var lastVal int64
				seqQuery := fmt.Sprintf(`SELECT last_value + 1 FROM pg_sequences WHERE schemaname = 'public' AND sequencename = '%s_%s_seq'`, tableName, col.Name)
				if err := conn.QueryRow(ctx, seqQuery).Scan(&lastVal); err == nil {
					schema.AutoIncrement = &lastVal
				}
				break
			}
		}
	case db.SQLite:
		var seqVal sql.NullInt64
		seqQuery := `SELECT seq FROM sqlite_sequence WHERE name = ?`
		if err := conn.QueryRow(ctx, seqQuery, tableName).Scan(&seqVal); err == nil && seqVal.Valid {
			nextVal := seqVal.Int64 + 1
			schema.AutoIncrement = &nextVal
		}
	}

	return schema, nil
}

// GetTableData returns paginated data from a table (max 500 rows)
func (ts *TableService) GetTableData(conn db.DBConnection, dbName, tableName string, page, pageSize int, sortCol, sortDir string) (*TableData, error) {
	startTime := time.Now()

	// Enforce max 500 rows
	if pageSize > 500 {
		pageSize = 500
	}
	if pageSize <= 0 {
		pageSize = 100
	}
	if page <= 0 {
		page = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbType := conn.GetDBType()

	// Build fully qualified table name with proper escaping to prevent SQL injection
	var fullTableName string
	switch dbType {
	case db.MySQL:
		escapedDb := strings.ReplaceAll(dbName, "`", "``")
		escapedTable := strings.ReplaceAll(tableName, "`", "``")
		fullTableName = fmt.Sprintf("`%s`.`%s`", escapedDb, escapedTable)
	case db.PostgreSQL:
		escapedTable := strings.ReplaceAll(tableName, `"`, `""`)
		fullTableName = fmt.Sprintf(`"%s"`, escapedTable)
	case db.SQLite:
		escapedTable := strings.ReplaceAll(tableName, "`", "``")
		fullTableName = fmt.Sprintf("`%s`", escapedTable)
	}

	// Build query with pagination and optional sort
	offset := (page - 1) * pageSize

	// Build ORDER BY clause
	orderClause := ""
	if sortCol != "" {
		dir := "ASC"
		if strings.ToLower(sortDir) == "desc" {
			dir = "DESC"
		}
		switch dbType {
		case db.MySQL, db.SQLite:
			orderClause = fmt.Sprintf(" ORDER BY `%s` %s", strings.ReplaceAll(sortCol, "`", "``"), dir)
		case db.PostgreSQL:
			orderClause = fmt.Sprintf(` ORDER BY "%s" %s`, strings.ReplaceAll(sortCol, `"`, `""`), dir)
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s%s LIMIT %d OFFSET %d",
		fullTableName, orderClause, pageSize, offset)

	queryStart := time.Now()
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get table data: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Read rows
	var data [][]interface{}
	for rows.Next() {
		// Create a slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Convert []byte to string safely for JSON serialization
		row := make([]interface{}, len(columns))
		for i, v := range values {
			row[i] = safeConvertValue(v)
		}

		data = append(data, row)
	}

	queryTime := time.Since(queryStart)

	// Get accurate total count
	var total int64
	rowCount := len(data)
	if rowCount < pageSize {
		// Last page — exact count
		total = int64(offset + rowCount)
	} else {
		// Full page — need COUNT(*) for accurate total
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", fullTableName)
		countRow := conn.QueryRow(ctx, countQuery)
		if err := countRow.Scan(&total); err != nil {
			// Fallback: estimate conservatively
			total = int64(offset + rowCount)
		}
	}

	totalTime := time.Since(startTime)
	_ = totalTime // reserved for future logging

	return &TableData{
		Columns:  columns,
		Rows:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, rows.Err()
}

// ExecuteSQL executes a SQL query and returns the result
func (ts *TableService) ExecuteSQL(conn db.DBConnection, sqlQuery string) (*QueryResult, error) {
	startTime := time.Now()
	result := &QueryResult{Success: true}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Determine if it's a SELECT query
	trimmedQuery := strings.TrimSpace(strings.ToUpper(sqlQuery))
	isSelect := strings.HasPrefix(trimmedQuery, "SELECT") || strings.HasPrefix(trimmedQuery, "SHOW") || strings.HasPrefix(trimmedQuery, "DESCRIBE") || strings.HasPrefix(trimmedQuery, "PRAGMA")

	if isSelect {
		// Execute query that returns rows
		rows, err := conn.Query(ctx, sqlQuery)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.ExecutionTime = float64(time.Since(startTime).Milliseconds())
			return result, nil
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.ExecutionTime = float64(time.Since(startTime).Milliseconds())
			return result, nil
		}

		// Read all rows
		var data [][]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			if err := rows.Scan(valuePtrs...); err != nil {
				result.Success = false
				result.Error = err.Error()
				result.ExecutionTime = float64(time.Since(startTime).Milliseconds())
				return result, nil
			}

			row := make([]interface{}, len(columns))
			for i, v := range values {
				row[i] = safeConvertValue(v)
			}

			data = append(data, row)
		}

		result.Data = &TableData{
			Columns:  columns,
			Rows:     data,
			Total:    int64(len(data)),
			Page:     1,
			PageSize: len(data),
		}
	} else {
		// Execute query that doesn't return rows (INSERT, UPDATE, DELETE, etc.)
		execResult, err := conn.Exec(ctx, sqlQuery)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.ExecutionTime = float64(time.Since(startTime).Milliseconds())
			return result, nil
		}

		affected, _ := execResult.RowsAffected()
		result.AffectedRows = affected
	}

	result.ExecutionTime = float64(time.Since(startTime).Milliseconds())
	return result, nil
}

// UpdateRow updates a single row in a table
func (ts *TableService) UpdateRow(conn db.DBConnection, dbName, tableName string, pkValues, updates map[string]interface{}) error {
	if len(pkValues) == 0 {
		return fmt.Errorf("primary key values are required")
	}
	if len(updates) == 0 {
		return fmt.Errorf("no updates provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	// Build SET clause
	var setClauses []string
	var args []interface{}
	argIndex := 1

	for col, val := range updates {
		if conn.GetDBType() == db.PostgreSQL {
			setClauses = append(setClauses, fmt.Sprintf("\"%s\" = $%d", col, argIndex))
		} else {
			setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", col))
		}
		args = append(args, val)
		argIndex++
	}

	// Build WHERE clause
	var whereClauses []string
	for col, val := range pkValues {
		if conn.GetDBType() == db.PostgreSQL {
			whereClauses = append(whereClauses, fmt.Sprintf("\"%s\" = $%d", col, argIndex))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("`%s` = ?", col))
		}
		args = append(args, val)
		argIndex++
	}

	// Build UPDATE query
	var query string
	if conn.GetDBType() == db.PostgreSQL {
		query = fmt.Sprintf("UPDATE \"%s\" SET %s WHERE %s",
			tableName,
			strings.Join(setClauses, ", "),
			strings.Join(whereClauses, " AND "))
	} else {
		query = fmt.Sprintf("UPDATE `%s` SET %s WHERE %s",
			tableName,
			strings.Join(setClauses, ", "),
			strings.Join(whereClauses, " AND "))
	}

	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update row: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("no rows were updated")
	}

	return nil
}

// InsertRow inserts a new row into a table
func (ts *TableService) InsertRow(conn db.DBConnection, dbName, tableName string, values map[string]interface{}) error {
	if len(values) == 0 {
		return fmt.Errorf("no values provided")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	var columns []string
	var placeholders []string
	var args []interface{}
	argIndex := 1

	for col, val := range values {
		if conn.GetDBType() == db.PostgreSQL {
			columns = append(columns, fmt.Sprintf("\"%s\"", col))
			placeholders = append(placeholders, fmt.Sprintf("$%d", argIndex))
		} else {
			columns = append(columns, fmt.Sprintf("`%s`", col))
			placeholders = append(placeholders, "?")
		}
		args = append(args, val)
		argIndex++
	}

	var query string
	if conn.GetDBType() == db.PostgreSQL {
		query = fmt.Sprintf("INSERT INTO \"%s\" (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "))
	} else {
		query = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "))
	}

	_, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert row: %w", err)
	}

	return nil
}

// DeleteRow deletes a row from a table
func (ts *TableService) DeleteRow(conn db.DBConnection, dbName, tableName string, pkValues map[string]interface{}) error {
	if len(pkValues) == 0 {
		return fmt.Errorf("primary key values are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	var whereClauses []string
	var args []interface{}
	argIndex := 1

	for col, val := range pkValues {
		if conn.GetDBType() == db.PostgreSQL {
			whereClauses = append(whereClauses, fmt.Sprintf("\"%s\" = $%d", col, argIndex))
		} else {
			whereClauses = append(whereClauses, fmt.Sprintf("`%s` = ?", col))
		}
		args = append(args, val)
		argIndex++
	}

	var query string
	if conn.GetDBType() == db.PostgreSQL {
		query = fmt.Sprintf("DELETE FROM \"%s\" WHERE %s",
			tableName,
			strings.Join(whereClauses, " AND "))
	} else {
		query = fmt.Sprintf("DELETE FROM `%s` WHERE %s",
			tableName,
			strings.Join(whereClauses, " AND "))
	}

	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete row: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("no rows were deleted")
	}

	return nil
}

// BatchOps represents a set of batch operations for a single table
type BatchOps struct {
	Updates []BatchUpdate `json:"updates"`
	Inserts []BatchInsert `json:"inserts"`
	Deletes []BatchDelete `json:"deletes"`
}

// BatchUpdate represents a single UPDATE operation
type BatchUpdate struct {
	PkValues map[string]interface{} `json:"pkValues"`
	Updates  map[string]interface{} `json:"updates"`
}

// BatchInsert represents a single INSERT operation
type BatchInsert struct {
	Values map[string]interface{} `json:"values"`
}

// BatchDelete represents a single DELETE operation
type BatchDelete struct {
	PkValues map[string]interface{} `json:"pkValues"`
}

// BatchResult holds the result of a batch modify operation
type BatchResult struct {
	Updated int `json:"updated"`
	Inserted int `json:"inserted"`
	Deleted int `json:"deleted"`
}

// BatchModify executes multiple UPDATE/INSERT/DELETE operations within a single
// database transaction. If any operation fails, the entire batch is rolled back.
func (ts *TableService) BatchModify(conn db.DBConnection, dbName, tableName string, ops BatchOps) (*BatchResult, error) {
	if len(ops.Updates) == 0 && len(ops.Inserts) == 0 && len(ops.Deletes) == 0 {
		return &BatchResult{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return nil, fmt.Errorf("failed to select database: %w", err)
		}
	}

	dbConn := conn.GetDB()
	if dbConn == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // no-op if already committed

	isPostgres := conn.GetDBType() == db.PostgreSQL
	result := &BatchResult{}

	// Helper: quote identifier
	quoteId := func(name string) string {
		if isPostgres {
			return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
		}
		return fmt.Sprintf("`%s`", strings.ReplaceAll(name, "`", "``"))
	}

	// Helper: placeholder
	placeholder := func(idx int) string {
		if isPostgres {
			return fmt.Sprintf("$%d", idx)
		}
		return "?"
	}

	// ── Execute UPDATEs ──
	for i, upd := range ops.Updates {
		if len(upd.PkValues) == 0 || len(upd.Updates) == 0 {
			continue
		}
		var setClauses []string
		var args []interface{}
		argIdx := 1

		for col, val := range upd.Updates {
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", quoteId(col), placeholder(argIdx)))
			args = append(args, val)
			argIdx++
		}
		var whereClauses []string
		for col, val := range upd.PkValues {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = %s", quoteId(col), placeholder(argIdx)))
			args = append(args, val)
			argIdx++
		}

		query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
			quoteId(tableName),
			strings.Join(setClauses, ", "),
			strings.Join(whereClauses, " AND "))

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return nil, fmt.Errorf("update row %d failed: %w", i+1, err)
		}
		result.Updated++
	}

	// ── Execute INSERTs ──
	for i, ins := range ops.Inserts {
		if len(ins.Values) == 0 {
			continue
		}
		var columns []string
		var placeholders []string
		var args []interface{}
		argIdx := 1

		for col, val := range ins.Values {
			columns = append(columns, quoteId(col))
			placeholders = append(placeholders, placeholder(argIdx))
			args = append(args, val)
			argIdx++
		}

		query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			quoteId(tableName),
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "))

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return nil, fmt.Errorf("insert row %d failed: %w", i+1, err)
		}
		result.Inserted++
	}

	// ── Execute DELETEs ──
	for i, del := range ops.Deletes {
		if len(del.PkValues) == 0 {
			continue
		}
		var whereClauses []string
		var args []interface{}
		argIdx := 1

		for col, val := range del.PkValues {
			whereClauses = append(whereClauses, fmt.Sprintf("%s = %s", quoteId(col), placeholder(argIdx)))
			args = append(args, val)
			argIdx++
		}

		query := fmt.Sprintf("DELETE FROM %s WHERE %s",
			quoteId(tableName),
			strings.Join(whereClauses, " AND "))

		if _, err := tx.ExecContext(ctx, query, args...); err != nil {
			return nil, fmt.Errorf("delete row %d failed: %w", i+1, err)
		}
		result.Deleted++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}
