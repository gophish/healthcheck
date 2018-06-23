package template

import (
	"bytes"
	"text/template"
)

// TextTemplate is the filepath to the template used when creating the
// text/plain MIME part of the generated email.
const TextTemplate = "./template/templates/email_template.txt"

// HTMLTemplate is the filepath to the template used when creating the
// text/html MIME part of the generated email.
const HTMLTemplate = "./template/templates/email_template.html"

// ExecuteTemplate creates a templated string based on the provided
// template filename and data.
func ExecuteTemplate(filename string, data interface{}) (string, error) {
	buff := bytes.Buffer{}
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		return buff.String(), err
	}
	err = tmpl.Execute(&buff, data)
	return buff.String(), err
}
