// La declaración 'package api' indica que este archivo pertenece al paquete 'api'.
// En Go, los paquetes son la forma de organizar y reutilizar código. Todos los archivos
// dentro del mismo directorio deben pertenecer al mismo paquete.
package api

// El bloque 'import' agrupa todas las dependencias externas que este archivo necesita.
// Go requiere que todos los paquetes importados sean utilizados en el código.
import (
	// "bytes" proporciona funciones para la manipulación de slices de bytes,
	// que usamos aquí para crear un buffer a partir de nuestro cuerpo de solicitud JSON.
	"bytes"
	// "encoding/base64" implementa la codificación y decodificación Base64,
	// útil para enviar credenciales de forma segura en las cabeceras HTTP.
	"encoding/base64"
	// "encoding/json" ofrece la funcionalidad para codificar (marshal) y decodificar (unmarshal)
	// datos en formato JSON.
	"encoding/json"
	// "fmt" implementa funciones para formatear I/O (entrada/salida), como imprimir en la consola
	// o construir strings con formato.
	"fmt"
	// "io" proporciona primitivas básicas de I/O. Aquí usamos 'ReadAll' para leer el cuerpo
	// de una respuesta HTTP en caso de error.
	"io"
	// "net/http" es el corazón de la funcionalidad de red en Go. Proporciona un cliente y servidor HTTP.
	"net/http"
	// "time" ofrece funcionalidades para medir y representar el tiempo. Lo usamos para
	// configurar un tiempo de espera (timeout) en nuestro cliente HTTP.
	"time"

	// "saintnet.com/m/internal/models" es una importación de un paquete interno del proyecto.
	// Esto sugiere que el proyecto sigue una estructura de directorios estándar en Go, donde 'internal'
	// contiene código que solo es accesible dentro de este proyecto. 'models' probablemente define
	// las estructuras de datos (structs) que representan los objetos de negocio.
	"saintnet.com/m/internal/models"
)

// SaintClient es una estructura (struct) que define un cliente para interactuar con la API de Saint.
// Una 'struct' en Go es un tipo de dato compuesto que agrupa cero o más campos de datos con nombre.
type SaintClient struct {
	// BaseURL es un campo de tipo 'string' que almacenará la URL base de la API.
	BaseURL string
	// HTTPClient es un puntero a un 'http.Client'. Usar un puntero (*) es eficiente
	// porque evita copiar la estructura completa del cliente cada vez que se pasa como argumento.
	// Este cliente gestionará todas las comunicaciones HTTP.
	HTTPClient *http.Client
	// Pragma es un campo de tipo 'string' que guardará el token de autenticación
	// recibido de la API después de un inicio de sesión exitoso.
	Pragma string
}

// NewSaintClient es una función constructora que crea y devuelve un nuevo puntero a un SaintClient.
// En Go, no hay constructores como en otros lenguajes (ej. __init__ en Python o un constructor de clase en Java).
// Es una convención idiomática en Go crear una función 'New<Tipo>' para inicializar un tipo de dato.
func NewSaintClient(baseURL string) *SaintClient {
	// La sintaxis '&SaintClient{...}' crea una nueva instancia de la struct SaintClient
	// y devuelve un puntero a ella. Esto se conoce como un "literal de struct compuesta".
	return &SaintClient{
		// Asigna el valor del parámetro 'baseURL' al campo 'BaseURL' de la struct.
		BaseURL: baseURL,
		// Inicializa el cliente HTTP.
		HTTPClient: &http.Client{
			// Establece un 'Timeout' (tiempo de espera) de un minuto para todas las solicitudes
			// hechas con este cliente. Esto previene que el programa se quede colgado
			// indefinidamente si la API no responde. 'time.Minute' es una constante del paquete 'time'.
			Timeout: time.Minute,
		},
	}
}

