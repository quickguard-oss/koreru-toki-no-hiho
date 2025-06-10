package logger

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

/*
FormatAsTable converts data into a formatted table string representation.
It takes a slice of headers and a 2D slice of body content, then formats them
into a well-aligned table structure with proper spacing.
*/
func FormatAsTable(headers []string, body [][]string) string {
	slog.Debug("Converting data to table format for display")

	headerMaxWidth := make([]int, len(headers))

	for i, header := range headers {
		headerMaxWidth[i] = len(header)
	}

	for _, cols := range body {
		for i, col := range cols {
			if headerMaxWidth[i] < len(col) {
				headerMaxWidth[i] = len(col)
			}
		}
	}

	formats := make([]string, len(headers))

	// NOTE: Create format string like "%-10s" for left-aligned display
	for i, width := range headerMaxWidth {
		formats[i] = fmt.Sprintf("%%-%ds", width)
	}

	cells := make([][]string, len(body)+1)

	cells[0] = make([]string, len(headers))

	for i, col := range headers {
		cells[0][i] = fmt.Sprintf(formats[i], strings.ToUpper(col))
	}

	for i, cols := range body {
		cells[i+1] = make([]string, len(headers))

		for j, col := range cols {
			cells[i+1][j] = fmt.Sprintf(formats[j], col)
		}
	}

	rows := make([]string, len(body)+1)

	for i, cols := range cells {
		rows[i] = strings.Join(cols, "   ")
	}

	slog.Debug("Data successfully formatted as table")

	return strings.Join(rows, "\n")
}

/*
FormatAsJSON converts data into a JSON string representation.
It takes a slice of attribute names and a 2D slice of body content,
then formats them into a JSON array of objects where each object's keys
are the attribute names and values are the corresponding body content.
*/
func FormatAsJSON(attrs []string, body [][]string) (string, error) {
	slog.Debug("Converting data to JSON format")

	result := make([]map[string]string, len(body))

	for i, values := range body {
		obj := make(map[string]string, len(attrs))

		for j, v := range values {
			obj[attrs[j]] = v
		}

		result[i] = obj
	}

	jsonBytes, err := json.Marshal(result)

	if err != nil {
		return "", fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	slog.Debug("Data successfully formatted as JSON")

	return string(jsonBytes), nil
}
