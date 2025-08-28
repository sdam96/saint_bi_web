// internal/services/summary.go

// La declaración 'package services' indica que este archivo pertenece al paquete 'services'.
// Este paquete encapsula la lógica de negocio principal de la aplicación, como los cálculos
// de KPIs y la orquestación de llamadas a la API externa.
package services

import (
	// "fmt" para formatear strings de error.
	"fmt"
	// "log" para registrar información y errores en la terminal.
	"log"
	// "sort" para ordenar los slices de los rankings (Top 5).
	"sort"
	// "sync" proporciona primitivas de sincronización como WaitGroups y Mutexes,
	// esenciales para manejar la concurrencia de forma segura.
	"sync"
	// "time" para parsear, comparar y manipular fechas y horas.
	"time"

	// Se importan nuestros paquetes internos.
	"saintnet.com/m/internal/api"    // Para interactuar con la API de Saint.
	"saintnet.com/m/internal/models" // Para usar las estructuras de datos del negocio.
)

// GetConsolidatedSummary orquesta el proceso de cálculo del resumen para la vista "Consolidada".
//
// LÓGICA:
// Esta función es altamente concurrente para maximizar el rendimiento.
// 1. Inicia sesión en TODAS las conexiones de la base de datos de forma paralela.
// 2. Una vez autenticadas, descarga TODOS los datos necesarios de CADA conexión, también en paralelo.
// 3. Une (merge) todos los datos de todas las conexiones en una única gran estructura de datos.
// 4. Calcula los KPIs para el período actual y el anterior usando esta estructura de datos consolidada.
// 5. Ensambla y devuelve el resumen comparativo final.
//
// PARÁMETROS:
//   - connections ([]models.Connection): Un slice con la información de todas las conexiones a procesar.
//   - currentStart, currentEnd, prevStart, prevEnd (time.Time): Los rangos de fecha para el período actual y el anterior.
//
// RETORNA:
//   - *models.ComparativeSummary: Un puntero al resumen comparativo completo.
//   - error: Un error si alguna de las operaciones (login, obtención de datos) falla.
func GetConsolidatedSummary(connections []models.Connection, currentStart, currentEnd, prevStart, prevEnd time.Time) (*models.ComparativeSummary, error) {
	// --- 1. Autenticación en Paralelo ---
	// SINTAXIS de Go: `make([]*api.SaintClient, 0, len(connections))` crea un slice de punteros a SaintClient.
	// Se preasigna la capacidad (`len(connections)`) para mejorar la eficiencia y evitar múltiples reasignaciones de memoria.
	clients := make([]*api.SaintClient, 0, len(connections))
	// Un Mutex (Exclusión Mutua) es un candado que previene que múltiples goroutines escriban
	// en el mismo slice (`clients`) al mismo tiempo, evitando condiciones de carrera.
	var clientsMutex sync.Mutex
	var wgClients sync.WaitGroup
	// Un canal (chan) es un conducto tipado a través del cual puedes enviar y recibir valores.
	// Se usa para comunicar errores desde las goroutines de vuelta a la función principal.
	errs := make(chan error, len(connections))

	// SINTAXIS de Go: `for _, conn := range connections` itera sobre el slice de conexiones.
	// El guion bajo `_` descarta el índice, ya que solo nos interesa el valor `conn`.
	for _, conn := range connections {
		// `wgClients.Add(1)` incrementa el contador del WaitGroup.
		wgClients.Add(1)
		// `go func(...)` inicia una nueva goroutine (un hilo de ejecución ligero).
		// El código dentro de la función se ejecutará concurrentemente.
		go func(c models.Connection) {
			// `defer wgClients.Done()` decrementa el contador del WaitGroup cuando la goroutine termina.
			defer wgClients.Done()
			client := api.NewSaintClient(c.ApiURL)
			if err := client.Login(c.ApiUser, c.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093"); err != nil {
				// Si el login falla, se envía el error al canal. `%w` envuelve el error original.
				errs <- fmt.Errorf("error al iniciar sesión en '%s': %w", c.Alias, err)
				return
			}
			// Se bloquea el mutex antes de escribir en el slice compartido.
			clientsMutex.Lock()
			clients = append(clients, client)
			clientsMutex.Unlock() // Se desbloquea inmediatamente después.
		}(conn)
	}
	// `wgClients.Wait()` bloquea la ejecución hasta que el contador del WaitGroup llegue a cero,
	// es decir, hasta que todas las goroutines de login hayan terminado.
	wgClients.Wait()
	close(errs) // Se cierra el canal para indicar que no se enviarán más errores.

	// Se leen todos los errores del canal.
	for err := range errs {
		if err != nil {
			return nil, err
		}
	}

	// --- 2. Obtención de Datos en Paralelo ---
	// LÓGICA: Este bloque es el corazón de la consolidación. Su objetivo es descargar todos los
	// datos (facturas, productos, clientes, etc.) de todos los clientes de API autenticados
	// de forma simultánea y unirlos en una única estructura de datos (`allData`).

	// SINTAXIS de Go: `allData := &apiData{}` crea una nueva instancia de la struct `apiData`
	// y `allData` se convierte en un puntero a esa instancia. Usamos un puntero para que la función
	// `mergeApiData` pueda modificar directamente la instancia original.
	allData := &apiData{}
	// `var dataMutex sync.Mutex` declara un Mutex para proteger el acceso concurrente a `allData`.
	// Es crucial porque múltiples goroutines intentarán escribir en `allData` al mismo tiempo.
	var dataMutex sync.Mutex
	// `var wgData sync.WaitGroup` declara un WaitGroup para sincronizar la finalización de
	// todas las goroutines de descarga de datos.
	var wgData sync.WaitGroup
	// `dataErrs := make(chan error, len(clients))` crea un canal para recolectar errores
	// que puedan ocurrir durante la descarga de datos de cualquiera de las conexiones.
	dataErrs := make(chan error, len(clients))

	// Se itera sobre cada cliente de API que se autenticó exitosamente.
	for _, client := range clients {
		// `wgData.Add(1)` incrementa el contador, indicando que una nueva tarea concurrente va a comenzar.
		wgData.Add(1)
		// `go func(cl *api.SaintClient) { ... }(client)` inicia una nueva goroutine.
		// SINTAXIS de Go: Esta es una "función anónima" o "closure". Le pasamos `client` como
		// argumento (`cl`) para asegurarnos de que cada goroutine capture la copia correcta
		// de la variable `client` del bucle. Esto previene un error común en Go donde todas
		// las goroutines podrían terminar usando la última variable del bucle.
		go func(cl *api.SaintClient) {
			// `defer wgData.Done()` asegura que el contador del WaitGroup se decremente
			// sin importar cómo termine la goroutine (con o sin error).
			defer wgData.Done()
			// `fetchAllAPIData(cl)` es la función que descarga TODOS los datos de UNA sola conexión.
			data, err := fetchAllAPIData(cl)
			if err != nil {
				// Si hay un error, se envía al canal de errores y la goroutine termina.
				dataErrs <- err
				return
			}
			// `dataMutex.Lock()` adquiere el "candado". Solo una goroutine puede ejecutar el código
			// entre Lock() y Unlock() a la vez. Esto es vital para evitar que se corrompan los datos
			// en el slice `allData` al ser modificado por múltiples hilos.
			dataMutex.Lock()
			// `mergeApiData` une los datos recién descargados (`data`) con el contenedor global (`allData`).
			mergeApiData(allData, data)
			// `dataMutex.Unlock()` libera el candado, permitiendo que otra goroutine pueda acceder.
			dataMutex.Unlock()
		}(client)
	}
	// `wgData.Wait()` detiene la ejecución de la función `GetConsolidatedSummary` hasta que
	// todas las goroutines de descarga hayan llamado a `wgData.Done()`.
	wgData.Wait()
	// `close(dataErrs)` cierra el canal de errores, señalando que ya no se enviarán más valores.
	// Esto es necesario para que el bucle `for...range` de abajo sepa cuándo terminar.
	close(dataErrs)

	// SINTAXIS de Go: `for err := range dataErrs` itera sobre el canal hasta que se cierre.
	// Recibirá cada error que haya sido enviado al canal.
	for err := range dataErrs {
		if err != nil {
			// Si se encuentra cualquier error, se detiene todo el proceso y se devuelve el error.
			return nil, err
		}
	}

	// --- 3. Cálculo de Resúmenes en Paralelo ---
	// LÓGICA: Una vez que tenemos todos los datos consolidados en `allData`, calculamos los
	// resúmenes para el período actual y el anterior. Hacemos esto también en paralelo,
	// ya que un cálculo puede ser independiente del otro, ahorrando tiempo.

	var wgSummaries sync.WaitGroup
	// Se declaran punteros a `ManagementSummary`. Serán `nil` inicialmente. Las goroutines
	// asignarán las structs calculadas a estas variables.
	var currentSummary, previousSummary *models.ManagementSummary
	var errCurrent, errPrevious error // Variables para capturar errores de cada goroutine.

	// Se le dice al WaitGroup que vamos a esperar a que 2 goroutines terminen.
	wgSummaries.Add(2)
	go func() {
		defer wgSummaries.Done()
		currentSummary, errCurrent = calculateSummaryForPeriod(allData, currentStart, currentEnd)
	}()
	go func() {
		defer wgSummaries.Done()
		previousSummary, errPrevious = calculateSummaryForPeriod(allData, prevStart, prevEnd)
	}()
	// Se espera a que ambas goroutines de cálculo terminen.
	wgSummaries.Wait()

	// Se comprueban los errores después de que ambas tareas han finalizado.
	if errCurrent != nil {
		return nil, errCurrent
	}
	if errPrevious != nil {
		return nil, errPrevious
	}

	// --- 4. Ensamblaje Final de la Respuesta ---
	// LÓGICA: Se construye la estructura `ComparativeSummary` final.
	// SINTAXIS de Go: `*currentSummary` y `*previousSummary` desreferencian los punteros,
	// obteniendo el valor de la struct `ManagementSummary` a la que apuntan. Esto es necesario
	// porque los campos `CurrentPeriod` y `PreviousPeriod` en `ComparativeSummary` son structs, no punteros.
	finalSummary := &models.ComparativeSummary{
		CurrentPeriod:            *currentSummary,
		PreviousPeriod:           *previousSummary,
		TotalNetSalesComparative: calculateComparativeData(currentSummary.TotalNetSales, previousSummary.TotalNetSales),
		GrossProfitComparative:   calculateComparativeData(currentSummary.GrossProfit, previousSummary.GrossProfit),
		AverageTicketComparative: calculateComparativeData(currentSummary.AverageTicket, previousSummary.AverageTicket),
	}

	log.Println("[INFO] Resumen consolidado calculado exitosamente.")
	return finalSummary, nil
}

