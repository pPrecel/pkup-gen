package raport

import (
	"fmt"
	"path"

	"github.com/nguyenthenguyen/docx"
)

const (
	DocxPeriodFromTmpl   = "pkupGenPeriodFrom"
	DocxPeriodTillTmpl   = "pkupGenPeriodTill"
	DocxApprovalDateTmpl = "pkupGenApprovalDate"
	DocxResultsTmpl      = "pkupGenResults"
)

type Values struct {
	PeriodFrom   string
	PeriodTill   string
	ApprovalDate string
	Result       []string
}

type templateRenderer struct {
	tmplPath string
}

func newFromTemplate(path string) *templateRenderer {
	return &templateRenderer{
		tmplPath: path,
	}
}

func (tr *templateRenderer) RenderToFile(dir, filename string, values Values) error {
	r, err := docx.ReadDocxFile(tr.tmplPath)
	if err != nil {
		return fmt.Errorf("failed to load docx template: %s", err.Error())
	}

	resultString := ""
	for i := range values.Result {
		resultString += fmt.Sprintf("- %s\n", values.Result[i])
	}

	docx1 := r.Editable()
	docx1.Replace(DocxPeriodFromTmpl, values.PeriodFrom, -1)
	docx1.Replace(DocxPeriodTillTmpl, values.PeriodTill, -1)
	docx1.Replace(DocxApprovalDateTmpl, values.ApprovalDate, -1)
	docx1.Replace(DocxResultsTmpl, resultString, -1)

	return docx1.WriteToFile(path.Join(dir, filename))
}
