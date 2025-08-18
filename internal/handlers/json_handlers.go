// internal/handlers/json_helper.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// respondWithJSON es una función auxiliar para escribir respuestas JSON estandarizadas.
// Centraliza el proceso de serializar datos a JSON y establecer las cabeceras HTTP correctas.
//
// Parámetros:
//
//	w: El http.ResponseWriter para escribir la respuesta.
//	code: El código de estado HTTP (ej. 200 para OK, 400 para Bad Request).
//	payload: Los datos a ser codificados en JSON. Puede ser cualquier struct o mapa.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	// json.Marshal convierte la estructura de datos de Go (payload) en un slice de bytes en formato JSON.
	response, err := json.Marshal(payload)
	if err != nil {
		// Si ocurre un error durante la serialización, es un problema del servidor.
		log.Printf("Error al serializar JSON: %v", err)
		// Se responde con un error 500 Internal Server Error.
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error Interno del Servidor"))
		return
	}

	// Se establece la cabecera 'Content-Type' para que el cliente (navegador/Vue)
	// sepa que la respuesta es JSON y cómo interpretarla.
	w.Header().Set("Content-Type", "application/json")
	// Se escribe el código de estado HTTP en la cabecera de la respuesta.
	w.WriteHeader(code)
	// Se escribe el cuerpo de la respuesta JSON.
	w.Write(response)
}

// respondWithError es una función auxiliar para enviar respuestas de error JSON consistentes.
func respondWithError(w http.ResponseWriter, code int, message string) {
	// Se utiliza un mapa para crear una estructura JSON simple: {"error": "mensaje"}
	respondWithJSON(w, code, map[string]string{"error": message})
}
