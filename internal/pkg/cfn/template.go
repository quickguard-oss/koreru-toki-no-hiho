package cfn

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"text/template"
)

/*
templateData represents the data used to populate CloudFormation templates.
*/
type templateData struct {
	GeneratorName     string // generator name
	GeneratorVersion  string // version of the generator
	DBIdentifier      string // DB cluster/instance identifier
	DBIdentifierShort string // shortened DB identifier for display
	DBType            string // type of the DB (see `internal/pkg/rds`)
	Qualifier         string // unique qualifier specific to the stack
}

/*
GenerateTemplateBody generates a CloudFormation template.
*/
func GenerateTemplateBody(dbIdentifier string, dbIdentifierShort string, dbType string, qualifier string) (string, error) {
	slog.Debug("Generating CloudFormation template",
		"dbIdentifier", dbIdentifier,
		"dbIdentifierShort", dbIdentifierShort,
		"dbType", dbType,
		"qualifier", qualifier,
	)

	t := template.New("cfn")

	t, err := t.Funcs(customFuncMap(t)).Parse(templateStr)

	if err != nil {
		return "", fmt.Errorf("failed to parse Golang template: %w", err)
	}

	data := templateData{
		GeneratorName:     generatorName,
		GeneratorVersion:  generatorVersion,
		DBIdentifier:      dbIdentifier,
		DBIdentifierShort: dbIdentifierShort,
		DBType:            dbType,
		Qualifier:         qualifier,
	}

	var buf bytes.Buffer

	if err = t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute Golang template: %w", err)
	}

	slog.Debug("CloudFormation template generated successfully")

	return buf.String(), nil
}

/*
customFuncMap returns a FuncMap with custom functions registered.
*/
func customFuncMap(t *template.Template) template.FuncMap {
	fm := template.FuncMap{}

	// include: evaluates and returns a string from a defined partial template.
	fm["include"] = func(name string, data any) (string, error) {
		var buf bytes.Buffer

		if err := t.ExecuteTemplate(&buf, name, data); err != nil {
			return "", fmt.Errorf("failed to execute partial template '%s': %w", name, err)
		}

		return buf.String(), nil
	}

	// indent: inserts a specified number of spaces at the beginning of each line.
	fm["indent"] = func(depth int, str string) string {
		indent := strings.Repeat(" ", depth)

		return indent + strings.Replace(str, "\n", "\n"+indent, -1)
	}

	return fm
}
