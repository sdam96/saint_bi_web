// La declaración 'package services' indica que este archivo forma parte del paquete 'services'.
// En Go, un paquete es una colección de archivos fuente en el mismo directorio que se compilan juntos.
// Este paquete contiene la lógica de negocio principal de la aplicación.
package services

import (
	// "fmt" es un paquete estándar que implementa funciones para formatear I/O (entrada/salida),
	// como la creación de strings de error con formato.
	"fmt"
	// "strings" es un paquete estándar que proporciona funciones para la manipulación de strings.
	// Aquí lo usamos para la función 'ToLower', que convierte un string a minúsculas.
	"strings"
	// "sync" proporciona primitivas de sincronización, como los WaitGroups.
	// 'sync.WaitGroup' nos permite esperar a que un conjunto de goroutines (hilos ligeros) terminen su ejecución.
	"sync"

	// "time" es un paquete estándar para medir y representar el tiempo. Lo usamos para parsear
	// las fechas de los filtros y realizar comparaciones.
	"time"

	// Importamos nuestros paquetes internos. Esto permite que este archivo utilice
	// las structs y funciones definidas en otros lugares de nuestro proyecto.
	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/models"
)

// GetTransactions obtiene y filtra una lista de documentos según el tipo y rango de fechas.
// Esta función implementa la lógica de negocio principal para el primer nivel del drilldown.
//
// Parámetros:
//   - client (*api.SaintClient): Es un puntero al cliente de la API, ya autenticado por el middleware.
//   - docType (string): Identifica el tipo de documento a buscar (ej: "invoices-credit").
//   - start, end (time.Time): Definen el rango de fechas para el cual se deben filtrar las transacciones.
//
// Retorna:
//   - interface{}: Un valor de cualquier tipo. En este caso, devolverá un slice de transacciones
//     (ej: []models.Invoice). Usar 'interface{}' nos da la flexibilidad de devolver diferentes tipos
//     de slices desde la misma función.
//   - error: Un objeto de error si algo falla, o 'nil' si la operación es exitosa.
func GetTransactions(client *api.SaintClient, docType string, start, end time.Time) (interface{}, error) {
	// La sentencia 'switch' en Go es una forma eficiente de escribir una secuencia de if-else-if.
	// Compara el valor de 'strings.ToLower(docType)' con los diferentes casos definidos.
	switch strings.ToLower(docType) {

	// CASOS PARA VENTAS (FACTURAS): Estos dos casos manejan la lógica de negocio del archivo Dart.
	case "invoices-credit", "invoices-cash":
		var allInvoices []models.Invoice
		var allReceivables []models.AccReceivable
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			allInvoices, _ = client.GetInvoices()
		}()
		go func() {
			defer wg.Done()
			allReceivables, _ = client.GetAccReceivables()
		}()
		wg.Wait()

		outstandingReceivables := make(map[string]struct{})
		for _, ar := range allReceivables {
			if ar.Saldo != nil && *ar.Saldo > 0 && ar.NumeroD != nil {
				outstandingReceivables[*ar.NumeroD] = struct{}{}
			}
		}

		// --- CORRECCIÓN CLAVE ---
		// 'make([]models.Invoice, 0)' inicializa un slice de facturas con longitud y capacidad 0.
		// A diferencia de 'var filteredInvoices []models.Invoice', esto crea un slice "no nulo" pero vacío.
		// Al codificarlo a JSON, se convertirá en '[]' (un array vacío), que es lo que el frontend espera,
		// en lugar de 'null', que causaba el error.
		filteredInvoices := make([]models.Invoice, 0)
		// --- FIN DE LA CORRECCIÓN ---

		for _, inv := range allInvoices {
			if inv.FechaE == nil {
				continue
			}
			date, err := time.Parse("2006-01-02 15:04:05", *inv.FechaE)
			if err != nil || date.Before(start) || date.After(end) {
				continue
			}

			_, hasOutstandingBalance := outstandingReceivables[*inv.NumeroD]

			if docType == "invoices-credit" && hasOutstandingBalance {
				filteredInvoices = append(filteredInvoices, inv)
			} else if docType == "invoices-cash" && !hasOutstandingBalance {
				filteredInvoices = append(filteredInvoices, inv)
			}
		}
		return filteredInvoices, nil

	// CASOS PARA CXC Y CXP: Estos son más sencillos, solo filtran por fecha.
	case "receivables":
		allReceivables, err := client.GetAccReceivables()
		if err != nil {
			return nil, err
		}
		// Se aplica la misma corrección: inicializar con make() para evitar un slice nulo.
		filtered := make([]models.AccReceivable, 0)
		for _, acc := range allReceivables {
			if acc.FechaE != nil {
				if date, err := time.Parse("2006-01-02 15:04:05", *acc.FechaE); err == nil && !date.Before(start) && !date.After(end) {
					filtered = append(filtered, acc)
				}
			}
		}
		return filtered, nil

	case "payables":
		allPayables, err := client.GetAccPayables()
		if err != nil {
			return nil, err
		}
		// Se aplica la misma corrección: inicializar con make() para evitar un slice nulo.
		filtered := make([]models.AccPayable, 0)
		for _, acc := range allPayables {
			if acc.FechaE != nil {
				if date, err := time.Parse("2006-01-02 15:04:05", *acc.FechaE); err == nil && !date.Before(start) && !date.After(end) {
					filtered = append(filtered, acc)
				}
			}
		}
		return filtered, nil
	}
	// Si el 'docType' no coincide con ninguno de los casos anteriores, devolvemos un error claro.
	return nil, fmt.Errorf("tipo de documento '%s' no es válido o no está implementado", docType)
}

