package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

type Message struct {
	// Aquí solo guardamos los IDs de las películas seleccionadas por el usuario
	MovieIDs []int `json:"movieIds"`
}

var (
	clients   = make(map[*websocket.Conn]bool) // Mapa para los clientes WebSocket conectados
	broadcast = make(chan Message)             // Canal para transmitir mensajes
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true }, // Permite cualquier origen
	}
	mu sync.Mutex // Mutex para sincronizar el acceso a la variable clients
)

// Maneja las conexiones WebSocket
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Actualiza la conexión HTTP a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error al actualizar a WebSocket: %v\n", err)
		return
	}
	defer ws.Close()

	// Agrega el cliente a la lista de clientes conectados
	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	// Lee los mensajes del WebSocket (aunque en este caso solo estamos esperando las recomendaciones)
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			// Si hay un error (por ejemplo, la conexión se cierra), eliminamos al cliente de la lista
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			break
		}
	}
}

// Maneja los mensajes enviados desde la API REST (IDs de películas seleccionadas por el usuario)
func handleAPI(w http.ResponseWriter, r *http.Request) {
	var msg Message
	// Decodificamos el cuerpo de la solicitud que contiene los IDs de las películas seleccionadas
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Error al decodificar el mensaje", http.StatusBadRequest)
		return
	}

	// Aquí normalmente harías el procesamiento de las recomendaciones.
	// Simulamos que obtenemos algunas recomendaciones basadas en los IDs recibidos:
	recommendations := []Message{
		{MovieIDs: []int{40, 10, 13}}, // Simulamos una recomendación
	}

	// Emitimos las recomendaciones a todos los clientes conectados a través de WebSocket
	for _, recommendation := range recommendations {
		broadcast <- recommendation
	}

	// Respondemos con un estado de éxito
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

// Envía los mensajes (recomendaciones) a todos los clientes WebSocket conectados
func handleMessages() {
	for {
		// Esperamos a que llegue un mensaje del canal "broadcast"
		msg := <-broadcast

		mu.Lock()
		// Enviamos el mensaje a todos los clientes conectados
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error al enviar mensaje: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func main() {
	// Configuración de CORS usando la configuración predeterminada (permitir todos los orígenes)
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handleConnections) // Conexión WebSocket
	mux.HandleFunc("/api", handleAPI)        // API REST para recibir los IDs de películas seleccionadas

	// Aplica CORS a todas las rutas
	handler := cors.Default().Handler(mux)

	// Inicia la goroutine que maneja los mensajes
	go handleMessages()

	// Inicia el servidor en el puerto 8080
	fmt.Println("Servidor iniciado en :8080")
	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal("Error en el servidor: ", err)
	}
}
