package handlers

import (
	"html/template"
)

// GlobalTemplates es una variable global para almacenar las plantillas.
// Esto simplifica el acceso desde todos los handlers del paquete.
var templates *template.Template

// InitHandlers inicializa el paquete de handlers con las plantillas necesarias.
func InitHandlers(tpl *template.Template) {
	templates = tpl
}