// Login se autentica contra la API y guarda el token Pragma.
// Esta es una función de método, asociada al tipo 'SaintClient'.
// La sintaxis '(c *SaintClient)' se llama "receptor" (receiver) y significa que este método
// opera sobre una instancia de 'SaintClient'. 'c' es el nombre que le damos a la instancia
// dentro del método. Se usa un puntero (*SaintClient) para poder modificar la instancia original (ej. guardar c.Pragma).
func (c *SaintClient) Login(username, password, apiKey, apiID string) error {
	// Concatena el usuario y la contraseña con un ":" y lo codifica en Base64.
	// Esto es parte del estándar de Autenticación Básica HTTP.
	// []byte(...) convierte el string a un slice de bytes, que es lo que la función EncodeToString espera.
	credentials := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

	// Crea un mapa de string a string para el cuerpo de la solicitud JSON.
	// map[keyType]valueType es la sintaxis para definir un mapa en Go.
	body := map[string]string{"terminal": "SAINT_BI_WEB"}
	// 'json.Marshal' convierte el mapa de Go a un slice de bytes en formato JSON.
	// El guion bajo '_' es el "identificador en blanco" (blank identifier) de Go, usado para descartar
	// valores de retorno que no necesitamos (en este caso, el posible error de Marshal).
	jsonBody, _ := json.Marshal(body)

	// 'http.NewRequest' crea una nueva solicitud HTTP.
	// "POST" es el método HTTP.
	// fmt.Sprintf construye la URL completa del endpoint de login.
	// 'bytes.NewBuffer(jsonBody)' crea un buffer en memoria que implementa la interfaz 'io.Reader', requerido por NewRequest para el cuerpo.
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/main/login", c.BaseURL), bytes.NewBuffer(jsonBody))
	// Es una buena práctica en Go manejar los errores inmediatamente después de que una función los devuelve.
	// Si 'err' no es 'nil', significa que ocurrió un error y la función retorna inmediatamente.
	if err != nil {
		return err
	}

	// Se establecen las cabeceras (headers) de la solicitud HTTP usando el método Set.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+credentials) // Autenticación Básica
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("x-api-id", apiID)

	// 'c.HTTPClient.Do(req)' envía la solicitud a la API y devuelve una respuesta y un posible error.
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	// 'defer res.Body.Close()' es una declaración clave en Go. 'defer' pospone la ejecución de una
	// llamada a función hasta que la función que la contiene ('Login' en este caso) está a punto de retornar.
	// Esto garantiza que el cuerpo de la respuesta se cierre siempre, liberando los recursos de red,
	// sin importar si la función termina con éxito o con un error. Es la forma idiomática de gestionar recursos.
	defer res.Body.Close()

	// Comprueba si el código de estado de la respuesta no es 200 OK.
	// http.StatusOK es una constante del paquete net/http para mayor legibilidad.
	if res.StatusCode != http.StatusOK {
		// 'fmt.Errorf' crea un nuevo objeto de error con un mensaje formateado.
		// Permite envolver errores con más contexto.
		return fmt.Errorf("fallo al iniciar sesión, status: %s", res.Status)
	}

	// Obtiene el valor de la cabecera "Pragma" de la respuesta.
	c.Pragma = res.Header.Get("Pragma")
	// Se valida que el token Pragma no esté vacío.
	if c.Pragma == "" {
		return fmt.Errorf("no se recibió el token Pragma")
	}

	// Devuelve 'nil' para indicar que no hubo errores. En Go, 'nil' es el valor cero para punteros,
	// interfaces, mapas, slices, canales y funciones, y comúnmente se usa para indicar la ausencia de error.
	return nil
}

