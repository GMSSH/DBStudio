package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"regexp"
	"sqlmanager/pkg/db"
	"strings"
	"time"
)

// ExportFormat represents the export format
type ExportFormat string

const (
	ExportCSV       ExportFormat = "csv"
	ExportJSON      ExportFormat = "json"
	ExportSQLInsert ExportFormat = "sql"
)

// ExportResult contains the exported data as a string
type ExportResult struct {
	Data     string `json:"data"`
	FileName string `json:"fileName"`
	Format   string `json:"format"`
	RowCount int    `json:"rowCount"`
}

// ImportResult contains the result of an import operation
type ImportResult struct {
	Imported int      `json:"imported"`
	Errors   []string `json:"errors"`
}

// ImportExportService handles data import and export
type ImportExportService struct{}

func NewImportExportService() *ImportExportService {
	return &ImportExportService{}
}

// ExportTable exports table data in the specified format
func (s *ImportExportService) ExportTable(conn db.DBConnection, dbName, tableName string, format ExportFormat, limit int) (*ExportResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return nil, fmt.Errorf("failed to select database: %w", err)
		}
	}

	// Build query
	if limit <= 0 || limit > 100000 {
		limit = 100000
	}

	var fullTableName string
	switch conn.GetDBType() {
	case db.MySQL:
		fullTableName = fmt.Sprintf("`%s`.`%s`", escapeMysql(dbName), escapeMysql(tableName))
	case db.PostgreSQL:
		fullTableName = fmt.Sprintf(`"%s"`, escapePg(tableName))
	case db.SQLite:
		fullTableName = fmt.Sprintf("`%s`", escapeMysql(tableName))
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", fullTableName, limit)
	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var data [][]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make([]interface{}, len(columns))
		for i, v := range values {
			row[i] = safeConvertValue(v)
		}
		data = append(data, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var output string
	var ext string

	switch format {
	case ExportCSV:
		output, err = s.toCSV(columns, data)
		ext = "csv"
	case ExportJSON:
		output, err = s.toJSON(columns, data)
		ext = "json"
	case ExportSQLInsert:
		output, err = s.toSQLInsert(tableName, columns, data, conn.GetDBType())
		ext = "sql"
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}

	if err != nil {
		return nil, err
	}

	return &ExportResult{
		Data:     output,
		FileName: fmt.Sprintf("%s_%s.%s", tableName, time.Now().Format("20060102_150405"), ext),
		Format:   string(format),
		RowCount: len(data),
	}, nil
}

// ExportTableDDL exports the CREATE TABLE DDL
func (s *ImportExportService) ExportTableDDL(conn db.DBConnection, dbName, tableName string) (*ExportResult, error) {
	ds := NewTableDesignerService()
	ddl, err := ds.GetTableDDL(conn, dbName, tableName)
	if err != nil {
		return nil, err
	}

	return &ExportResult{
		Data:     ddl,
		FileName: fmt.Sprintf("%s_schema.sql", tableName),
		Format:   "sql",
		RowCount: 0,
	}, nil
}

// ImportCSV imports CSV data into a table
func (s *ImportExportService) ImportCSV(conn db.DBConnection, dbName, tableName string, csvData string, mapping map[string]string, headerRow bool) (*ImportResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return nil, fmt.Errorf("failed to select database: %w", err)
		}
	}

	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 {
		return &ImportResult{Imported: 0}, nil
	}

	var sourceColumns []string
	startRow := 0
	if headerRow && len(records) > 0 {
		sourceColumns = records[0]
		startRow = 1
	} else {
		sourceColumns = make([]string, len(records[0]))
		for i := range records[0] {
			sourceColumns[i] = fmt.Sprintf("col%d", i+1)
		}
	}

	if len(mapping) == 0 {
		mapping = make(map[string]string, len(sourceColumns))
		for _, col := range sourceColumns {
			mapping[col] = col
		}
	}

	result := &ImportResult{}
	dbType := conn.GetDBType()

	for i := startRow; i < len(records); i++ {
		row := records[i]
		if len(row) != len(sourceColumns) {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: column count mismatch (%d vs %d)", i+1, len(row), len(sourceColumns)))
			continue
		}

		// Build INSERT statement
		var colParts []string
		var placeholders []string
		var args []interface{}

		for j, sourceColumn := range sourceColumns {
			targetColumn := strings.TrimSpace(mapping[sourceColumn])
			if targetColumn == "" {
				continue
			}

			switch dbType {
			case db.MySQL, db.SQLite:
				colParts = append(colParts, fmt.Sprintf("`%s`", escapeMysql(targetColumn)))
				placeholders = append(placeholders, "?")
			case db.PostgreSQL:
				colParts = append(colParts, fmt.Sprintf(`"%s"`, escapePg(targetColumn)))
				placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
			}
			val := row[j]
			if val == "" || strings.ToUpper(val) == "NULL" {
				args = append(args, nil)
			} else {
				args = append(args, val)
			}
		}

		if len(colParts) == 0 {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: no mapped columns", i+1))
			continue
		}

		var query string
		switch dbType {
		case db.MySQL:
			query = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
				escapeMysql(tableName), strings.Join(colParts, ", "), strings.Join(placeholders, ", "))
		case db.PostgreSQL:
			query = fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES (%s)`,
				escapePg(tableName), strings.Join(colParts, ", "), strings.Join(placeholders, ", "))
		case db.SQLite:
			query = fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
				escapeMysql(tableName), strings.Join(colParts, ", "), strings.Join(placeholders, ", "))
		}

		if _, err := conn.Exec(ctx, query, args...); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: %v", i+1, err))
		} else {
			result.Imported++
		}
	}

	return result, nil
}

// ImportTableSQL imports SQL data into the current table.
// Only INSERT INTO statements that target the current table are allowed.
func (s *ImportExportService) ImportTableSQL(conn db.DBConnection, dbName, tableName string, sqlData string) (*ImportResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if strings.TrimSpace(sqlData) == "" {
		return nil, fmt.Errorf("SQL data is required")
	}

	statements, err := splitSQLStatements(sqlData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}
	if len(statements) == 0 {
		return &ImportResult{Imported: 0}, nil
	}

	targetTable := normalizeIdentifier(tableName)
	targetDatabase := normalizeIdentifier(dbName)

	for index, stmt := range statements {
		stmtDatabase, stmtTable, valid := parseInsertTarget(stmt)
		if !valid {
			return nil, fmt.Errorf("statement %d is not a valid INSERT INTO for table %s", index+1, tableName)
		}
		if stmtTable != targetTable {
			return nil, fmt.Errorf("statement %d targets table %s, expected %s", index+1, stmtTable, targetTable)
		}
		if stmtDatabase != "" && targetDatabase != "" && stmtDatabase != targetDatabase {
			return nil, fmt.Errorf("statement %d targets database %s, expected %s", index+1, stmtDatabase, targetDatabase)
		}
	}

	execConn, err := conn.GetDB().Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL connection: %w", err)
	}
	defer execConn.Close()

	if dbName != "" && conn.GetDBType() == db.MySQL {
		if _, err := execConn.ExecContext(ctx, fmt.Sprintf("USE `%s`", escapeMysql(dbName))); err != nil {
			return nil, fmt.Errorf("failed to select database: %w", err)
		}
	}

	result := &ImportResult{}
	for index, stmt := range statements {
		execResult, err := execConn.ExecContext(ctx, stmt)
		if err != nil {
			return nil, fmt.Errorf("statement %d failed: %w", index+1, err)
		}

		if affected, affectedErr := execResult.RowsAffected(); affectedErr == nil && affected > 0 {
			result.Imported += int(affected)
		} else {
			result.Imported++
		}
	}

	return result, nil
}

// --- Format helpers ---

func (s *ImportExportService) toCSV(columns []string, data [][]interface{}) (string, error) {
	var sb strings.Builder
	w := csv.NewWriter(&sb)

	if err := w.Write(columns); err != nil {
		return "", err
	}

	for _, row := range data {
		record := make([]string, len(row))
		for i, v := range row {
			if v == nil {
				record[i] = ""
			} else {
				record[i] = fmt.Sprintf("%v", v)
			}
		}
		if err := w.Write(record); err != nil {
			return "", err
		}
	}
	w.Flush()
	return sb.String(), w.Error()
}

func (s *ImportExportService) toJSON(columns []string, data [][]interface{}) (string, error) {
	result := make([]map[string]interface{}, 0, len(data))
	for _, row := range data {
		m := make(map[string]interface{}, len(columns))
		for i, col := range columns {
			m[col] = row[i]
		}
		result = append(result, m)
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *ImportExportService) toSQLInsert(tableName string, columns []string, data [][]interface{}, dbType db.DBType) (string, error) {
	if len(data) == 0 {
		return fmt.Sprintf("-- No data in table %s\n", tableName), nil
	}

	var sb strings.Builder

	var quotedTable string
	var quoteCol func(string) string

	switch dbType {
	case db.MySQL:
		quotedTable = fmt.Sprintf("`%s`", escapeMysql(tableName))
		quoteCol = func(c string) string { return fmt.Sprintf("`%s`", escapeMysql(c)) }
	case db.PostgreSQL:
		quotedTable = fmt.Sprintf(`"%s"`, escapePg(tableName))
		quoteCol = func(c string) string { return fmt.Sprintf(`"%s"`, escapePg(c)) }
	default:
		quotedTable = fmt.Sprintf("`%s`", escapeMysql(tableName))
		quoteCol = func(c string) string { return fmt.Sprintf("`%s`", escapeMysql(c)) }
	}

	colParts := make([]string, len(columns))
	for i, c := range columns {
		colParts[i] = quoteCol(c)
	}
	colList := strings.Join(colParts, ", ")

	for _, row := range data {
		vals := make([]string, len(row))
		for i, v := range row {
			if v == nil {
				vals[i] = "NULL"
			} else {
				str := fmt.Sprintf("%v", v)
				escaped := strings.ReplaceAll(str, "'", "''")
				vals[i] = fmt.Sprintf("'%s'", escaped)
			}
		}
		sb.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s);\n",
			quotedTable, colList, strings.Join(vals, ", ")))
	}

	return sb.String(), nil
}

var insertIntoPattern = regexp.MustCompile(`(?is)^\s*INSERT\s+INTO\s+((?:` + "`[^`]+`" + `|"(?:[^"]|"")+"|[A-Za-z0-9_]+)(?:\s*\.\s*(?:` + "`[^`]+`" + `|"(?:[^"]|"")+"|[A-Za-z0-9_]+))?)`)

func splitSQLStatements(sqlText string) ([]string, error) {
	var statements []string
	var current strings.Builder
	var quote rune
	inLineComment := false
	inBlockComment := false

	flush := func() {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
		current.Reset()
	}

	for i := 0; i < len(sqlText); i++ {
		ch := rune(sqlText[i])
		var next rune
		if i+1 < len(sqlText) {
			next = rune(sqlText[i+1])
		}

		if inLineComment {
			if ch == '\n' {
				inLineComment = false
				current.WriteRune(ch)
			}
			continue
		}

		if inBlockComment {
			if ch == '*' && next == '/' {
				inBlockComment = false
				i++
			}
			continue
		}

		if quote != 0 {
			current.WriteRune(ch)
			if ch == quote {
				if quote == '\'' && next == '\'' {
					current.WriteRune(next)
					i++
					continue
				}
				if quote == '"' && next == '"' {
					current.WriteRune(next)
					i++
					continue
				}
				quote = 0
			}
			if ch == '\\' && next != 0 {
				current.WriteRune(next)
				i++
			}
			continue
		}

		if ch == '-' && next == '-' {
			inLineComment = true
			i++
			continue
		}
		if ch == '#' {
			inLineComment = true
			continue
		}
		if ch == '/' && next == '*' {
			inBlockComment = true
			i++
			continue
		}

		if ch == '\'' || ch == '"' || ch == '`' {
			quote = ch
			current.WriteRune(ch)
			continue
		}

		if ch == ';' {
			flush()
			continue
		}

		current.WriteRune(ch)
	}

	if quote != 0 || inBlockComment {
		return nil, fmt.Errorf("unterminated SQL string or comment")
	}

	flush()
	return statements, nil
}

func parseInsertTarget(stmt string) (string, string, bool) {
	matches := insertIntoPattern.FindStringSubmatch(stmt)
	if len(matches) < 2 {
		return "", "", false
	}

	parts := strings.Split(matches[1], ".")
	if len(parts) == 1 {
		return "", normalizeIdentifier(parts[0]), true
	}
	if len(parts) == 2 {
		return normalizeIdentifier(parts[0]), normalizeIdentifier(parts[1]), true
	}
	return "", "", false
}

func normalizeIdentifier(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) >= 2 {
		first := trimmed[0]
		last := trimmed[len(trimmed)-1]
		if (first == '`' && last == '`') || (first == '"' && last == '"') {
			trimmed = trimmed[1 : len(trimmed)-1]
		}
	}
	return strings.ToLower(strings.TrimSpace(trimmed))
}
