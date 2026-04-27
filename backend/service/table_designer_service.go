package service

import (
	"context"
	"fmt"
	"sqlmanager/pkg/db"
	"strings"
	"time"
)

// ColumnDef represents a column definition for table creation/alteration
type ColumnDef struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Length       string `json:"length,omitempty"`
	NotNull      bool   `json:"notNull"`
	DefaultValue string `json:"defaultValue,omitempty"`
	Comment      string `json:"comment,omitempty"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
	AutoIncrement bool  `json:"autoIncrement"`
	Unsigned     bool   `json:"unsigned"`
}

// IndexDef represents an index definition
type IndexDef struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
}

// TableDef represents a complete table definition for creation
type TableDef struct {
	Name    string      `json:"name"`
	Columns []ColumnDef `json:"columns"`
	Indexes []IndexDef  `json:"indexes"`
	Engine  string      `json:"engine,omitempty"`  // MySQL: InnoDB etc.
	Charset string      `json:"charset,omitempty"` // MySQL: utf8mb4 etc.
	Comment string      `json:"comment,omitempty"`
}

// AlterOp represents a single ALTER TABLE operation
type AlterOp struct {
	Action     string    `json:"action"` // "addColumn"|"modifyColumn"|"dropColumn"|"renameColumn"|"addIndex"|"dropIndex"
	Column     ColumnDef `json:"column,omitempty"`
	OldName    string    `json:"oldName,omitempty"`
	Index      IndexDef  `json:"index,omitempty"`
}

// TableDesignerService handles DDL operations
type TableDesignerService struct{}

func NewTableDesignerService() *TableDesignerService {
	return &TableDesignerService{}
}

// CreateTable creates a new table from a TableDef
func (s *TableDesignerService) CreateTable(conn db.DBConnection, dbName string, def TableDef) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	ddl, err := s.BuildCreateDDL(conn.GetDBType(), def)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, ddl)
	return err
}

// AlterTable executes a series of ALTER TABLE operations
func (s *TableDesignerService) AlterTable(conn db.DBConnection, dbName, tableName string, ops []AlterOp) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	dbType := conn.GetDBType()
	for _, op := range ops {
		stmt, err := s.buildAlterStatement(dbType, tableName, op)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("alter failed (%s): %w", op.Action, err)
		}
	}
	return nil
}

// DropTable drops a table
func (s *TableDesignerService) DropTable(conn db.DBConnection, dbName, tableName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	var query string
	switch conn.GetDBType() {
	case db.MySQL:
		query = fmt.Sprintf("DROP TABLE IF EXISTS `%s`", escapeMysql(tableName))
	case db.PostgreSQL:
		query = fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, escapePg(tableName))
	case db.SQLite:
		query = fmt.Sprintf("DROP TABLE IF EXISTS `%s`", escapeMysql(tableName))
	default:
		return fmt.Errorf("unsupported database type")
	}

	_, err := conn.Exec(ctx, query)
	return err
}

// RenameTable renames a table
func (s *TableDesignerService) RenameTable(conn db.DBConnection, dbName, oldName, newName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return fmt.Errorf("failed to select database: %w", err)
		}
	}

	var query string
	switch conn.GetDBType() {
	case db.MySQL:
		query = fmt.Sprintf("RENAME TABLE `%s` TO `%s`", escapeMysql(oldName), escapeMysql(newName))
	case db.PostgreSQL:
		query = fmt.Sprintf(`ALTER TABLE "%s" RENAME TO "%s"`, escapePg(oldName), escapePg(newName))
	case db.SQLite:
		query = fmt.Sprintf("ALTER TABLE `%s` RENAME TO `%s`", escapeMysql(oldName), escapeMysql(newName))
	default:
		return fmt.Errorf("unsupported database type")
	}

	_, err := conn.Exec(ctx, query)
	return err
}

// GetTableDDL returns the CREATE TABLE DDL for a table
func (s *TableDesignerService) GetTableDDL(conn db.DBConnection, dbName, tableName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if dbName != "" {
		if err := conn.SelectDatabase(dbName); err != nil {
			return "", fmt.Errorf("failed to select database: %w", err)
		}
	}

	switch conn.GetDBType() {
	case db.MySQL:
		currentDB := dbName
		if currentDB == "" {
			currentDB = conn.GetConfig().Database
		}

		var objectType string
		typeRow := conn.QueryRow(ctx,
			"SELECT table_type FROM information_schema.tables WHERE table_schema = ? AND table_name = ? LIMIT 1",
			currentDB,
			tableName,
		)
		if err := typeRow.Scan(&objectType); err != nil {
			return "", fmt.Errorf("failed to detect object type: %w", err)
		}

		if objectType == "VIEW" {
			row := conn.QueryRow(ctx, fmt.Sprintf("SHOW CREATE VIEW `%s`", escapeMysql(tableName)))
			var name, ddl, charsetClient, collationConnection string
			if err := row.Scan(&name, &ddl, &charsetClient, &collationConnection); err != nil {
				return "", fmt.Errorf("failed to get view DDL: %w", err)
			}
			return ddl, nil
		}

		row := conn.QueryRow(ctx, fmt.Sprintf("SHOW CREATE TABLE `%s`", escapeMysql(tableName)))
		var name, ddl string
		if err := row.Scan(&name, &ddl); err != nil {
			return "", fmt.Errorf("failed to get table DDL: %w", err)
		}
		return ddl, nil

	case db.PostgreSQL:
		rows, err := conn.Query(ctx, `
			SELECT table_type
			FROM information_schema.tables
			WHERE table_catalog = $1 AND table_schema = current_schema() AND table_name = $2
			LIMIT 1
		`, dbName, tableName)
		if err != nil {
			return "", fmt.Errorf("failed to detect object type: %w", err)
		}
		defer rows.Close()

		var objectType string
		if rows.Next() {
			if err := rows.Scan(&objectType); err != nil {
				return "", fmt.Errorf("failed to read object type: %w", err)
			}
		}

		if objectType == "VIEW" {
			viewRow := conn.QueryRow(ctx, `SELECT pg_get_viewdef($1::regclass, true)`, tableName)
			var ddl string
			if err := viewRow.Scan(&ddl); err != nil {
				return "", fmt.Errorf("failed to get view DDL: %w", err)
			}
			return fmt.Sprintf("CREATE OR REPLACE VIEW %s AS\n%s", tableName, ddl), nil
		}

		// Build DDL from information_schema for PostgreSQL tables
		ts := NewTableService()
		schema, err := ts.GetTableSchema(conn, dbName, tableName)
		if err != nil {
			return "", err
		}
		return s.buildPostgresDDL(tableName, schema), nil

	case db.SQLite:
		rows, err := conn.Query(ctx, "SELECT sql FROM sqlite_master WHERE type IN ('table', 'view') AND name=?", tableName)
		if err != nil {
			return "", err
		}
		defer rows.Close()
		if rows.Next() {
			var ddl string
			rows.Scan(&ddl)
			return ddl, nil
		}
		return "", fmt.Errorf("table not found: %s", tableName)
	}
	return "", fmt.Errorf("unsupported database type")
}

// BuildCreateDDL generates a CREATE TABLE statement from a TableDef
func (s *TableDesignerService) BuildCreateDDL(dbType db.DBType, def TableDef) (string, error) {
	if def.Name == "" {
		return "", fmt.Errorf("table name is required")
	}
	if len(def.Columns) == 0 {
		return "", fmt.Errorf("at least one column is required")
	}

	var sb strings.Builder
	var primaryKeys []string

	switch dbType {
	case db.MySQL:
		sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", escapeMysql(def.Name)))
		colDefs := []string{}
		for _, col := range def.Columns {
			colDefs = append(colDefs, s.mysqlColumnDef(col))
			if col.IsPrimaryKey {
				primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", escapeMysql(col.Name)))
			}
		}
		if len(primaryKeys) > 0 {
			colDefs = append(colDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
		}
		for _, idx := range def.Indexes {
			cols := []string{}
			for _, c := range idx.Columns {
				cols = append(cols, fmt.Sprintf("`%s`", escapeMysql(c)))
			}
			unique := ""
			if idx.Unique {
				unique = "UNIQUE "
			}
			colDefs = append(colDefs, fmt.Sprintf("  %sKEY `%s` (%s)", unique, escapeMysql(idx.Name), strings.Join(cols, ", ")))
		}
		sb.WriteString(strings.Join(colDefs, ",\n"))
		sb.WriteString("\n)")
		engine := def.Engine
		if engine == "" {
			engine = "InnoDB"
		}
		charset := def.Charset
		if charset == "" {
			charset = "utf8mb4"
		}
		sb.WriteString(fmt.Sprintf(" ENGINE=%s DEFAULT CHARSET=%s", engine, charset))
		if def.Comment != "" {
			sb.WriteString(fmt.Sprintf(" COMMENT='%s'", strings.ReplaceAll(def.Comment, "'", "''")))
		}

	case db.PostgreSQL:
		sb.WriteString(fmt.Sprintf(`CREATE TABLE "%s" (`+"\n", escapePg(def.Name)))
		colDefs := []string{}
		for _, col := range def.Columns {
			colDefs = append(colDefs, s.pgColumnDef(col))
			if col.IsPrimaryKey {
				primaryKeys = append(primaryKeys, fmt.Sprintf(`"%s"`, escapePg(col.Name)))
			}
		}
		if len(primaryKeys) > 0 {
			colDefs = append(colDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
		}
		sb.WriteString(strings.Join(colDefs, ",\n"))
		sb.WriteString("\n)")
		// Indexes are created separately in PostgreSQL
		for _, idx := range def.Indexes {
			cols := []string{}
			for _, c := range idx.Columns {
				cols = append(cols, fmt.Sprintf(`"%s"`, escapePg(c)))
			}
			unique := ""
			if idx.Unique {
				unique = "UNIQUE "
			}
			sb.WriteString(fmt.Sprintf(";\nCREATE %sINDEX \"%s\" ON \"%s\" (%s)",
				unique, escapePg(idx.Name), escapePg(def.Name), strings.Join(cols, ", ")))
		}

	case db.SQLite:
		sb.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", escapeMysql(def.Name)))
		colDefs := []string{}
		for _, col := range def.Columns {
			colDefs = append(colDefs, s.sqliteColumnDef(col))
			if col.IsPrimaryKey {
				primaryKeys = append(primaryKeys, fmt.Sprintf("`%s`", escapeMysql(col.Name)))
			}
		}
		if len(primaryKeys) > 1 {
			colDefs = append(colDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
		}
		sb.WriteString(strings.Join(colDefs, ",\n"))
		sb.WriteString("\n)")

	default:
		return "", fmt.Errorf("unsupported database type")
	}

	return sb.String(), nil
}

// --- Internal helpers ---

func (s *TableDesignerService) mysqlColumnDef(col ColumnDef) string {
	typePart := col.Type
	if col.Length != "" {
		typePart += "(" + col.Length + ")"
	}
	if col.Unsigned {
		typePart += " UNSIGNED"
	}
	parts := []string{fmt.Sprintf("  `%s` %s", escapeMysql(col.Name), typePart)}
	if col.NotNull {
		parts = append(parts, "NOT NULL")
	} else {
		parts = append(parts, "NULL")
	}
	if col.AutoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}
	if col.DefaultValue != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT '%s'", strings.ReplaceAll(col.DefaultValue, "'", "''")))
	}
	if col.Comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''")))
	}
	return strings.Join(parts, " ")
}

func (s *TableDesignerService) pgColumnDef(col ColumnDef) string {
	typePart := col.Type
	if col.AutoIncrement && (strings.ToLower(col.Type) == "integer" || strings.ToLower(col.Type) == "int") {
		typePart = "SERIAL"
	} else if col.Length != "" {
		typePart += "(" + col.Length + ")"
	}
	parts := []string{fmt.Sprintf(`  "%s" %s`, escapePg(col.Name), typePart)}
	if col.NotNull {
		parts = append(parts, "NOT NULL")
	}
	if col.DefaultValue != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT '%s'", strings.ReplaceAll(col.DefaultValue, "'", "''")))
	}
	return strings.Join(parts, " ")
}

func (s *TableDesignerService) sqliteColumnDef(col ColumnDef) string {
	parts := []string{fmt.Sprintf("  `%s` %s", escapeMysql(col.Name), col.Type)}
	if col.IsPrimaryKey && col.AutoIncrement {
		parts = append(parts, "PRIMARY KEY AUTOINCREMENT")
	} else if col.IsPrimaryKey {
		parts = append(parts, "PRIMARY KEY")
	}
	if col.NotNull && !col.IsPrimaryKey {
		parts = append(parts, "NOT NULL")
	}
	if col.DefaultValue != "" {
		parts = append(parts, fmt.Sprintf("DEFAULT '%s'", strings.ReplaceAll(col.DefaultValue, "'", "''")))
	}
	return strings.Join(parts, " ")
}

func (s *TableDesignerService) buildAlterStatement(dbType db.DBType, tableName string, op AlterOp) (string, error) {
	switch dbType {
	case db.MySQL:
		return s.mysqlAlterStatement(tableName, op)
	case db.PostgreSQL:
		return s.pgAlterStatement(tableName, op)
	case db.SQLite:
		return s.sqliteAlterStatement(tableName, op)
	}
	return "", fmt.Errorf("unsupported database type")
}

func (s *TableDesignerService) mysqlAlterStatement(tableName string, op AlterOp) (string, error) {
	tbl := fmt.Sprintf("`%s`", escapeMysql(tableName))
	switch op.Action {
	case "addColumn":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tbl, s.mysqlColumnDef(op.Column)), nil
	case "modifyColumn":
		return fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s", tbl, s.mysqlColumnDef(op.Column)), nil
	case "dropColumn":
		return fmt.Sprintf("ALTER TABLE %s DROP COLUMN `%s`", tbl, escapeMysql(op.Column.Name)), nil
	case "renameColumn":
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN `%s` TO `%s`", tbl, escapeMysql(op.OldName), escapeMysql(op.Column.Name)), nil
	case "addIndex":
		cols := []string{}
		for _, c := range op.Index.Columns {
			cols = append(cols, fmt.Sprintf("`%s`", escapeMysql(c)))
		}
		unique := ""
		if op.Index.Unique {
			unique = "UNIQUE "
		}
		return fmt.Sprintf("ALTER TABLE %s ADD %sINDEX `%s` (%s)", tbl, unique, escapeMysql(op.Index.Name), strings.Join(cols, ", ")), nil
	case "dropIndex":
		return fmt.Sprintf("ALTER TABLE %s DROP INDEX `%s`", tbl, escapeMysql(op.Index.Name)), nil
	}
	return "", fmt.Errorf("unknown alter action: %s", op.Action)
}

func (s *TableDesignerService) pgAlterStatement(tableName string, op AlterOp) (string, error) {
	tbl := fmt.Sprintf(`"%s"`, escapePg(tableName))
	switch op.Action {
	case "addColumn":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tbl, s.pgColumnDef(op.Column)), nil
	case "modifyColumn":
		// PostgreSQL uses ALTER COLUMN for type changes
		typePart := op.Column.Type
		if op.Column.Length != "" {
			typePart += "(" + op.Column.Length + ")"
		}
		return fmt.Sprintf(`ALTER TABLE %s ALTER COLUMN "%s" TYPE %s`, tbl, escapePg(op.Column.Name), typePart), nil
	case "dropColumn":
		return fmt.Sprintf(`ALTER TABLE %s DROP COLUMN "%s"`, tbl, escapePg(op.Column.Name)), nil
	case "renameColumn":
		return fmt.Sprintf(`ALTER TABLE %s RENAME COLUMN "%s" TO "%s"`, tbl, escapePg(op.OldName), escapePg(op.Column.Name)), nil
	case "addIndex":
		cols := []string{}
		for _, c := range op.Index.Columns {
			cols = append(cols, fmt.Sprintf(`"%s"`, escapePg(c)))
		}
		unique := ""
		if op.Index.Unique {
			unique = "UNIQUE "
		}
		return fmt.Sprintf(`CREATE %sINDEX "%s" ON %s (%s)`, unique, escapePg(op.Index.Name), tbl, strings.Join(cols, ", ")), nil
	case "dropIndex":
		return fmt.Sprintf(`DROP INDEX IF EXISTS "%s"`, escapePg(op.Index.Name)), nil
	}
	return "", fmt.Errorf("unknown alter action: %s", op.Action)
}

func (s *TableDesignerService) sqliteAlterStatement(tableName string, op AlterOp) (string, error) {
	tbl := fmt.Sprintf("`%s`", escapeMysql(tableName))
	switch op.Action {
	case "addColumn":
		return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s", tbl, s.sqliteColumnDef(op.Column)), nil
	case "renameColumn":
		return fmt.Sprintf("ALTER TABLE %s RENAME COLUMN `%s` TO `%s`", tbl, escapeMysql(op.OldName), escapeMysql(op.Column.Name)), nil
	}
	return "", fmt.Errorf("SQLite only supports addColumn and renameColumn via ALTER TABLE")
}

func (s *TableDesignerService) buildPostgresDDL(tableName string, schema *TableSchema) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`CREATE TABLE "%s" (`+"\n", tableName))
	colDefs := []string{}
	for _, col := range schema.Columns {
		def := fmt.Sprintf(`  "%s" %s`, col.Name, col.Type)
		if !col.Nullable {
			def += " NOT NULL"
		}
		if col.DefaultValue != "" {
			def += " DEFAULT " + col.DefaultValue
		}
		colDefs = append(colDefs, def)
	}
	if len(schema.PrimaryKey) > 0 {
		pks := []string{}
		for _, pk := range schema.PrimaryKey {
			pks = append(pks, fmt.Sprintf(`"%s"`, pk))
		}
		colDefs = append(colDefs, fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(pks, ", ")))
	}
	sb.WriteString(strings.Join(colDefs, ",\n"))
	sb.WriteString("\n)")
	return sb.String()
}

// --- Escape helpers ---

func escapeMysql(s string) string {
	return strings.ReplaceAll(s, "`", "``")
}

func escapePg(s string) string {
	return strings.ReplaceAll(s, `"`, `""`)
}
