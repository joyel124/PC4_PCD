package main

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// Mapa para almacenar las recomendaciones recibidas de cada nodo
var recommendationsCount = make(map[int]int)
var mu sync.Mutex
var wg sync.WaitGroup

// Lista de IPs de los nodos cliente en la red
var nodeIPs = []string{
	"172.20.0.2:9002", // IP y puerto del nodo 1
	"172.20.0.3:9002", // IP y puerto del nodo 2
	"172.20.0.4:9002", // IP y puerto del nodo 3
}

// Función que maneja la conexión con el nodo cliente
func handleNodeConnection(conn net.Conn, favoriteMovieIDs []int) {
	defer conn.Close()

	// Timeout de 10 minutos en la conexión
	conn.SetDeadline(time.Now().Add(600 * time.Second))

	// Enviar el array de IDs de películas favoritas al nodo cliente
	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(favoriteMovieIDs); err != nil {
		fmt.Println("Error al enviar IDs de películas al nodo:", err)
		wg.Done()
		return
	}
	fmt.Printf("IDs de películas favoritas %v enviados al nodo.\n", favoriteMovieIDs)

	// Recibir recomendaciones del nodo
	var recommendations []int
	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&recommendations); err != nil {
		fmt.Println("Error al recibir recomendaciones del nodo:", err)
		wg.Done()
		return
	}

	// Almacenar las recomendaciones en el mapa compartido
	mu.Lock()
	for _, movieID := range recommendations {
		recommendationsCount[movieID]++
	}
	mu.Unlock()

	wg.Done()
}

// Función que maneja la conexión con la API
func handleAPIConnection(conn net.Conn) {
	defer conn.Close()

	// Configura el timeout para la conexión
	conn.SetDeadline(time.Now().Add(600 * time.Second))

	// Recibir los IDs de películas favoritas desde la API
	var favoriteMovieIDs []int
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&favoriteMovieIDs); err != nil {
		fmt.Println("Error al decodificar IDs de películas favoritas desde la API:", err)
		return
	}

	// Iniciar la recepción de recomendaciones de los nodos
	wg.Add(len(nodeIPs))
	for _, nodeIP := range nodeIPs {
		// Conectar a cada nodo según su IP en la bitácora
		conn, err := net.Dial("tcp", nodeIP)
		if err != nil {
			fmt.Printf("Error al conectar con el nodo %s: %v\n", nodeIP, err)
			wg.Done()
			continue
		}

		go handleNodeConnection(conn, favoriteMovieIDs)
	}

	// Esperar a que todos los nodos terminen de enviar recomendaciones
	wg.Wait()

	// Recopilar y enviar las recomendaciones al cliente API
	finalRecommendations := gatherFinalRecommendations()
	response, _ := json.Marshal(finalRecommendations)
	conn.Write(response)
}

// Reúne las recomendaciones basadas en frecuencia y las ordena
func gatherFinalRecommendations() []int {
	var sortedRecommendations []int
	mu.Lock()
	for movieID := range recommendationsCount {
		sortedRecommendations = append(sortedRecommendations, movieID)
	}
	mu.Unlock()
	return sortedRecommendations
}

func main() {
	// Iniciar servidor en el puerto 9002
	listener, err := net.Listen("tcp", "172.20.0.5:9002")
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor escuchando en el puerto 9002")

	// Escuchar por conexiones entrantes desde la API
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión de la API:", err)
			continue
		}

		// Manejar cada conexión de la API en una goroutine
		go handleAPIConnection(conn)
	}
}
