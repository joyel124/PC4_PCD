package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Content string `json:"content"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	broadcast = make(chan Message)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	mu sync.Mutex
)

// Maneja conexiones WebSocket
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error al actualizar a WebSocket: %v\n", err)
		return
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			mu.Lock()
			delete(clients, ws)
			mu.Unlock()
			break
		}
	}
}

// Maneja mensajes de la API REST
func handleAPI(w http.ResponseWriter, r *http.Request) {
	var msg Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, "Error al decodificar el mensaje", http.StatusBadRequest)
		return
	}

	broadcast <- msg
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(msg)
}

// Envía mensajes a todos los clientes WebSocket conectados
func handleMessages() {
	for {
		msg := <-broadcast
		mu.Lock()
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
	// Rutas
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/api", handleAPI)

	// Ejecuta el manejador de mensajes en una goroutine
	go handleMessages()

	// Inicia el servidor
	fmt.Println("Servidor iniciado en :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error en el servidor: ", err)
	}
}
