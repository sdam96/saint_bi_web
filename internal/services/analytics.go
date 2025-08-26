// internal/services/analytics.go
package services

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"saintnet.com/m/internal/api"
	"saintnet.com/m/internal/models"
)

// (Las structs SalesDataPoint y SalesForecast no cambian)
type SalesDataPoint struct {
	Date  time.Time `json:"date"`
	Sales float64   `json:"sales"`
}

type SalesForecast struct {
	HistoricalData []SalesDataPoint `json:"historicalData"`
	TrendLine      []SalesDataPoint `json:"trendLine"`
}

// calculateForecastLogic es una nueva función privada que contiene la lógica matemática.
// No llama a la API, solo procesa las facturas que se le pasan como argumento.
func calculateForecastLogic(invoices []models.Invoice, start, end time.Time) *SalesForecast {
	salesByDay := make(map[time.Time]float64)
	for _, inv := range invoices {
		if inv.FechaE != nil && inv.MtoTotal != nil {
			date, err := time.Parse("2006-01-02 15:04:05", *inv.FechaE)
			if err == nil && !date.Before(start) && !date.After(end) {
				day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
				salesByDay[day] += *inv.MtoTotal
			}
		}
	}

	if len(salesByDay) < 2 {
		return &SalesForecast{HistoricalData: []SalesDataPoint{}, TrendLine: []SalesDataPoint{}}
	}

	var dataPoints []SalesDataPoint
	for date, sales := range salesByDay {
		dataPoints = append(dataPoints, SalesDataPoint{Date: date, Sales: sales})
	}
	sort.Slice(dataPoints, func(i, j int) bool { return dataPoints[i].Date.Before(dataPoints[j].Date) })

	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(dataPoints))
	for i, p := range dataPoints {
		x := float64(i)
		sumX += x
		sumY += p.Sales
		sumXX += x * x
		sumXY += x * p.Sales
	}

	m := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	b := (sumY - m*sumX) / n

	var trendLine []SalesDataPoint
	for i, p := range dataPoints {
		forecastedSales := m*float64(i) + b
		trendLine = append(trendLine, SalesDataPoint{Date: p.Date, Sales: forecastedSales})
	}

	return &SalesForecast{
		HistoricalData: dataPoints,
		TrendLine:      trendLine,
	}
}

// CalculateSalesForecast ahora es más simple. Obtiene los datos de UNA conexión
// y los pasa a la función de lógica.
func CalculateSalesForecast(client *api.SaintClient, start, end time.Time) (*SalesForecast, error) {
	invoices, err := client.GetInvoices()
	if err != nil {
		return nil, err
	}
	forecast := calculateForecastLogic(invoices, start, end)
	return forecast, nil
}

// **NUEVA FUNCIÓN**
// GetConsolidatedSalesForecast obtiene facturas de TODAS las conexiones en paralelo
// y luego pasa la lista combinada a la misma función de lógica.
func GetConsolidatedSalesForecast(connections []models.Connection, start, end time.Time) (*SalesForecast, error) {
	var allInvoices []models.Invoice
	var invoicesMutex sync.Mutex
	var wg sync.WaitGroup
	errs := make(chan error, len(connections))

	for _, conn := range connections {
		wg.Add(1)
		go func(c models.Connection) {
			defer wg.Done()
			client := api.NewSaintClient(c.ApiURL)
			// Las credenciales de la API externa (apiKey, apiID) podrían estar hardcodeadas o venir de la configuración.
			if err := client.Login(c.ApiUser, c.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093"); err != nil {
				errs <- fmt.Errorf("error al iniciar sesión en '%s': %w", c.Alias, err)
				return
			}

			invoices, err := client.GetInvoices()
			if err != nil {
				errs <- err
				return
			}
			invoicesMutex.Lock()
			allInvoices = append(allInvoices, invoices...)
			invoicesMutex.Unlock()
		}(conn)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return nil, err // Si alguna conexión falla, devolvemos el error.
		}
	}

	// Llama a la misma lógica central, pero con los datos combinados.
	forecast := calculateForecastLogic(allInvoices, start, end)
	return forecast, nil
}

type MarketBasketResult struct {
	ItemA      string  `json:"itemA"`
	ItemB      string  `json:"itemB"`
	Confidence float64 `json:"confidence"`
	Support    float64 `json:"support"`
}