// GetComparativeSummary es el orquestador principal del servicio para UNA SOLA conexión.
//
// LÓGICA:
// 1. Llama a `fetchAllAPIData` para obtener todos los datos de la API para el cliente proporcionado.
// 2. Calcula el resumen para el período actual y el anterior en paralelo para optimizar.
// 3. Ensambla la respuesta final, incluyendo los datos comparativos.
func GetComparativeSummary(client *api.SaintClient, currentStart, currentEnd, prevStart, prevEnd time.Time) (*models.ComparativeSummary, error) {
	allData, err := fetchAllAPIData(client)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	var currentSummary, previousSummary *models.ManagementSummary
	var errCurrent, errPrevious error

	wg.Add(2)
	go func() {
		defer wg.Done()
		currentSummary, errCurrent = calculateSummaryForPeriod(allData, currentStart, currentEnd)
	}()
	go func() {
		defer wg.Done()
		previousSummary, errPrevious = calculateSummaryForPeriod(allData, prevStart, prevEnd)
	}()
	wg.Wait()

	if errCurrent != nil {
		return nil, errCurrent
	}
	if errPrevious != nil {
		return nil, errPrevious
	}

	finalSummary := &models.ComparativeSummary{
		CurrentPeriod:            *currentSummary,
		PreviousPeriod:           *previousSummary,
		TotalNetSalesComparative: calculateComparativeData(currentSummary.TotalNetSales, previousSummary.TotalNetSales),
		GrossProfitComparative:   calculateComparativeData(currentSummary.GrossProfit, previousSummary.GrossProfit),
		AverageTicketComparative: calculateComparativeData(currentSummary.AverageTicket, previousSummary.AverageTicket),
	}

	log.Println("[INFO] Resumen gerencial comparativo calculado exitosamente.")
	return finalSummary, nil
}

