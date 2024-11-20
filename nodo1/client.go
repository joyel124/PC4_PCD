package main

import (
	"encoding/gob"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
)

// Estructura para almacenar la matriz de calificaciones
type RatingData struct {
	Ratings map[int]map[int]float64
}

// Calcular similitud de cosenos entre dos películas
func calculateCosineSimilarity(movie1, movie2 map[int]float64) float64 {
	var dotProduct, normA, normB float64

	for userID, rating1 := range movie1 {
		if rating2, exists := movie2[userID]; exists {
			dotProduct += rating1 * rating2
		}
		normA += rating1 * rating1
	}
	for _, rating2 := range movie2 {
		normB += rating2 * rating2
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// Crear vectores de películas desde RatingData
func buildMovieVectors(data RatingData) map[int]map[int]float64 {
	movieVectors := make(map[int]map[int]float64)

	// Construir las representaciones de películas por usuario
	for userID, movies := range data.Ratings {
		for movieID, rating := range movies {
			if movieVectors[movieID] == nil {
				movieVectors[movieID] = make(map[int]float64)
			}
			movieVectors[movieID][userID] = rating
		}
	}

	return movieVectors
}

// Buscar películas similares a las favoritas
func findSimilarMovies(favoriteMovieIDs []int, data RatingData) []int {
	movieRatings := buildMovieVectors(data)
	similarities := make(map[int]float64)

	// Recorremos las películas favoritas
	for _, favID := range favoriteMovieIDs {
		favVector, exists := movieRatings[favID]
		if !exists {
			fmt.Printf("La película %d no está en los datos.\n", favID)
			continue
		}

		// Recorremos todas las películas y calculamos similitudes
		for movieID, vector := range movieRatings {
			if movieID != favID {
				// Calculamos la similitud entre la película favorita y otras
				similarity := calculateCosineSimilarity(favVector, vector)
				// Acumulamos la similitud
				similarities[movieID] += similarity
				// Mostrar la similitud en consola (opcional)
				// fmt.Printf("Similitud entre %d y %d: %f\n", favID, movieID, similarity)
			}
		}
	}

	// Ordenar las películas por similitud y devolver las más relevantes
	return sortMoviesByScore(similarities, 5)
}

// Ordenar películas por puntuación
func sortMoviesByScore(scores map[int]float64, limit int) []int {
	type movieScore struct {
		movieID int
		score   float64
	}

	// Crear una lista de las películas y sus puntuaciones
	var movieList []movieScore
	for movieID, score := range scores {
		movieList = append(movieList, movieScore{movieID, score})
	}

	// Ordenar por puntuación (de mayor a menor)
	sort.Slice(movieList, func(i, j int) bool {
		return movieList[i].score > movieList[j].score
	})

	// Recoger las mejores recomendaciones hasta el límite
	var sortedMovieIDs []int
	for i, movie := range movieList {
		if i >= limit {
			break
		}
		sortedMovieIDs = append(sortedMovieIDs, movie.movieID)
	}

	return sortedMovieIDs
}

// Función para enviar el resultado procesado al servidor
func sendResult(conn net.Conn, result []int) error {
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(result)
}

// Función para manejar la conexión con el servidor
func handleServerConnection(conn net.Conn) {
	fmt.Println("Conexión establecida con el servidor")

	// Decodificar el paquete recibido desde el servidor
	var payload struct {
		FavoriteMovieIDs []int
		RatingData       RatingData
	}

	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&payload); err != nil {
		fmt.Println("Error al recibir datos del servidor:", err)
		conn.Close()
		return
	}

	fmt.Printf("Películas favoritas recibidas: %v\n", payload.FavoriteMovieIDs)
	fmt.Println("Datos de calificación recibidos exitosamente.")

	// Generar recomendaciones para las películas favoritas
	recommendations := findSimilarMovies(payload.FavoriteMovieIDs, payload.RatingData)
	// Recomendaciones de ejemplo:
	// recommendations := []int{1, 2, 3, 4, 5}
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
	listener, err := net.Listen("tcp", "172.20.0.2:9002")
	if err != nil {
		fmt.Println("Error al iniciar el cliente:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Esperando conexiones entrantes en el puerto 172.20.0.2:9002...")

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
