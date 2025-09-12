package parse

import (
	"crypto/md5"
	"fmt"
	"regexp"
	"strings"
)

type QueryParser struct {
	normalizationRules []NormalizationRule
}

type NormalizationRule struct {
	Pattern     *regexp.Regexp
	Replacement string
}

func NewQueryParser() *QueryParser {
	rules := []NormalizationRule{
		{regexp.MustCompile(`\$\d+`), "?"},
		{regexp.MustCompile(`'[^']*'`), "?"},
		{regexp.MustCompile(`\b\d+\b`), "?"},
		{regexp.MustCompile(`\s+`), " "},
		{regexp.MustCompile(`\(\s*\?\s*(,\s*\?\s*)*\)`), "(?)"},
		{regexp.MustCompile(`IN\s*\(\s*\?\s*(,\s*\?\s*)*\)`), "IN (?)"},
		{regexp.MustCompile(`VALUES\s*\(\s*\?\s*(,\s*\?\s*)*\)`), "VALUES (?)"},
	}

	return &QueryParser{normalizationRules: rules}
}

func (qp *QueryParser) NormalizeQuery(query string) string {
	normalized := strings.TrimSpace(query)
	normalized = strings.ToUpper(normalized)

	for _, rule := range qp.normalizationRules {
		normalized = rule.Pattern.ReplaceAllString(normalized, rule.Replacement)
	}

	return strings.TrimSpace(normalized)
}

func (qp *QueryParser) GenerateFingerprint(query string) string {
	normalized := qp.NormalizeQuery(query)
	hash := md5.Sum([]byte(normalized))
	return fmt.Sprintf("%x", hash)
}

func (qp *QueryParser) ExtractTables(query string) []string {
	var tables []string

	// Simple regex to find table names after FROM and JOIN
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		regexp.MustCompile(`(?i)\bJOIN\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		regexp.MustCompile(`(?i)\bINTO\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		regexp.MustCompile(`(?i)\bUPDATE\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(query, -1)
		for _, match := range matches {
			if len(match) > 1 {
				table := strings.ToLower(match[1])
				if !contains(tables, table) {
					tables = append(tables, table)
				}
			}
		}
	}

	return tables
}

func (qp *QueryParser) DetectQueryType(query string) string {
	query = strings.ToUpper(strings.TrimSpace(query))

	if strings.HasPrefix(query, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(query, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(query, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(query, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(query, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(query, "ALTER") {
		return "ALTER"
	} else if strings.HasPrefix(query, "DROP") {
		return "DROP"
	}

	return "OTHER"
}

func (qp *QueryParser) HasSequentialScan(query string) bool {
	// Simple heuristics to detect potential seq scans
	query = strings.ToUpper(query)

	// Look for patterns that often cause seq scans
	seqScanPatterns := []string{
		"WHERE.*LIKE",      // LIKE without leading wildcard
		"WHERE.*!=",        // Inequality operators
		"WHERE.*<>",        // Not equals
		"WHERE.*OR",        // OR conditions
		"WHERE.*ILIKE",     // Case insensitive LIKE
		"ORDER BY.*RANDOM", // Random ordering
	}

	for _, pattern := range seqScanPatterns {
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return true
		}
	}

	return false
}

func (qp *QueryParser) HasCorrelatedSubquery(query string) bool {
	// Simple detection of correlated subqueries
	query = strings.ToUpper(query)

	// Look for subqueries that reference outer query
	patterns := []string{
		`SELECT.*\(.*SELECT.*WHERE.*=.*\.`,     // Subquery with table reference
		`EXISTS.*\(.*SELECT.*WHERE.*=.*\.`,     // EXISTS with correlation
		`NOT EXISTS.*\(.*SELECT.*WHERE.*=.*\.`, // NOT EXISTS with correlation
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return true
		}
	}

	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
