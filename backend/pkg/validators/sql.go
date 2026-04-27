package validators

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	// Valid identifier pattern: alphanumeric and underscore only
	identifierPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// ValidateIdentifier validates SQL identifiers (table names, database names, column names)
// to prevent SQL injection attacks
func ValidateIdentifier(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("identifier cannot be empty")
	}
	
	if len(name) > 64 {
		return fmt.Errorf("identifier too long (max 64 characters)")
	}
	
	// Only allow alphanumeric and underscore
	if !identifierPattern.MatchString(name) {
		return fmt.Errorf("identifier contains invalid characters (only a-z, A-Z, 0-9, _ allowed)")
	}
	
	// Prevent reserved keywords (common ones)
	reserved := []string{
		"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "CREATE",
		"ALTER", "TABLE", "DATABASE", "INDEX", "VIEW", "TRIGGER",
	}
	upperName := strings.ToUpper(name)
	for _, keyword := range reserved {
		if upperName == keyword {
			return fmt.Errorf("identifier cannot be a reserved keyword: %s", name)
		}
	}
	
	return nil
}

// EscapeIdentifier escapes SQL identifiers for safe use in queries
// This is a secondary defense layer - validation should be done first
func EscapeIdentifier(name string, dbType string) string {
	switch dbType {
	case "mysql":
		// Escape backticks by doubling them
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	case "postgres":
		// Escape double quotes by doubling them
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	case "sqlite":
		// SQLite uses double quotes
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	default:
		return name
	}
}

// ValidateIdentifiers validates multiple identifiers at once
func ValidateIdentifiers(names ...string) error {
	for _, name := range names {
		if err := ValidateIdentifier(name); err != nil {
			return err
		}
	}
	return nil
}