func calculateMarketBasketLogic(items []models.InvoiceItem, products []models.Product, start, end time.Time) []MarketBasketResult {
	productNames := make(map[string]string)
	for _, p := range products {
		if p.CodProd != nil && p.Descrip != nil {
			productNames[*p.CodProd] = *p.Descrip
		}
	}

	var filteredItems []models.InvoiceItem

	if !start.IsZero() && !end.IsZero() {
		for _, item := range items {
			if item.FechaE != nil {
				date, err := time.Parse("2006-01-02 15:04:05", *item.FechaE)
				if err == nil && !date.Before(start) && !date.After(end) {
					filteredItems = append(filteredItems, item)
				}
			}
		}
	} else {
		filteredItems = items
	}

	baskets := make(map[string]map[string]bool)
	for _, item := range filteredItems {
		if item.NumeroD != nil && item.CodItem != nil {
			if _, ok := baskets[*item.NumeroD]; !ok {
				baskets[*item.NumeroD] = make(map[string]bool)
			}
			baskets[*item.NumeroD][*item.CodItem] = true
		}
	}
	itemCounts := make(map[string]int)
	pairCounts := make(map[string]int)
	totalBaskets := float64(len(baskets))

	if totalBaskets == 0 {
		return []MarketBasketResult{}
	}

	for _, basket := range baskets {
		var basketItems []string
		for item := range basket {
			basketItems = append(basketItems, item)
			itemCounts[item]++
		}
		for i := 0; i < len(basketItems)-1; i++ {
			for j := i + 1; j < len(basketItems); j++ {
				pairKey1 := fmt.Sprintf("%s|%s", basketItems[i], basketItems[j])
				pairKey2 := fmt.Sprintf("%s|%s", basketItems[j], basketItems[i])
				pairCounts[pairKey1]++
				pairCounts[pairKey2]++
			}
		}
	}

	var results []MarketBasketResult
	for pair, count := range pairCounts {
		itemsInPair := strings.Split(pair, "|")
		itemA, itemB := itemsInPair[0], itemsInPair[1]

		support := float64(count) / totalBaskets
		confidence := float64(count) / float64(itemCounts[itemA])

		if confidence > 0.1 && support > 0.01 {
			nameA, okA := productNames[itemA]
			nameB, okB := productNames[itemB]
			if okA && okB {
				results = append(results, MarketBasketResult{
					ItemA: nameA, ItemB: nameB, Confidence: confidence, Support: support,
				})
			}
		}
	}

	sort.Slice(results, func(i, j int) bool { return results[i].Confidence > results[j].Confidence })

	if len(results) > 20 {
		return results[:20]
	}
	return results
}

func CalculateMarketBasket(client *api.SaintClient, start, end time.Time) ([]MarketBasketResult, error) {
	var items []models.InvoiceItem
	var products []models.Product
	var wg sync.WaitGroup
	errs := make(chan error, 2)

	wg.Add(2)
	go func() { defer wg.Done(); items, _ = client.GetInvoiceItems() }()
	go func() { defer wg.Done(); products, _ = client.GetProducts() }()
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return nil, err
		}
	}
	// Se pasan las fechas a la lógica de cálculo.
	return calculateMarketBasketLogic(items, products, start, end), nil
}

func GetConsolidatedMarketBasket(connections []models.Connection, start, end time.Time) ([]MarketBasketResult, error) {
	var allItems []models.InvoiceItem
	var allProducts []models.Product
	var itemsMutex, productsMutex sync.Mutex
	var wg sync.WaitGroup
	errs := make(chan error, len(connections))

	for _, conn := range connections {
		wg.Add(1)
		go func(c models.Connection) {
			defer wg.Done()
			client := api.NewSaintClient(c.ApiURL)
			if err := client.Login(c.ApiUser, c.ApiPassword, "B5D31933-C996-476C-B116-EF212A41479A", "1093"); err != nil {
				errs <- err
				return
			}

			items, err := client.GetInvoiceItems()
			if err != nil {
				errs <- err
				return
			}
			itemsMutex.Lock()
			allItems = append(allItems, items...)
			itemsMutex.Unlock()

			products, err := client.GetProducts()
			if err != nil {
				errs <- err
				return
			}
			productsMutex.Lock()
			allProducts = append(allProducts, products...)
			productsMutex.Unlock()
		}(conn)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			return nil, err
		}
	}
	// Se pasan las fechas a la lógica de cálculo.
	return calculateMarketBasketLogic(allItems, allProducts, start, end), nil
}
