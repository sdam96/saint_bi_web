// internal/handlers/drilldown.go
package handlers

import (
	// "log" para registrar errores en el servidor.
	"log"
	// "net/http" para manejar las solicitudes y respuestas HTTP.
	"net/http"
	// "time" para parsear y manejar las fechas del filtro.
	"time"

	// Importamos nuestro paquete de servicios, donde residirá la lógica de negocio.
	"saintnet.com/m/internal/services"
)

// GetTransactionsList es el handler para la ruta GET /api/transactions.
// Su propósito es devolver una lista de documentos (facturas, cxc, cxp)
// filtrados por un tipo específico y un rango de fechas.
func GetTransactionsList(w http.ResponseWriter, r *http.Request) {
	// 1. Obtenemos el cliente API que el middleware ya autenticó y nos dejó en el contexto.
	client := getClientFromContext(r)
	if client == nil {
		respondWithError(w, http.StatusUnauthorized, "Cliente API no disponible. Por favor, seleccione una conexión.")
		return
	}

	// 2. Extraemos los parámetros de la URL. r.URL.Query().Get("clave") lee valores como "?clave=valor".
	docType := r.URL.Query().Get("type") // ej: "invoices", "receivables", "payables"
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	// 3. Validamos que los parámetros necesarios estén presentes.
	if docType == "" || startDateStr == "" || endDateStr == "" {
		respondWithError(w, http.StatusBadRequest, "Los parámetros 'type', 'startDate' y 'endDate' son requeridos")
		return
	}

	// 4. Parseamos las fechas recibidas como string al formato time.Time de Go.
	// El layout "2006-01-02" es el formato estándar que Go usa para interpretar fechas "AAAA-MM-DD".
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Formato de startDate inválido, use AAAA-MM-DD")
		return
	}
	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Formato de endDate inválido, use AAAA-MM-DD")
		return
	}
	// Ajustamos la hora de la fecha final a las 23:59:59 para asegurar que se incluyan
	// todas las transacciones de ese día en el filtro.
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// 5. Delegamos la lógica de negocio al paquete de servicios.
	// El handler no sabe CÓMO se obtienen los datos, solo pide lo que necesita.
	transactions, err := services.GetTransactions(client, docType, startDate, endDate)
	if err != nil {
		log.Printf("Error obteniendo transacciones desde el servicio: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al obtener la lista de transacciones")
		return
	}

	// 6. Si todo sale bien, enviamos los datos como una respuesta JSON.
	respondWithJSON(w, http.StatusOK, transactions)
}

// GetTransactionDetail es el handler para rutas como GET /api/transaction/invoice/123.
// Devuelve el detalle completo y agregado de una única transacción.
func GetTransactionDetail(w http.ResponseWriter, r *http.Request) {
	client := getClientFromContext(r)
	if client == nil {
		respondWithError(w, http.StatusUnauthorized, "Cliente API no disponible")
		return
	}

	// 'r.PathValue' es una función de Go 1.22+ que extrae segmentos dinámicos de la ruta.
	// Por ejemplo, en "/api/transaction/invoice/123", docType será "invoice" y docID será "123".
	docType := r.PathValue("type")
	docID := r.PathValue("id")

	// De nuevo, delegamos la lógica compleja de buscar y agregar los datos al paquete de servicios.
	detail, err := services.GetTransactionDetail(client, docType, docID)
	if err != nil {
		log.Printf("Error obteniendo detalle de la transacción desde el servicio: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error al obtener el detalle de la transacción: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, detail)
}