// FullTransactionDetail es una 'struct' que define una estructura de datos personalizada.
// Su propósito es agrupar toda la información relevante de una transacción en un solo objeto,
// para enviarla al frontend en una respuesta JSON limpia y predecible.
type FullTransactionDetail struct {
	// 'Document' e 'Items' son de tipo 'interface{}'. Esto nos permite asignarles
	// diferentes tipos de structs (ej. models.Invoice, models.Purchase) sin cambiar
	// la definición de FullTransactionDetail. Es una forma de polimorfismo en Go.
	Document interface{} `json:"document"`
	Items    interface{} `json:"items"`

	// 'Customer', 'Seller' y 'Supplier' son punteros a structs del paquete models.
	// Usamos punteros (*) para que puedan tener un valor 'nil' si no se encuentran
	// o no aplican (ej. un proveedor en una factura de venta).
	// La etiqueta `json:"customer,omitempty"` le dice al codificador JSON de Go
	// que si el campo 'Customer' es nulo (nil), debe omitirlo completamente de la
	// respuesta JSON, en lugar de incluir "customer": null.
	Customer *models.Customer `json:"customer,omitempty"`
	Seller   *models.Seller   `json:"seller,omitempty"`
	Supplier *models.Supplier `json:"supplier,omitempty"`
}

// GetTransactionDetail obtiene todos los datos relacionados con un único documento.
// Esta función es un ejemplo de cómo agregar datos de múltiples fuentes (facturas, items, clientes, etc.)
// en una sola respuesta cohesiva para el frontend.
func GetTransactionDetail(client *api.SaintClient, docType, docID string) (*FullTransactionDetail, error) {
	// Usamos un 'switch' para manejar diferentes tipos de documentos en el futuro.
	// Por ahora, solo hemos implementado la lógica para "invoice".
	switch strings.ToLower(docType) {
	case "invoice":
		// --- 1. Obtención de Datos Maestros en Paralelo ---
		// Para encontrar los detalles de una factura (items, cliente, vendedor), necesitamos
		// las listas completas de esos datos. Para acelerar el proceso, las solicitamos
		// todas al mismo tiempo usando goroutines.
		var wg sync.WaitGroup
		var allInvoiceItems []models.InvoiceItem
		var allCustomers []models.Customer
		var allSellers []models.Seller

		wg.Add(3) // Indicamos que vamos a esperar 3 operaciones.
		go func() { defer wg.Done(); allInvoiceItems, _ = client.GetInvoiceItems() }()
		go func() { defer wg.Done(); allCustomers, _ = client.GetCustomers() }()
		go func() { defer wg.Done(); allSellers, _ = client.GetSellers() }()
		wg.Wait() // Esperamos a que las 3 goroutines terminen.

		// --- 2. Búsqueda del Documento Principal ---
		// La API de Saint no permite filtrar facturas por NumeroD, así que traemos todas.
		allInvoices, err := client.GetInvoices()
		if err != nil {
			return nil, err
		}

		var document models.Invoice
		found := false // Una bandera para saber si encontramos la factura.
		for _, inv := range allInvoices {
			// Los campos en los modelos son punteros, por lo que debemos comprobar que no sean 'nil'
			// antes de desreferenciarlos con el operador '*'.
			if inv.NumeroD != nil && *inv.NumeroD == docID {
				document = inv
				found = true
				break // 'break' sale del bucle 'for' tan pronto como encontramos lo que buscamos.
			}
		}
		if !found {
			return nil, fmt.Errorf("factura con NumeroD '%s' no encontrada", docID)
		}

		// --- 3. Búsqueda de Datos Relacionados ---
		// Ahora que tenemos la factura, usamos sus IDs (NumeroD, CodClie, CodVend)
		// para encontrar los registros correspondientes en las listas que ya obtuvimos.
		var items []models.InvoiceItem
		for _, item := range allInvoiceItems {
			if item.NumeroD != nil && *item.NumeroD == docID {
				items = append(items, item)
			}
		}

		var customer *models.Customer
		if document.CodClie != nil {
			for _, c := range allCustomers {
				if c.CodClie != nil && *c.CodClie == *document.CodClie {
					temp := c        // Creamos una copia de 'c'
					customer = &temp // Asignamos la dirección de memoria de la copia
					break
				}
			}
		}

		var seller *models.Seller
		if document.CodVend != nil {
			for _, s := range allSellers {
				if s.CodVend != nil && *s.CodVend == *document.CodVend {
					temp := s
					seller = &temp
					break
				}
			}
		}

		// --- 4. Ensamblaje de la Respuesta Final ---
		// Se crea una instancia de nuestra struct 'FullTransactionDetail' y se puebla con todos
		// los datos que hemos encontrado.
		// La sintaxis '&FullTransactionDetail{...}' crea la struct y devuelve un puntero a ella.
		return &FullTransactionDetail{Document: document, Items: items, Customer: customer, Seller: seller}, nil
	}

	return nil, fmt.Errorf("obtener detalle para el tipo '%s' aún no está implementado", docType)
}
