package cfn

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateTemplateBody(t *testing.T) {
	testCases := []struct {
		name              string
		dbIdentifier      string
		dbIdentifierShort string
		dbType            string
		qualifier         string
		wantErr           bool
		expectFile        string
	}{
		{
			name:              "Aurora",
			dbIdentifier:      "aurora-db-identifier",
			dbIdentifierShort: "aurora-db-i",
			dbType:            "aurora",
			qualifier:         "abcdef",
			wantErr:           false,
			expectFile:        "aurora.yml",
		},
		{
			name:              "RDS",
			dbIdentifier:      "rds-db-identifier",
			dbIdentifierShort: "rds-db-ide",
			dbType:            "rds",
			qualifier:         "ghijklm",
			wantErr:           false,
			expectFile:        "rds.yml",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GenerateTemplateBody(tc.dbIdentifier, tc.dbIdentifierShort, tc.dbType, tc.qualifier)

			if tc.wantErr {
				assert.Error(t, err, "Expected an error to be returned")
			} else {
				assert.NoError(t, err, "Unexpected error occurred")

				expected := readTestFile(t, tc.expectFile)

				assert.Equal(t, expected, got, "Generated template does not match expected output")
			}
		})
	}
}

/*
readTestFile reads a testdata file.
*/
func readTestFile(t *testing.T, filename string) string {
	t.Helper()

	filePath := filepath.Join("testdata", "templates", filename)

	data, err := os.ReadFile(filePath)

	if err != nil {
		t.Fatalf("failed to read test file '%s': %v", filePath, err)
	}

	return strings.TrimSpace(string(data))
}
