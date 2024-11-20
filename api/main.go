package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

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

// handleAPI maneja las solicitudes de recomendaciones
func handleAPI(w http.ResponseWriter, r *http.Request) {
	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Error al decodificar el mensaje", http.StatusBadRequest)
		return
	}

	fmt.Printf("Películas recibidas en la API: %v\n", msg.MovieIDs)

	// Envía los IDs de películas favoritas al servidor de recomendaciones
	recommendations, err := requestRecommendations(msg.MovieIDs)
	if err != nil {
		http.Error(w, "Error al obtener recomendaciones", http.StatusInternalServerError)
		return
	}

	fmt.Printf("Recomendaciones enviadas por el nodo servidor: %v\n", recommendations)

	// Enviamos las recomendaciones a los clientes conectados por WebSocket
	broadcast <- Message{MovieIDs: recommendations}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(Message{MovieIDs: recommendations})
}

// requestRecommendations conecta al servidor en el puerto TCP 9002 y obtiene recomendaciones
func requestRecommendations(favoriteIDs []int) ([]int, error) {
	// Conecta al servidor de recomendaciones en el puerto 9002
	conn, err := net.Dial("tcp", "172.20.0.5:9002")
	if err != nil {
		log.Printf("Error al conectar con el servidor de recomendaciones: %v", err)
		return nil, err
	}
	defer conn.Close()

	// Envía los IDs de películas favoritas como JSON
	data, err := json.Marshal(favoriteIDs)
	if err != nil {
		log.Printf("Error al serializar los IDs de películas favoritas: %v", err)
		return nil, err
	}

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Error al enviar los datos al servidor: %v", err)
		return nil, err
	}

	// Configura un tiempo de espera para la respuesta
	conn.SetReadDeadline(time.Now().Add(600 * time.Second))

	// Lee la respuesta del servidor como JSON
	var recommendedIDs []int
	decoder := json.NewDecoder(conn) // Usamos json.NewDecoder para leer directamente de la conexión
	err = decoder.Decode(&recommendedIDs)
	if err != nil {
		log.Printf("Error al decodificar la respuesta: %v", err)
		return nil, err
	}

	return recommendedIDs, nil
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