// apiData es una struct interna que actúa como un contenedor temporal para todos los
// datos obtenidos de la API. Esto simplifica pasar los datos entre funciones.
type apiData struct {
	invoices     []models.Invoice
	invoiceItems []models.InvoiceItem
	purchases    []models.Purchase
	receivables  []models.AccReceivable
	payables     []models.AccPayable
	products     []models.Product
	customers    []models.Customer
	sellers      []models.Seller
}

// mergeApiData combina los datos de un `src` (fuente) en `dest` (destino).
//
// SINTAXIS de Go: `append(dest.invoices, src.invoices...)` es la forma de unir dos slices.
// El `...` (operador de expansión) despliega los elementos del slice `src` para que `append`
// los pueda añadir individualmente al slice `dest`.
func mergeApiData(dest *apiData, src *apiData) {
	dest.invoices = append(dest.invoices, src.invoices...)
	dest.invoiceItems = append(dest.invoiceItems, src.invoiceItems...)
	dest.purchases = append(dest.purchases, src.purchases...)
	dest.receivables = append(dest.receivables, src.receivables...)
	dest.payables = append(dest.payables, src.payables...)
	dest.products = append(dest.products, src.products...)
	dest.customers = append(dest.customers, src.customers...)
	dest.sellers = append(dest.sellers, src.sellers...)
}

