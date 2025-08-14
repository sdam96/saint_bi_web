package main

import (
	"embed"
	"html/template"
)

//go:embed templates
var templateFS embed.FS

// ParseTemplates carga y analiza los archivos HTML desde el sistema de archivos embebido.
func ParseTemplates() *template.Template {
	tpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		// Panic es apropiado aquí porque si las plantillas no cargan, la aplicación no puede funcionar.
		panic(err)
	}
	return tpl
}
