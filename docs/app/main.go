package main

import (
	"fmt"
	"github.com/vaiktorg/grimoire/docs"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	wiki := docs.NewWiki(docs.Config{
		Project:  "GBNCX",
		Desc:     "Smart Model Rocket.",
		DocsPath: "docs/wiki",
	})

	mux.Handle("/", wiki)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		fmt.Println(err)
		return
	}
}