// fetchAllAPIData ejecuta todas las llamadas a la API de forma concurrente para un único cliente.
//
// LÓGICA:
// Se crea un slice de funciones anónimas (`apiCalls`). Cada función representa una llamada a un endpoint
// de la API. Este patrón permite iterar sobre ellas y lanzar cada llamada en su propia goroutine,
// haciendo que todas las descargas de datos ocurran simultáneamente en lugar de una por una.
func fetchAllAPIData(client *api.SaintClient) (*apiData, error) {
	data := &apiData{}
	var wg sync.WaitGroup
	errs := make(chan error, 8) // Canal con buffer para 8 posibles errores.

	apiCalls := []func() error{
		func() error { var err error; data.invoices, err = client.GetInvoices(); return err },
		func() error { var err error; data.invoiceItems, err = client.GetInvoiceItems(); return err },
		func() error { var err error; data.purchases, err = client.GetPurchases(); return err },
		func() error { var err error; data.receivables, err = client.GetAccReceivables(); return err },
		func() error { var err error; data.payables, err = client.GetAccPayables(); return err },
		func() error { var err error; data.products, err = client.GetProducts(); return err },
		func() error { var err error; data.customers, err = client.GetCustomers(); return err },
		func() error { var err error; data.sellers, err = client.GetSellers(); return err },
	}

	for _, call := range apiCalls {
		wg.Add(1)
		go func(apiCall func() error) {
			defer wg.Done()
			if err := apiCall(); err != nil {
				errs <- err
			}
		}(call)
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			log.Printf("[ERROR] Falla al obtener datos de la API de Saint: %v", err)
			return nil, err
		}
	}
	return data, nil
}

