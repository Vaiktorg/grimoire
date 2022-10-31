package static

import (
	"embed"
	_ "embed"
)

//go:embed *.gohtml
var TemplateFS embed.FS
