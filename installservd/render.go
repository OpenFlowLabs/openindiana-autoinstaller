package installservd

import (
	"io"
	"text/template"

	"github.com/labstack/echo"
)

type Template struct {
	templates map[string]*template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates[name].Execute(w, data)
}

func (t *Template) Load(i *Installservd) {
	t.templates = make(map[string]*template.Template)
	for _, p := range Profiles {
		for _, tem := range p.Templates {
			name := i.getAssetPath(*tem)
			t.templates[name] = template.Must(template.ParseFiles(name))
		}
	}
}