// genericGet realiza una solicitud GET a un endpoint y decodifica la respuesta.
// Esta función es un buen ejemplo de cómo evitar la duplicación de código (principio DRY).
// 'endpoint' es el path específico del recurso que queremos obtener.
// 'target' es una interfaz vacía ('interface{}'), lo que permite pasar cualquier tipo de dato.
// Se espera que 'target' sea un puntero a la variable donde queremos almacenar el resultado decodificado.
func (c *SaintClient) genericGet(endpoint string, target interface{}) error {
	// Comprueba si el cliente está autenticado antes de hacer una solicitud.
	if c.Pragma == "" {
		return fmt.Errorf("no autenticado, falta el token Pragma")
	}

	// Crea una solicitud GET. Como no hay cuerpo, el último argumento es 'nil'.
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/v1/adm/%s", c.BaseURL, endpoint), nil)
	if err != nil {
		return err
	}
	// Establece la cabecera Pragma para la autenticación en las solicitudes subsiguientes.
	req.Header.Set("Pragma", c.Pragma)

	// Envía la solicitud.
	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	// Garantiza el cierre del cuerpo de la respuesta.
	defer res.Body.Close()

	// Si la solicitud no fue exitosa...
	if res.StatusCode != http.StatusOK {
		// ...lee el cuerpo de la respuesta para dar más contexto al error.
		// 'io.ReadAll' lee del 'res.Body' (que es un io.Reader) hasta EOF.
		bodyBytes, _ := io.ReadAll(res.Body)
		// Devuelve un error detallado que incluye el endpoint, el estado y el cuerpo de la respuesta.
		// 'string(bodyBytes)' convierte el slice de bytes a un string para que sea legible.
		return fmt.Errorf("error en la solicitud a %s: %s, body: %s", endpoint, res.Status, string(bodyBytes))
	}

	// 'json.NewDecoder(res.Body)' crea un decodificador que lee directamente del stream del cuerpo de la respuesta.
	// '.Decode(target)' decodifica el JSON en la variable apuntada por 'target'.
	// Esto es más eficiente que leer todo el cuerpo en memoria primero y luego decodificarlo.
	return json.NewDecoder(res.Body).Decode(target)
}

// --- Funciones para obtener datos de la API ---
// Esta sección contiene métodos públicos que exponen de forma amigable la funcionalidad
// de 'genericGet' para endpoints específicos. Esto crea una API de cliente clara y fácil de usar.

// GetInvoices obtiene una lista de facturas.
// Devuelve un slice de 'models.Invoice' y un error.
// La sintaxis de retorno '([]models.Invoice, error)' define múltiples valores de retorno, un patrón muy común en Go.
func (c *SaintClient) GetInvoices() ([]models.Invoice, error) {
	// 'var invoices []models.Invoice' declara un slice vacío de Invoice. Un slice es una vista flexible y
	// dinámica sobre un array subyacente.
	var invoices []models.Invoice
	// Llama a genericGet, pasando el nombre del endpoint y un puntero a 'invoices' (&invoices) para que
	// la función pueda modificar la variable 'invoices' original y llenarla con los datos decodificados.
	err := c.genericGet("invoices", &invoices)
	// Devuelve el slice poblado (o vacío si no hay datos) y el error (que será 'nil' si todo fue bien).
	return invoices, err
}

// GetInvoiceItems obtiene los ítems de las facturas.
func (c *SaintClient) GetInvoiceItems() ([]models.InvoiceItem, error) {
	var items []models.InvoiceItem
	err := c.genericGet("invoiceitems", &items)
	return items, err
}

// GetPurchases obtiene una lista de compras.
func (c *SaintClient) GetPurchases() ([]models.Purchase, error) {
	var purchases []models.Purchase
	err := c.genericGet("purchases", &purchases)
	return purchases, err
}

// GetAccReceivables obtiene las cuentas por cobrar.
func (c *SaintClient) GetAccReceivables() ([]models.AccReceivable, error) {
	var accReceivables []models.AccReceivable
	err := c.genericGet("accreceivables", &accReceivables)
	return accReceivables, err
}

// GetAccPayables obtiene las cuentas por pagar.
func (c *SaintClient) GetAccPayables() ([]models.AccPayable, error) {
	var accPayables []models.AccPayable
	err := c.genericGet("accpayables", &accPayables)
	return accPayables, err
}

// GetProducts obtiene la lista de productos.
func (c *SaintClient) GetProducts() ([]models.Product, error) {
	var products []models.Product
	err := c.genericGet("products", &products)
	return products, err
}

// GetCustomers obtiene la lista de clientes.
func (c *SaintClient) GetCustomers() ([]models.Customer, error) {
	var customers []models.Customer
	err := c.genericGet("customers", &customers)
	return customers, err
}

// GetSellers obtiene la lista de vendedores.
func (c *SaintClient) GetSellers() ([]models.Seller, error) {
	var sellers []models.Seller
	err := c.genericGet("sellers", &sellers)
	return sellers, err
}
