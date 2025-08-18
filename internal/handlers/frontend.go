// internal/handlers/frontend.go
package handlers

import (
	"io/fs"
	"net/http"
	"os"
	"path"
)

// FrontendHandler ahora es una función que RECIBE el sistema de archivos embebido (fs.FS).
// Esta es la firma correcta que coincide con la llamada que se hace desde main.go.
// Se elimina la directiva 'embed' de este archivo porque la hemos centralizado en main.go,
// que es una práctica más robusta y soluciona los problemas de rutas relativas.
func FrontendHandler(distFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := path.Clean(r.URL.Path)

		// Lógica para SPA (Single-Page Application):
		// Intentamos abrir el archivo solicitado desde el sistema de archivos que nos pasaron.
		_, err := distFS.Open(filePath[1:]) // Se omite el '/' inicial
		if os.IsNotExist(err) {
			// Si el archivo no existe, es una ruta del lado del cliente (ej. /dashboard).
			// Servimos siempre el 'index.html' principal.
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
