package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
)

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

// Calcular la similitud de coseno entre dos películas
func cosineSimilarityBetweenMovies(movie1, movie2 map[int]float64) float64 {
	var dotProduct, normA, normB float64
	for userID, rating1 := range movie1 {
		if rating2, exists := movie2[userID]; exists {
			dotProduct += rating1 * rating2
			normA += rating1 * rating1
			normB += rating2 * rating2
		}
	}
	if normA == 0 || normB == 0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Generar recomendaciones basadas en las películas favoritas
func generateMovieRecommendations(data RatingData, favoriteMovies []int) []int {
	similarMoviesScores := make(map[int]float64)

	// Comparar cada película favorita con todas las demás
	for _, movieID := range favoriteMovies {
		for otherMovieID, otherMovieRatings := range data.Ratings {
			if otherMovieID == movieID {
				continue // No compararse con la misma película
			}

			// Obtener la similitud entre las dos películas
			similarity := cosineSimilarityBetweenMovies(data.Ratings[movieID], otherMovieRatings)

			// Almacenar las películas similares y sus puntajes
			similarMoviesScores[otherMovieID] += similarity
		}
	}

	// Generar un array con las películas más recomendadas
	recommendedMovies := []int{}
	for movieID := range similarMoviesScores {
		recommendedMovies = append(recommendedMovies, movieID)
	}

	return recommendedMovies
}

// Función para enviar el resultado procesado al servidor
func sendResult(conn net.Conn, result []int) error {
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(result)
}

// Función para manejar la conexión con el servidor
func handleServerConnection(conn net.Conn) {
	fmt.Println("Conexión establecida con el servidor")

	// Decodificar el array de películas favoritas recibido desde el servidor
	var favoriteMovies []int
	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&favoriteMovies); err != nil {
		fmt.Println("Error al recibir las películas favoritas:", err)
		conn.Close()
		return
	}
	fmt.Printf("Películas favoritas recibidas: %v\n", favoriteMovies)

	fmt.Println("Cargando dataset...")

	// Cargar dataset local
	_, err := loadNetflixData("/var/my-data/dataset.csv")
	if err != nil {
		fmt.Println("Error al cargar dataset:", err)
		return
	}

	fmt.Println("Dataset cargado exitosamente.")

	// Generar recomendaciones para las películas favoritas
	// recommendations := generateMovieRecommendations(dataset, favoriteMovies)
	// Recomendaciones de ejemplo:
	recommendations := []int{11, 12, 13, 14, 15}
	fmt.Printf("Recomendaciones generadas para las películas favoritas: %v\n", recommendations)

	// Enviar recomendaciones al servidor
	if err := sendResult(conn, recommendations); err != nil {
		fmt.Println("Error al enviar recomendaciones:", err)
	} else {
		fmt.Println("Recomendaciones enviadas al servidor exitosamente.")
	}
}

func main() {
	// Iniciar el servidor y escuchar por conexiones entrantes
	listener, err := net.Listen("tcp", "172.20.0.4:9002")
	if err != nil {
		fmt.Println("Error al iniciar el cliente:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Esperando conexiones entrantes en el puerto 172.20.0.4:9002...")

	// Escuchar por conexiones entrantes desde el servidor
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err)
			continue
		}

		// Procesar la conexión en una goroutine
		handleServerConnection(conn)
	}
}
