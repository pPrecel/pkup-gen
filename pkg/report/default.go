package report

import (
	"bytes"
	"text/template"

	"github.com/pPrecel/PKUP/internal/file"
)

const (
	defaultTemplate = `period:
{{.PeriodFrom}} - {{.PeriodTill}}

approvalDate:
{{.ApprovalDate}}

result:
{{range .Result}}
- {{ . -}}
{{end}}
`
)

type defaultRenderer struct {
	template string
}

func newDefault() *defaultRenderer {
	return &defaultRenderer{
		template: defaultTemplate,
	}
}

func (dr *defaultRenderer) RenderToFile(dir, filename string, values Values) error {
	tmpl, err := template.New(filename).Parse(dr.template)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(nil)
	err = tmpl.Execute(buf, values)
	if err != nil {
		return err
	}

	return file.Create(dir, filename, buf.String())
}