// calculateSummaryForPeriod es el corazón del servicio. Contiene toda la lógica
// para calcular los KPIs a partir de un conjunto de datos ya cargado.
func calculateSummaryForPeriod(data *apiData, startDate, endDate time.Time) (*models.ManagementSummary, error) {
	summary := &models.ManagementSummary{}
	now := time.Now()

	// LÓGICA: Se crea un mapa para acceder rápidamente a las cabeceras de factura por su NumeroD.
	// Esto es mucho más eficiente que buscar en el slice de facturas repetidamente dentro de un bucle.
	invoiceHeaderMap := make(map[string]models.Invoice)
	for _, inv := range data.invoices {
		if inv.NumeroD != nil {
			invoiceHeaderMap[*inv.NumeroD] = inv
		}
	}

	// LÓGICA: Se filtran las facturas que están dentro del rango de fechas y se calculan los KPIs de ventas.
	var invoicesInPeriod []models.Invoice
	for _, inv := range data.invoices {
		if inv.FechaE != nil {
			// SINTAXIS de Go: `time.Parse` convierte un string a un objeto `time.Time` usando un layout de referencia.
			// Las funciones `date.Before()` y `date.After()` se usan para la comparación de fechas.
			if date, err := time.Parse("2006-01-02 15:04:05", *inv.FechaE); err == nil && !date.Before(startDate) && !date.After(endDate) {
				invoicesInPeriod = append(invoicesInPeriod, inv)
				// SINTAXIS de Go: Se debe comprobar que los punteros no sean `nil` antes de desreferenciarlos con `*`.
				if inv.MtoTotal != nil {
					summary.TotalNetSales += *inv.MtoTotal
				}
				if inv.Credito != nil {
					summary.TotalNetSalesCredit += *inv.Credito
				}
				if inv.Contado != nil {
					summary.TotalNetSalesCash += *inv.Contado
				}
				if inv.CostoPrd != nil {
					summary.CostOfGoodsSold += *inv.CostoPrd
				}
				if inv.MtoTax != nil {
					summary.SalesVAT += *inv.MtoTax
				}
				if inv.RetenIVA != nil {
					summary.SalesIVAWithheld += *inv.RetenIVA
				}
			}
		}
	}

	summary.TotalInvoices = len(invoicesInPeriod)
	if summary.TotalInvoices > 0 {
		summary.AverageTicket = summary.TotalNetSales / float64(summary.TotalInvoices)
	}

	summary.GrossProfit = summary.TotalNetSales - summary.CostOfGoodsSold
	if summary.TotalNetSales > 0 {
		summary.GrossProfitMargin = (summary.GrossProfit / summary.TotalNetSales) * 100
	}

	// LÓGICA: Se calculan los KPIs de Cuentas por Cobrar.
	for _, r := range data.receivables {
		if r.Saldo != nil && *r.Saldo > 0 {
			summary.TotalReceivables += *r.Saldo
			if r.FechaV != nil {
				if vencimiento, err := time.Parse("2006-01-02 15:04:05", *r.FechaV); err == nil && vencimiento.Before(now) {
					summary.OverdueReceivables += *r.Saldo
				}
			}
		}
	}

	daysInRange := endDate.Sub(startDate).Hours() / 24
	if summary.TotalNetSalesCredit > 0 {
		summary.ReceivablesTurnoverDays = (summary.TotalReceivables / summary.TotalNetSalesCredit) * daysInRange
	}
	if summary.TotalReceivables > 0 {
		summary.ReceivablePercentage = (summary.OverdueReceivables / summary.TotalReceivables) * 100
	}

	// LÓGICA: Se calculan los KPIs de Cuentas por Pagar e Impuestos.
	var totalPurchasesCredit float64
	for _, p := range data.purchases {
		if p.FechaE != nil {
			if date, err := time.Parse("2006-01-02 15:04:05", *p.FechaE); err == nil && !date.Before(startDate) && !date.After(endDate) {
				if p.Credito != nil {
					totalPurchasesCredit += *p.Credito
				}
				if p.MtoTax != nil {
					summary.PurchasesVAT += *p.MtoTax
				}
				if p.RetenIVA != nil {
					summary.PurchasesIVAWithheld += *p.RetenIVA
				}
			}
		}
	}

	for _, p := range data.payables {
		if p.Saldo != nil && *p.Saldo > 0 {
			summary.TotalPayables += *p.Saldo
			if p.FechaV != nil {
				if vencimiento, err := time.Parse("2006-01-02 15:04:05", *p.FechaV); err == nil && vencimiento.Before(now) {
					summary.OverduePayables += *p.Saldo
				}
			}
		}
	}
	if totalPurchasesCredit > 0 {
		summary.PayablesTurnoverDays = (summary.TotalPayables / totalPurchasesCredit) * daysInRange
	}

	summary.VATPayable = summary.SalesVAT - summary.PurchasesVAT

	// LÓGICA: Se calculan los totales generales, estos no dependen del rango de fechas.
	for _, c := range data.customers {
		if c.Activo != nil && *c.Activo == 1 {
			summary.TotalActiveClients++
			if c.Saldo != nil && *c.Saldo > 0 {
				summary.ActiveClientsWithDebt++
			}
		}
	}
	for _, p := range data.products {
		if p.Activo != nil && *p.Activo == 1 {
			summary.TotalActiveProducts++
		}
	}

	// LÓGICA: Se calculan todos los rankings Top 5 llamando a las funciones auxiliares.
	summary.Top5ClientsBySales = rankItems(calculateSalesByClient(data.invoiceItems, invoiceHeaderMap, data.customers))
	summary.Top5ProductsBySales = rankItems(calculateSalesByProduct(data.invoiceItems, data.products))
	summary.Top5SellersBySales = rankItems(calculateSalesBySeller(data.invoiceItems, invoiceHeaderMap, data.sellers))
	summary.Top5ProductsByProfit = rankItems(calculateProfitByProduct(data.invoiceItems, data.products))

	return summary, nil
}

// calculateComparativeData es una función auxiliar que calcula la diferencia porcentual entre dos valores.
func calculateComparativeData(current, previous float64) models.ComparativeData {
	data := models.ComparativeData{
		Value:         current,
		PreviousValue: previous,
	}
	if previous != 0 {
		data.PercentageChange = ((current - previous) / previous) * 100
	} else if current > 0 {
		// LÓGICA: Si el valor anterior era 0 y el actual es positivo, se considera un crecimiento del 100%.
		data.PercentageChange = 100
	}
	return data
}

