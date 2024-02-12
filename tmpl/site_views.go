package site

import (
	"html/template"
)

type HTMLData[T any] struct {
	Head HeadData
	Menu MenuData
	Nav  NavData
	Body BodyData[T]
}

type HeadData struct {
	CSS template.CSS
	JS  template.JS

	Title string
	Icon  string
}

type MenuData struct {
	MenuIcon string
	MenuName string

	MenuItems map[string]string //K: Label; V: URL
}

type BodyData[T any] struct {
	Main   MainData[T]
	Footer FooterData
}

type NavData map[string]NavItems //K: Group Name; V: NavItems
type NavItems map[string]string  //K: Label; V: URL

type MainData[T any] struct {
	Content T
}

type FooterData template.HTML
