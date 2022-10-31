package documentstore

import (
	"fmt"
	"io"
	"strings"
)

func (d *Dir) Print(lvl int, w io.Writer) {

	indent := 4
	idx := 0

	_, _ = fmt.Fprintf(w, "%s"+"%s\n", strings.Repeat(" ", lvl), d.Meta.Name+"/")

	for _, f := range d.Files {
		_, _ = fmt.Fprintf(w, "%s"+"%s\n", strings.Repeat(" ", lvl+indent), "- "+f.Meta.Name)
		idx++
	}

	for _, d := range d.Dirs {
		d.Print(lvl+indent, w)
	}
}