// --- Funciones Auxiliares para Rankings ---
// LÓGICA: Las siguientes cuatro funciones (`calculate...`) tienen un patrón similar:
// 1. Crean un mapa de nombres (ej. CodClie a Descrip) para una búsqueda eficiente.
// 2. Crean un mapa para agregar los valores (ej. ventas por cliente).
// 3. Iteran sobre los datos relevantes (ej. items de factura) y suman los valores en el mapa de agregación.
// 4. Devuelven el mapa con los totales.

func calculateSalesByClient(items []models.InvoiceItem, headerMap map[string]models.Invoice, customers []models.Customer) map[string]float64 {
	salesMap := make(map[string]float64)
	nameMap := make(map[string]string)
	for _, c := range customers {
		if c.CodClie != nil && c.Descrip != nil {
			nameMap[*c.CodClie] = *c.Descrip
		}
	}
	for _, item := range items {
		if item.NumeroD == nil || item.TotalItem == nil {
			continue // Salta a la siguiente iteración si los datos necesarios son nulos.
		}
		if header, ok := headerMap[*item.NumeroD]; ok {
			if header.CodClie != nil {
				if clientName, nameOk := nameMap[*header.CodClie]; nameOk {
					salesMap[clientName] += *item.TotalItem
				}
			}
		}
	}
	return salesMap
}

func calculateSalesBySeller(items []models.InvoiceItem, headerMap map[string]models.Invoice, sellers []models.Seller) map[string]float64 {
	salesMap := make(map[string]float64)
	nameMap := make(map[string]string)
	for _, s := range sellers {
		if s.CodVend != nil && s.Descrip != nil {
			nameMap[*s.CodVend] = *s.Descrip
		}
	}
	for _, item := range items {
		if item.NumeroD == nil || item.TotalItem == nil {
			continue
		}
		if header, ok := headerMap[*item.NumeroD]; ok {
			if header.CodVend != nil {
				if sellerName, nameOk := nameMap[*header.CodVend]; nameOk {
					salesMap[sellerName] += *item.TotalItem
				}
			}
		}
	}
	return salesMap
}

func calculateSalesByProduct(items []models.InvoiceItem, products []models.Product) map[string]float64 {
	salesMap := make(map[string]float64)
	nameMap := make(map[string]string)
	for _, p := range products {
		if p.CodProd != nil && p.Descrip != nil {
			nameMap[*p.CodProd] = *p.Descrip
		}
	}
	for _, item := range items {
		if item.CodItem == nil || item.TotalItem == nil {
			continue
		}
		if productName, ok := nameMap[*item.CodItem]; ok {
			salesMap[productName] += *item.TotalItem
		}
	}
	return salesMap
}

func calculateProfitByProduct(items []models.InvoiceItem, products []models.Product) map[string]float64 {
	profitMap := make(map[string]float64)
	productMap := make(map[string]models.Product)
	for _, p := range products {
		if p.CodProd != nil {
			productMap[*p.CodProd] = p
		}
	}
	for _, item := range items {
		if item.CodItem == nil {
			continue
		}
		if product, ok := productMap[*item.CodItem]; ok {
			if item.Precio != nil && product.CostAct != nil && item.Cantidad != nil && product.Descrip != nil {
				profit := (*item.Precio - *product.CostAct) * *item.Cantidad
				profitMap[*product.Descrip] += profit
			}
		}
	}
	return profitMap
}

// rankItems convierte un mapa de resultados en un slice ordenado y lo corta a los 5 mejores.
func rankItems(itemsMap map[string]float64) []models.RankedItem {
	var ranked []models.RankedItem
	for name, value := range itemsMap {
		ranked = append(ranked, models.RankedItem{Name: name, Value: value})
	}

	// SINTAXIS de Go: `sort.Slice` ordena un slice usando una función de comparación personalizada (un closure).
	// La función devuelve `true` si el elemento en el índice `i` debe ir antes que el del índice `j`.
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Value > ranked[j].Value // Orden descendente
	})

	// SINTAXIS de Go: `ranked[:5]` es una "operación de slice" que devuelve un nuevo slice
	// conteniendo los elementos desde el inicio hasta el índice 4 (los 5 primeros).
	if len(ranked) > 5 {
		return ranked[:5]
	}
	return ranked
}
