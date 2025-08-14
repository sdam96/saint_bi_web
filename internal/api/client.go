package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"saintnet.com/m/internal/models"
)

type SaintClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Pragma     string
}

func NewSaintClient(baseURL string) *SaintClient {
	return &SaintClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

// Login se autentica contra la API y guarda el token Pragma.
func (c *SaintClient) Login(username, password, apiKey, apiID string) error {
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	body := map[string]string{"terminal": "SAINT_BI_WEB"}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/main/login", c.BaseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+credentials)
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("x-api-id", apiID)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("fallo al iniciar sesión, status: %s", res.Status)
	}

	c.Pragma = res.Header.Get("Pragma")
	if c.Pragma == "" {
		return fmt.Errorf("no se recibió el token Pragma")
	}

	return nil
}

// genericGet realiza una solicitud GET a un endpoint y decodifica la respuesta.
func (c *SaintClient) genericGet(endpoint string, target interface{}) error {
	if c.Pragma == "" {
		return fmt.Errorf("no autenticado, falta el token Pragma")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/adm/%s", c.BaseURL, endpoint), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Pragma", c.Pragma)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error en la solicitud a %s: %s, body: %s", endpoint, res.Status, string(bodyBytes))
	}

	return json.NewDecoder(res.Body).Decode(target)
}

// --- Funciones para obtener datos de la API ---

func (c *SaintClient) GetInvoices() ([]models.Invoice, error) {
	var invoices []models.Invoice
	err := c.genericGet("invoices", &invoices)
	return invoices, err
}

func (c *SaintClient) GetInvoiceItems() ([]models.InvoiceItem, error) {
	var items []models.InvoiceItem
	err := c.genericGet("invoiceitems", &items)
	return items, err
}

func (c *SaintClient) GetPurchases() ([]models.Purchase, error) {
	var purchases []models.Purchase
	err := c.genericGet("purchases", &purchases)
	return purchases, err
}

func (c *SaintClient) GetAccReceivables() ([]models.AccReceivable, error) {
	var accReceivables []models.AccReceivable
	err := c.genericGet("accreceivables", &accReceivables)
	return accReceivables, err
}

func (c *SaintClient) GetAccPayables() ([]models.AccPayable, error) {
	var accPayables []models.AccPayable
	err := c.genericGet("accpayables", &accPayables)
	return accPayables, err
}

func (c *SaintClient) GetProducts() ([]models.Product, error) {
	var products []models.Product
	err := c.genericGet("products", &products)
	return products, err
}

func (c *SaintClient) GetCustomers() ([]models.Customer, error) {
	var customers []models.Customer
	err := c.genericGet("customers", &customers)
	return customers, err
}

func (c *SaintClient) GetSellers() ([]models.Seller, error) {
	var sellers []models.Seller
	err := c.genericGet("sellers", &sellers)
	return sellers, err
}
