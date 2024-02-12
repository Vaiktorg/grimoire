package main

import (
	"github.com/vaiktorg/grimoire/serve"
	"github.com/vaiktorg/grimoire/site"
	"net/http"
)

func main() {
	s := site.NewSite(nil)

	srv := serve.NewServer(&serve.Config{
		AppName: "Site Tmpl Wrapper",
	})

	srv.MUX(func(mux *http.ServeMux) {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			s.RenderHTML(w, "home.index", nil)
		})
	})

	srv.ListenAndServe()
}
