package configo

import (
	"fmt"
	"strings"

	"github.com/vsysa/configo/internal/parser/env"
	"github.com/vsysa/configo/internal/parser/yaml"

	"unicode/utf8"
)

func GenerateYAMLTemplate(cfg interface{}, printDescription bool) string {
	return yaml.GenerateYAMLTemplate(cfg, printDescription)
}

// EnvHelpFormat defines the type of output format for environment variable docs.
type EnvHelpFormat int

const (
	// AsciiTable displays a nicely aligned ASCII table for terminal output.
	AsciiTable EnvHelpFormat = iota

	// Inline displays each environment variable on a single line, e.g.,
	// `ENV_VAR [default=...] # help text`
	Inline

	// MarkdownTable displays the environment variables in a Markdown table,
	// e.g., with `| EnvVar | Default | Description |`.
	MarkdownTable
)

// parseEnvStructure recursively scans the given type (and nested structs, if any),
// collecting environment variable information. If the current field is itself
// a struct and has an `env` tag, that tag becomes the prefix for the nested fields.
//
// For example, if a parent struct has `env:"db"` and the nested struct has a field
// with `env:"host"`, it will generate `DB_HOST`.
func GenerateEnvHelp(cfg interface{}, format EnvHelpFormat) string {
	lines := env.GetEnvs(cfg)

	// Choose the output format based on the 'format' parameter
	switch format {
	case Inline:
		return formatEnvHelpInline(lines)
	case MarkdownTable:
		return formatEnvHelpMarkdownTable(lines)
	case AsciiTable:
		fallthrough
	default:
		return formatEnvHelpAsciiTable(lines)
	}
}

// formatEnvHelpInline displays each environment variable on a single line.
// Example:
//
//	ENV_VAR [default=...] # help text
func formatEnvHelpInline(lines []env.EnvInfo) string {
	var sb strings.Builder
	for _, info := range lines {
		line := info.EnvVar

		if info.DefaultValue != "" {
			line += fmt.Sprintf(" [default=%s]", info.DefaultValue)
		}
		if info.HelpText != "" {
			line += fmt.Sprintf(" # %s", info.HelpText)
		}

		sb.WriteString(line + "\n")
	}
	return sb.String()
}

// formatEnvHelpMarkdownTable displays the environment variables in a Markdown table,
// e.g.:
//
// | Environment Variable | Default  | Description |
// |----------------------|----------|-------------|
// | HOST                | localhost| The hostname|
// | PORT                | 8080     | The port    |
func formatEnvHelpMarkdownTable(lines []env.EnvInfo) string {
	var sb strings.Builder
	sb.WriteString("| Environment Variable | Default | Description |\n")
	sb.WriteString("|----------------------|---------|-------------|\n")

	for _, info := range lines {
		defaultVal := info.DefaultValue
		if defaultVal == "" {
			defaultVal = "N/A"
		}
		help := info.HelpText
		if help == "" {
			help = "N/A"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", info.EnvVar, defaultVal, help))
	}
	return sb.String()
}

// formatEnvHelpAsciiTable dynamically measures column widths and prints
// an ASCII table that should look aligned in most terminals.
//
// Example output:
//
//	+----------------------+----------+----------------------+
//	| Environment Variable | Default  | Description          |
//	+----------------------+----------+----------------------+
//	| HOST                | localhost| The hostname         |
//	| PORT                | 8080     | The port number      |
//	| ENABLED             | true     | Enable the feature   |
//	| VERSION             | 1.0      | App version          |
//	+----------------------+----------+----------------------+
func formatEnvHelpAsciiTable(envLines []env.EnvInfo) string {
	// Define headers
	headerEnv := "Environment Variable"
	headerDef := "Default"
	headerHelp := "Description"

	// Find maximum width for each column by comparing the header and values
	envVarColWidth := len(headerEnv)
	defColWidth := len(headerDef)
	helpColWidth := len(headerHelp)

	for _, info := range envLines {
		if utf8.RuneCountInString(info.EnvVar) > envVarColWidth {
			envVarColWidth = utf8.RuneCountInString(info.EnvVar)
		}
		if utf8.RuneCountInString(info.DefaultValue) > defColWidth {
			defColWidth = utf8.RuneCountInString(info.DefaultValue)
		}
		if utf8.RuneCountInString(info.HelpText) > helpColWidth {
			helpColWidth = utf8.RuneCountInString(info.HelpText)
		}
	}

	// Helper function to create a horizontal rule
	makeLine := func() string {
		return "+" +
			strings.Repeat("-", envVarColWidth+2) + "+" +
			strings.Repeat("-", defColWidth+2) + "+" +
			strings.Repeat("-", helpColWidth+2) + "+"
	}

	// Helper function to format a single row of text with padding
	makeRow := func(col1, col2, col3 string) string {
		return fmt.Sprintf("| %-*s | %-*s | %-*s |",
			envVarColWidth, col1,
			defColWidth, col2,
			helpColWidth, col3,
		)
	}

	var sb strings.Builder

	// Top line
	sb.WriteString(makeLine() + "\n")

	// Header row
	sb.WriteString(makeRow(headerEnv, headerDef, headerHelp) + "\n")

	// Separator
	sb.WriteString(makeLine() + "\n")

	// Data rows
	for _, info := range envLines {
		sb.WriteString(makeRow(info.EnvVar, info.DefaultValue, info.HelpText) + "\n")
	}

	// Bottom line
	sb.WriteString(makeLine() + "\n")

	return sb.String()
}
