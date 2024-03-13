package config

import (
	"bytes"
	"embed"
	"path/filepath"
	"text/template"
)

const (
	templateFolder = "templates"
)

//go:embed all:templates
var templates embed.FS

// Render renders a templates in the package `templates` folder. The templates
// files are embedded in build-time.
func Render(template string, context interface{}) (string, error) {
	tplBytes, err := templates.ReadFile(filepath.Join(templateFolder, template))
	if err != nil {
		return "", err
	}
	return RenderTemplate(string(tplBytes), context)
}

func RenderTemplate(tpl string, context interface{}) (string, error) {
	result := bytes.NewBufferString("")
	tmpl, err := template.New("").Parse(tpl)
	if err != nil {
		return "", err
	}

	err = tmpl.Execute(result, context)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}
