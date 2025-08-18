// The 'package services' declaration indicates this file belongs to a package named 'services'.
// In a typical project structure, a 'services' package contains the core business logic,
// acting as an intermediary between the data layer (database, API clients) and the presentation layer (handlers).
package services

import (
	// "log" for printing informational messages.
	"log"
	// "sort" provides sorting algorithms. We use it to rank items like top clients and products.
	"sort"
	// "sync" provides synchronization primitives, such as WaitGroups and Mutexes. It's essential for managing concurrent operations.
	"sync"
	// "time" is used for handling dates and times, which is critical for filtering data by date ranges.
	"time"

	// Our internal packages for the API client and data models.
	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/models"
)

// CalculateManagementSummary obtains all necessary data from the API, performs calculations, and returns a consolidated summary.
// This function is a great example of orchestrating multiple tasks: concurrent data fetching, data processing, and aggregation.
func CalculateManagementSummary(client *api.SaintClient) (*models.ManagementSummary, error) {
	// --- Paso 1: Obtener todos los datos de la API en paralelo ---
	// La concurrencia es clave aquí para mejorar el rendimiento. En lugar de esperar por cada llamada a la API
	// una tras otra, las lanzamos todas al mismo timepo y esperamos a que todas terminen.

	// Declaración de variables que almacenarán los resultados de las llamadas a la API.
	var (
		invoices     []models.Invoice
		invoiceItems []models.InvoiceItem
		purchases    []models.Purchase
		receivables  []models.AccReceivable
		payables     []models.AccPayable
		products     []models.Product
		customers    []models.Customer
		sellers      []models.Seller
		// 'sync.WaitGroup' es una herramienta para esperar a que un conjunto de goroutines termine.
		// El flujo principal llama a 'Add' para establecer el número de goroutines a esperar,
		// y cada goroutine llama a 'Done' cuando termina. 'Wait' bloquea hasta que todas las goroutines hayan llamado a 'Done'.
		wg sync.WaitGroup
		// 'make(chan error, 8)' crea un canal buferizado para errores. Un canal es una tubería
		// a través de la cual las goroutines pueden comunicarse. Este canal puede contener hasta 8 errores
		// sin bloquear a la goroutine que envía el error. Esto es importante porque las goroutines
		// se ejecutan de forma independiente y no pueden devolver errores directamente.
		errs = make(chan error, 8)
	)

	// 'apiCalls' es un slice de funciones. Cada función envuelve una llamada a un método del cliente API.
	// Este patrón nos permite iterar y lanzar cada llamada de una manera genérica y limpia.
	apiCalls := []func() error{
		func() error { var err error; invoices, err = client.GetInvoices(); return err },
		func() error { var err error; invoiceItems, err = client.GetInvoiceItems(); return err },
		func() error { var err error; purchases, err = client.GetPurchases(); return err },
		func() error { var err error; receivables, err = client.GetAccReceivables(); return err },
		func() error { var err error; payables, err = client.GetAccPayables(); return err },
		func() error { var err error; products, err = client.GetProducts(); return err },
		func() error { var err error; customers, err = client.GetCustomers(); return err },
		func() error { var err error; sellers, err = client.GetSellers(); return err },
	}

	// Se itera sobre el slice de llamadas a la API para ejecutarlas concurrentemente.
	for _, call := range apiCalls {
		// 'wg.Add(1)' incrementa el contador del WaitGroup en uno por cada goroutine que vamos a lanzar.
		wg.Add(1)
		// 'go func(...)' inicia una nueva goroutine. La goroutine ejecuta la función anónima que le pasamos.
		// Es crucial pasar 'call' como un argumento a la goroutine ('(call)'). Si no lo hiciéramos,
		// todas las goroutines podrían terminar usando la última versión de 'call' en el bucle (un problema de clausura común).
		go func(apiCall func() error) {
			// 'defer wg.Done()' asegura que el contador del WaitGroup se decremente cuando la goroutine termine,
			// ya sea que la llamada a la API tenga éxito o falle.
			defer wg.Done()
			// Se ejecuta la llamada a la API.
			if err := apiCall(); err != nil {
				// Si hay un error, se envía al canal de errores 'errs'.
				errs <- err
			}
		}(call)
	}

	// 'wg.Wait()' bloquea la ejecución de la función 'CalculateManagementSummary' hasta que el contador
	// del WaitGroup sea cero, es decir, hasta que todas las goroutines hayan llamado a 'Done'.
	wg.Wait()
	// 'close(errs)' cierra el canal de errores. Esto es importante para que el bucle 'for range'
	// que viene a continuación sepa cuándo detenerse.
	close(errs)

	// Se itera sobre el canal de errores para comprobar si alguna de las llamadas falló.
	for err := range errs {
		if err != nil {
			// Si se encuentra un error, se registra y se devuelve inmediatamente,
			// deteniendo el cálculo del resumen.
			log.Printf("Error obteniendo datos de la API: %v", err)
			return nil, err
		}
	}

	// --- Paso 2: Realizar los cálculos ---
	// Si llegamos aquí, todos los datos se han obtenido con éxito.

	// Se inicializa un puntero a la struct 'ManagementSummary' que vamos a rellenar.
	summary := &models.ManagementSummary{}
	// Se obtienen las fechas que servirán como límites para los cálculos (ej. "últimos 30 días").
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// Crear mapas para facilitar la búsqueda de datos. Esto es una optimización de rendimiento.
	// En lugar de buscar en un slice una y otra vez (O(n)), creamos un mapa (O(1)) para un acceso rápido.
	invoiceHeaderMap := make(map[string]models.Invoice)
	for _, inv := range invoices {
		// Se verifica que los punteros no sean nulos antes de desreferenciarlos para evitar un 'panic'.
		if inv.NumeroD != nil {
			invoiceHeaderMap[*inv.NumeroD] = inv
		}
	}

	// Se calculan los totales de ventas y costos, filtrando por fecha.
	for _, inv := range invoices {
		if inv.FechaE != nil {
			// 'time.Parse' convierte el string de fecha del API al formato 'time.Time' de Go.
			// "2006-01-02 15:04:05" es la forma mnemotécnica en Go para especificar el layout de parseo.
			if date, err := time.Parse("2006-01-02 15:04:05", *inv.FechaE); err == nil && date.After(thirtyDaysAgo) {
				// Se acumulan los totales, siempre comprobando que los punteros no sean nulos.
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
			}
		}
	}
	summary.TotalInvoices = len(invoices)
	if summary.TotalInvoices > 0 {
		summary.AverageTicket = summary.TotalNetSales / float64(summary.TotalInvoices)
	}

	// Se calcula la utilidad y el margen bruto.
	summary.GrossProfit = summary.TotalNetSales - summary.CostOfGoodsSold
	if summary.TotalNetSales > 0 {
		summary.GrossProfitMargin = (summary.GrossProfit / summary.TotalNetSales) * 100
	}

	// Se calculan los KPIs de Cuentas por Cobrar.
	for _, r := range receivables {
		if r.Saldo != nil && *r.Saldo > 0 {
			summary.TotalReceivables += *r.Saldo
			if r.FechaV != nil {
				if vencimiento, err := time.Parse("2006-01-02 15:04:05", *r.FechaV); err == nil && vencimiento.Before(now) {
					summary.OverdueReceivables += *r.Saldo
				}
			}
		}
	}
	if summary.TotalNetSalesCredit > 0 {
		summary.ReceivablesTurnoverDays = (summary.TotalReceivables / summary.TotalNetSalesCredit) * 30
	}
	if summary.TotalReceivables > 0 {
		summary.ReceivablePercentage = (summary.OverdueReceivables / summary.TotalReceivables) * 100
	}

	// Se calculan los KPIs de Cuentas por Pagar.
	var totalPurchasesCredit float64
	for _, p := range purchases {
		if p.FechaE != nil {
			if date, err := time.Parse("2006-01-02 15:04:05", *p.FechaE); err == nil && date.After(thirtyDaysAgo) {
				if p.Credito != nil {
					totalPurchasesCredit += *p.Credito
				}
			}
		}
	}
	for _, p := range payables {
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
		summary.PayablesTurnoverDays = (summary.TotalPayables / totalPurchasesCredit) * 30
	}

	// Se cuentan clientes y productos activos.
	for _, c := range customers {
		if c.Activo != nil && *c.Activo == 1 {
			summary.TotalActiveClients++
			if c.Saldo != nil && *c.Saldo > 0 {
				summary.ActiveClientsWithDebt++
			}
		}
	}
	for _, p := range products {
		if p.Activo != nil && *p.Activo == 1 {
			summary.TotalActiveProducts++
		}
	}

	// Se calculan los totales de impuestos.
	for _, inv := range invoices {
		if inv.MtoTax != nil {
			summary.SalesVAT += *inv.MtoTax
		}
		if inv.RetenIVA != nil {
			summary.SalesIVAWithheld += *inv.RetenIVA
		}
	}
	for _, p := range purchases {
		if p.MtoTax != nil {
			summary.PurchasesVAT += *p.MtoTax
		}
		if p.RetenIVA != nil {
			summary.PurchasesIVAWithheld += *p.RetenIVA
		}
	}
	summary.VATPayable = summary.SalesVAT - summary.PurchasesVAT

	// --- Se calculan los Rankings Top 5 ---
	// Se delega la lógica de cálculo a funciones auxiliares para mantener el código más limpio.
	summary.Top5ClientsBySales = rankItems(calculateSalesByClient(invoiceItems, invoiceHeaderMap, customers))
	summary.Top5ProductsBySales = rankItems(calculateSalesByProduct(invoiceItems, products))
	summary.Top5SellersBySales = rankItems(calculateSalesBySeller(invoiceItems, invoiceHeaderMap, sellers))
	summary.Top5ProductsByProfit = rankItems(calculateProfitByProduct(invoiceItems, products))

	log.Println("Resumen gerencial completo calculado exitosamente.")
	return summary, nil
}

