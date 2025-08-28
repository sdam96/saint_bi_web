// La declaración 'package services' indica que este archivo forma parte del paquete 'services'.
// En Go, un paquete es una colección de archivos fuente en el mismo directorio que se compilan juntos.
// Este paquete contiene la lógica de negocio principal de la aplicación.
package services

import (
	// "fmt" es un paquete estándar que implementa funciones para formatear I/O (entrada/salida),
	// como la creación de strings de error con formato.
	"fmt"
	// "reflect" es un paquete de Go que permite la introspección y manipulación de tipos en tiempo de ejecución.
	// Lo usamos aquí en la función de consolidación para unir slices de diferentes tipos de manera genérica.
	"reflect"
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

// GetTransactions obtiene y filtra una lista de documentos para UNA SOLA conexión.
// Esta función implementa la lógica de negocio para el primer nivel del "drilldown" o desglose de datos.
//
// PARÁMETROS:
//   - client (*api.SaintClient): Es un puntero al cliente de la API, ya autenticado por el middleware.
//   - docType (string): Identifica el tipo de documento a buscar (ej: "invoices-credit", "receivables").
//   - start, end (time.Time): Definen el rango de fechas para el cual se deben filtrar las transacciones.
//
// RETORNA:
//   - interface{}: Un valor de cualquier tipo. En este caso, devolverá un slice de transacciones
//     (ej: []models.Invoice). Usar 'interface{}' nos da la flexibilidad de devolver diferentes tipos
//     de slices desde la misma función. Es una forma de polimorfismo en Go.
//   - error: Un objeto de error si algo falla, o 'nil' si la operación es exitosa.
func GetTransactions(client *api.SaintClient, docType string, start, end time.Time) (interface{}, error) {
	// La sentencia 'switch' en Go es una forma eficiente de escribir una secuencia de if-else-if.
	// Compara el valor de 'strings.ToLower(docType)' con los diferentes casos definidos.
	switch strings.ToLower(docType) {

	// CASOS PARA VENTAS (FACTURAS): Estos dos casos manejan la lógica más compleja.
	case "invoices-credit", "invoices-cash":
		// LÓGICA: Para determinar si una factura es a crédito o de contado, necesitamos dos fuentes de datos:
		// 1. La lista de todas las facturas (`allInvoices`).
		// 2. La lista de todas las cuentas por cobrar (`allReceivables`).
		// Una factura se considera a crédito si tiene un saldo pendiente en las cuentas por cobrar.
		// Para optimizar, obtenemos ambas listas en paralelo usando goroutines.
		var allInvoices []models.Invoice
		var allReceivables []models.AccReceivable
		var wg sync.WaitGroup
		wg.Add(2) // Le decimos al WaitGroup que vamos a esperar dos goroutines.
		go func() {
			defer wg.Done()
			allInvoices, _ = client.GetInvoices() // El error se ignora aquí por simplicidad, pero en producción se podría manejar.
		}()
		go func() {
			defer wg.Done()
			allReceivables, _ = client.GetAccReceivables()
		}()
		wg.Wait() // Esperamos a que ambas descargas terminen.

		// LÓGICA: Creamos un mapa para una búsqueda ultra-rápida. Un mapa nos permite verificar si una factura
		// tiene saldo pendiente en tiempo constante O(1), en lugar de tener que recorrer todo el slice de
		// cuentas por cobrar cada vez, lo cual sería muy ineficiente O(n).
		// SINTAXIS de Go: `make(map[string]struct{})` crea un mapa con claves de tipo string.
		// El valor `struct{}` es un tipo de struct vacía que no ocupa memoria. Se usa como un truco
		// para crear un "conjunto" (set), ya que solo nos interesa la existencia de la clave.
		outstandingReceivables := make(map[string]struct{})
		for _, ar := range allReceivables {
			if ar.Saldo != nil && *ar.Saldo > 0 && ar.NumeroD != nil {
				outstandingReceivables[*ar.NumeroD] = struct{}{}
			}
		}

		// 'make([]models.Invoice, 0)' inicializa un slice de facturas con longitud y capacidad 0.
		// Esto crea un slice "no nulo" pero vacío. Al codificarlo a JSON, se convertirá en '[]' (un array vacío),
		// que es lo que el frontend espera, en lugar de 'null', que podría causar errores.
		filteredInvoices := make([]models.Invoice, 0)

		// Se itera sobre cada factura para filtrarla según el rango de fechas y el tipo solicitado.
		for _, inv := range allInvoices {
			if inv.FechaE == nil {
				continue // `continue` salta a la siguiente iteración del bucle.
			}
			date, err := time.Parse("2006-01-02 15:04:05", *inv.FechaE)
			if err != nil || date.Before(start) || date.After(end) {
				continue
			}

			// Se comprueba si la factura existe en nuestro mapa de saldos pendientes.
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

// GetConsolidatedTransactions obtiene transacciones de TODAS las conexiones en paralelo y las une.
func GetConsolidatedTransactions(connections []models.Connection, docType string, start, end time.Time) (interface{}, error) {
	resultsChan := make(chan interface{}, len(connections))
	errChan := make(chan error, len(connections))
	var wg sync.WaitGroup

	for _, conn := range connections {
		wg.Add(1)
		go func(c models.Connection) {
			defer wg.Done()
			client := api.NewSaintClient(c.ApiURL)
			if err := client.Login(c.ApiUser, c.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093"); err != nil {
				errChan <- err
				return
			}
			// Cada goroutine llama a la misma función `GetTransactions` que usaría una conexión individual.
			result, err := GetTransactions(client, docType, start, end)
			if err != nil {
				errChan <- err
				return
			}
			resultsChan <- result
		}(conn)
	}
	wg.Wait()
	close(resultsChan)
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	// LÓGICA: Se unen los resultados de todas las conexiones.
	var combinedResults interface{}
	isFirst := true
	for res := range resultsChan {
		if isFirst {
			combinedResults = res
			isFirst = false
			continue
		}
		// SINTAXIS de Go: Se usa el paquete `reflect` para trabajar con los slices de tipo `interface{}`.
		// `reflect.ValueOf` obtiene una representación de reflexión del valor.
		// `reflect.AppendSlice` une dos slices de forma segura a nivel de reflexión.
		val1 := reflect.ValueOf(combinedResults)
		val2 := reflect.ValueOf(res)
		if val1.Kind() != reflect.Slice || val2.Kind() != reflect.Slice {
			continue // Omitimos si alguno de los resultados no es un slice.
		}
		combined := reflect.AppendSlice(val1, val2)
		combinedResults = combined.Interface() // Convertimos el resultado de reflexión de vuelta a `interface{}`.
	}

	// LÓGICA: Si ninguna conexión devolvió datos, `combinedResults` sería `nil`.
	// Para asegurar que el frontend siempre reciba un array `[]`, creamos un slice vacío
	// del tipo correcto si es necesario.
	if combinedResults == nil {
		switch strings.ToLower(docType) {
		case "invoices-credit", "invoices-cash":
			return make([]models.Invoice, 0), nil
		case "receivables":
			return make([]models.AccReceivable, 0), nil
		case "payables":
			return make([]models.AccPayable, 0), nil
		}
	}

	return combinedResults, nil
}

// FullTransactionDetail es una 'struct' que define una estructura de datos personalizada.
// Su propósito es agrupar toda la información relevante de una transacción en un solo objeto,
// para enviarla al frontend en una respuesta JSON limpia y predecible.
type FullTransactionDetail struct {
	// 'Document' e 'Items' son de tipo 'interface{}'. Esto nos permite asignarles
	// diferentes tipos de structs (ej. models.Invoice, models.Purchase) sin cambiar
	// la definición de FullTransactionDetail.
	Document interface{} `json:"document"`
	Items    interface{} `json:"items"`

	// 'Customer', 'Seller' y 'Supplier' son punteros a structs.
	// Usamos punteros (*) para que puedan tener un valor 'nil' si no aplican.
	// La etiqueta `json:"...,omitempty"` le dice al codificador JSON de Go
	// que si el campo es nulo (nil), debe omitirlo de la respuesta JSON.
	Customer *models.Customer `json:"customer,omitempty"`
	Seller   *models.Seller   `json:"seller,omitempty"`
	Supplier *models.Supplier `json:"supplier,omitempty"`
}

// GetTransactionDetail obtiene todos los datos relacionados con un único documento.
// LÓGICA: Esta función es un ejemplo de cómo agregar datos de múltiples fuentes (facturas, items, clientes, etc.)
// en una sola respuesta cohesiva para el frontend.
func GetTransactionDetail(client *api.SaintClient, docType, docID string) (*FullTransactionDetail, error) {
	switch strings.ToLower(docType) {
	case "invoice":
		// --- 1. Obtención de Datos Maestros en Paralelo ---
		var wg sync.WaitGroup
		var allInvoiceItems []models.InvoiceItem
		var allCustomers []models.Customer
		var allSellers []models.Seller

		wg.Add(3)
		go func() { defer wg.Done(); allInvoiceItems, _ = client.GetInvoiceItems() }()
		go func() { defer wg.Done(); allCustomers, _ = client.GetCustomers() }()
		go func() { defer wg.Done(); allSellers, _ = client.GetSellers() }()
		wg.Wait()

		// --- 2. Búsqueda del Documento Principal ---
		allInvoices, err := client.GetInvoices()
		if err != nil {
			return nil, err
		}

		var document models.Invoice
		found := false
		for _, inv := range allInvoices {
			if inv.NumeroD != nil && *inv.NumeroD == docID {
				document = inv
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("factura con NumeroD '%s' no encontrada", docID)
		}

		// --- 3. Búsqueda de Datos Relacionados ---
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
					temp := c
					customer = &temp
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
		return &FullTransactionDetail{Document: document, Items: items, Customer: customer, Seller: seller}, nil
	}

	return nil, fmt.Errorf("obtener detalle para el tipo '%s' aún no está implementado", docType)
}

// GetEntityDetail busca los detalles de una entidad específica (cliente, producto, vendedor).
func GetEntityDetail(client *api.SaintClient, entityType, entityID string) (interface{}, error) {
	switch strings.ToLower(entityType) {
	case "customer":
		allCustomers, err := client.GetCustomers()
		if err != nil {
			return nil, err
		}
		for _, c := range allCustomers {
			if c.CodClie != nil && *c.CodClie == entityID {
				return c, nil
			}
		}
		return nil, fmt.Errorf("cliente con ID '%s' no fue encontrado", entityID)

	case "seller":
		allSellers, err := client.GetSellers()
		if err != nil {
			return nil, err
		}
		for _, s := range allSellers {
			if s.CodVend != nil && *s.CodVend == entityID {
				return s, nil
			}
		}
		return nil, fmt.Errorf("vendedor con ID '%s' no fue encontrado", entityID)

	case "product":
		allProducts, err := client.GetProducts()
		if err != nil {
			return nil, err
		}
		for _, p := range allProducts {
			if p.CodProd != nil && *p.CodProd == entityID {
				return p, nil
			}
		}
		return nil, fmt.Errorf("producto con ID '%s' no fue encontrado", entityID)
	}
	return nil, fmt.Errorf("tipo de entidad '%s' no es válido", entityType)
}
