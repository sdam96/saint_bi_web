// La declaración 'package handlers' indica que este archivo pertenece al paquete 'handlers'.
// En Go, todos los archivos dentro de un mismo directorio deben pertenecer al mismo paquete.
// Este archivo servirá como un punto central para las variables o funciones
// compartidas a través de todos los manejadores (handlers).
package handlers

import (
	// Se importa el paquete 'api' para poder utilizar el tipo 'api.SaintClient'.
	// Esto nos permite crear un mapa de clientes de API.
	"saintnet.com/m/internal/api"
)

// 'var activeClients = make(map[int]*api.SaintClient)' declara una variable a nivel de paquete.
//
// Análisis de la Sintaxis:
//   - 'var': Es la palabra clave en Go para declarar una variable.
//   - 'activeClients': Es el nombre que le damos a nuestra variable.
//   - 'make(...)': Es una función integrada en Go que se usa para inicializar
//     tipos de referencia como mapas, slices y canales.
//   - 'map[int]*api.SaintClient': Este es el tipo de dato.
//   - 'map': Indica que estamos creando un mapa (conocido como diccionario o hash table en otros lenguajes).
//   - '[int]': Define el tipo de dato para las claves del mapa. En este caso, usaremos el ID del usuario (un entero).
//   - '*api.SaintClient': Define el tipo de dato para los valores del mapa. Será un puntero a una
//     instancia de SaintClient.
//
// Lógica de Negocio:
//
//	Esta variable 'activeClients' actuará como una caché en memoria. Almacenará los clientes
//	de la API que ya han iniciado sesión. La próxima vez que un usuario haga una solicitud,
//	en lugar de volver a iniciar sesión contra la API externa, reutilizaremos el cliente
//	almacenado aquí, mejorando significativamente el rendimiento y la velocidad de respuesta.
//	Al ser una variable a nivel de paquete, es accesible por CUALQUIER otro archivo .go
//	dentro del mismo paquete 'handlers' (como middleware.go, dashboard.go, etc.).
var activeClients = make(map[int]*api.SaintClient)
