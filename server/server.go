package main

import (
	"encoding/gob"
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

// Función que maneja la conexión con el nodo cliente
func handleNodeConnection(conn net.Conn, targetUserID int) {
	defer conn.Close()

	// Timeout de 10 minutos en la conexión
	conn.SetDeadline(time.Now().Add(600 * time.Second))

	// Enviar el ID del usuario objetivo al nodo cliente
	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(targetUserID); err != nil {
		fmt.Println("Error al enviar ID del usuario al nodo:", err)
		wg.Done()
		return
	}
	fmt.Printf("ID del usuario objetivo %d enviado al nodo.\n", targetUserID)

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

func main() {
	// Definir el ID del usuario objetivo (por ejemplo, 1488844)
	targetUserID := 1488844

	// Iniciar servidor
	listener, err := net.Listen("tcp", "172.20.0.5:9002")
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor escuchando en el puerto 9002")

	// Número de nodos que esperamos recibir
	numNodes := 3 // Definir la cantidad de nodos esperada
	wg.Add(numNodes)

	// Escuchar por conexiones entrantes
	for i := 0; i < numNodes; i++ {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err)
			continue
		}
		fmt.Println("Nodo conectado")

		// Manejar cada conexión en una goroutine
		go handleNodeConnection(conn, targetUserID)
	}

	// Esperar a que todos los nodos terminen de enviar recomendaciones
	wg.Wait()

	// Mostrar las recomendaciones finales basadas en frecuencia
	showFinalRecommendations()
}

// Mostrar las recomendaciones finales después de recibir de todos los nodos
func showFinalRecommendations() {
	fmt.Println("Recomendaciones finales para el usuario objetivo:")

	// Ordenar y mostrar las películas más recomendadas
	for movieID, count := range recommendationsCount {
		fmt.Printf("Película ID: %d, Recomendaciones: %d\n", movieID, count)
	}
}