// Las siguientes son funciones auxiliares para calcular los rankings.
// Separan la lógica de agregación de datos para cada dimensión (cliente, producto, vendedor).

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
			continue
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

// rankItems toma un mapa de [string]float64, lo convierte en un slice de RankedItem,
// lo ordena de mayor a menor y devuelve los 5 mejores resultados.
func rankItems(itemsMap map[string]float64) []models.RankedItem {
	var ranked []models.RankedItem
	// Se convierte el mapa a un slice para poder ordenarlo.
	for name, value := range itemsMap {
		ranked = append(ranked, models.RankedItem{Name: name, Value: value})
	}

	// 'sort.Slice' ordena el slice 'ranked' in-situ.
	// Le pasamos una función anónima (una clausura) que define cómo comparar dos elementos.
	// La función devuelve 'true' si el elemento en el índice 'i' debe ir antes que el elemento en 'j'.
	// 'ranked[i].Value > ranked[j].Value' resulta en un orden descendente.
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].Value > ranked[j].Value
	})

	// Se devuelve solo el top 5.
	if len(ranked) > 5 {
		// La sintaxis de slicing 'ranked[:5]' devuelve un nuevo slice que contiene los elementos
		// desde el índice 0 hasta el 4 (el 5 no se incluye).
		return ranked[:5]
	}
	return ranked
}
