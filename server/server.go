package main

import (
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// Mapa para almacenar las recomendaciones recibidas de cada nodo
var recommendationsCount = make(map[int]int)
var mu sync.Mutex
var wg sync.WaitGroup
var ratingData RatingData
var err error

// Lista de IPs de los nodos cliente en la red
var nodeIPs = []string{
	"172.20.0.2:9002", // IP y puerto del nodo 1
	"172.20.0.3:9002", // IP y puerto del nodo 2
	"172.20.0.4:9002", // IP y puerto del nodo 3
}

var nodeDatasets = []string{
	"/var/my-data/dataset_1.csv",
	"/var/my-data/dataset_1.csv",
	"/var/my-data/dataset_1.csv",
}

// Estructura para almacenar la matriz de calificaciones
type RatingData struct {
	Ratings map[int]map[int]float64
}

// Cargar datos de calificaciones
func loadNetflixData(filename string) (RatingData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return RatingData{}, err
	}
	defer file.Close()

	data := RatingData{Ratings: make(map[int]map[int]float64)}
	reader := csv.NewReader(file)

	// Leer encabezado
	_, err = reader.Read()
	if err != nil {
		return data, err
	}

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		movieID, _ := strconv.Atoi(record[0])
		customerID, _ := strconv.Atoi(record[1])
		rating, _ := strconv.ParseFloat(record[2], 64)

		if data.Ratings[customerID] == nil {
			data.Ratings[customerID] = make(map[int]float64)
		}
		data.Ratings[customerID][movieID] = rating
	}

	return data, nil
}

// Función que maneja la conexión con el nodo cliente
func handleNodeConnection(conn net.Conn, favoriteMovieIDs []int, ratingData RatingData, nodeIndex int) {
	defer conn.Close()

	// Timeout de 10 minutos en la conexión
	// conn.SetDeadline(time.Now().Add(600 * time.Second))

	// Crear un paquete que incluya las películas favoritas y el dataset
	payload := struct {
		FavoriteMovieIDs []int
		RatingData       RatingData
	}{
		FavoriteMovieIDs: favoriteMovieIDs,
		RatingData:       ratingData,
	}

	// Enviar el paquete al nodo cliente
	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(payload); err != nil {
		fmt.Println("Error al enviar datos al nodo:", err)
		wg.Done()
		return
	}

	fmt.Printf("Datos enviados al nodo %d: Películas favoritas: %v\n", nodeIndex+1, favoriteMovieIDs)

	// Recibir recomendaciones del nodo
	var recommendations []int
	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&recommendations); err != nil {
		fmt.Println("Error al recibir recomendaciones del nodo:", err)
		wg.Done()
		return
	}

	fmt.Printf("Recomendaciones recibidas del nodo %d: %v\n", nodeIndex+1, recommendations)

	// Almacenar las recomendaciones en el mapa compartido
	mu.Lock()
	for _, movieID := range recommendations {
		recommendationsCount[movieID]++
	}
	mu.Unlock()

	wg.Done()
}

// Función para verificar si un nodo está disponible
func checkNodeHealth(nodeIP string) bool {
	conn, err := net.DialTimeout("tcp", nodeIP, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Función para redirigir la tarea a otro nodo disponible
func handleReassignment(favoriteMovieIDs []int, ratingData RatingData) {
	for _, nodeIP := range nodeIPs {
		if checkNodeHealth(nodeIP) {
			conn, err := net.Dial("tcp", nodeIP)
			if err == nil {
				// Si el nodo está disponible, enviar los datos
				wg.Add(1)
				go handleNodeConnection(conn, favoriteMovieIDs, ratingData, 0) // El índice de nodo es irrelevante aquí
				return
			}
		}
	}
	fmt.Println("No hay nodos disponibles para reasignar la tarea.")
	wg.Done()
}

// Función que maneja la conexión con la API
func handleAPIConnection(conn net.Conn) {
	defer conn.Close()

	// Configura el timeout para la conexión
	//  conn.SetDeadline(time.Now().Add(600 * time.Second))

	// Recibir los IDs de películas favoritas desde la API
	var favoriteMovieIDs []int
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&favoriteMovieIDs); err != nil {
		fmt.Println("Error al decodificar IDs de películas favoritas desde la API:", err)
		return
	}

	fmt.Println("Películas favoritas recibidas desde la API:", favoriteMovieIDs)

	// Iniciar la conexión con los nodos clientes
	wg.Add(len(nodeIPs))
	for i, nodeIP := range nodeIPs {
		// Cargar el dataset correspondiente
		// ratingData, err := loadNetflixData(nodeDatasets[i])
		/*if err != nil {
			fmt.Printf("Error al cargar dataset %s: %v\n", nodeDatasets[i], err)
			wg.Done()
			continue
		}*/

		// Conectar a cada nodo según su IP en la bitácora
		conn, err := net.Dial("tcp", nodeIP)
		if err != nil {
			fmt.Printf("Error al conectar con el nodo %s: %v\n", nodeIP, err)
			handleReassignment(favoriteMovieIDs, ratingData)
			wg.Done()
			continue
		}

		go handleNodeConnection(conn, favoriteMovieIDs, ratingData, i)
	}

	// Esperar a que todos los nodos terminen de enviar recomendaciones
	wg.Wait()

	fmt.Println("Todas las recomendaciones han sido recibidas.")

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
	// Cargar los datos
	fmt.Println("Cargando datos...")
	ratingData, err = loadNetflixData(nodeDatasets[0])
	if err != nil {
		fmt.Printf("Error al cargar dataset %s: %v\n", nodeDatasets[0], err)
		os.Exit(1)
	}
	fmt.Println("Datos cargados exitosamente.")
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
