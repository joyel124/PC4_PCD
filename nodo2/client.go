package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"time"
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

// Calcular la similitud de coseno entre dos usuarios
func cosineSimilarity(user1, user2 map[int]float64) float64 {
	var dotProduct, normA, normB float64
	for movieID, rating1 := range user1 {
		if rating2, exists := user2[movieID]; exists {
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

// Encontrar el usuario más similar al usuario objetivo
func findMostSimilarUser(data RatingData, targetUser int) int {
	maxSim := 0.0
	similarUser := -1
	maxCommonRatings := 0

	for userID, ratings := range data.Ratings {
		if userID != targetUser {
			sim := cosineSimilarity(data.Ratings[targetUser], ratings)
			commonRatings := 0

			for movieID := range data.Ratings[targetUser] {
				if _, exists := ratings[movieID]; exists {
					commonRatings++
				}
			}

			if sim > maxSim || (sim == maxSim && commonRatings > maxCommonRatings) {
				maxSim = sim
				similarUser = userID
				maxCommonRatings = commonRatings
			}
		}
	}
	return similarUser
}

// Generar recomendaciones para el usuario objetivo basado en el usuario más similar
func generateRecommendations(data RatingData, targetUser int) []int {
	similarUser := findMostSimilarUser(data, targetUser)
	recommendations := []int{}

	for movieID, rating := range data.Ratings[similarUser] {
		if _, rated := data.Ratings[targetUser][movieID]; !rated && rating >= 3.0 {
			recommendations = append(recommendations, movieID)
		}
	}
	return recommendations
}

// Función para enviar el resultado procesado al servidor
func sendResult(conn net.Conn, result []int) error {
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(result)
}

// Conectar al servidor y recibir el ID del usuario
func connectToServerAndReceiveUserID() (net.Conn, int, error) {
	var conn net.Conn
	var err error

	for {
		conn, err = net.Dial("tcp", "172.20.0.5:9002")
		if err != nil {
			fmt.Println("Error al conectar con el servidor. Reintentando en 2 segundos...")
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Println("Conectado al servidor")

		// Decodificar el ID del usuario objetivo recibido desde el servidor
		var targetUserID int
		decoder := gob.NewDecoder(conn)
		if err := decoder.Decode(&targetUserID); err != nil {
			fmt.Println("Error al recibir ID del usuario:", err)
			conn.Close()
			return nil, 0, err
		}
		fmt.Printf("ID del usuario objetivo recibido: %d\n", targetUserID)
		return conn, targetUserID, nil
	}
}

func main() {
	// Conectar al servidor y recibir el ID del usuario
	conn, targetUserID, err := connectToServerAndReceiveUserID()
	if err != nil {
		fmt.Println("Error al conectarse al servidor:", err)
		return
	}
	defer conn.Close()

	// Cargar dataset local
	dataset, err := loadNetflixData("/var/my-data/dataset.csv")
	if err != nil {
		fmt.Println("Error al cargar dataset:", err)
		return
	}

	// Generar recomendaciones para el usuario recibido
	recommendations := generateRecommendations(dataset, targetUserID)
	fmt.Printf("Recomendaciones generadas para el usuario %d: %v\n", targetUserID, recommendations)

	// Enviar recomendaciones al servidor
	if err := sendResult(conn, recommendations); err != nil {
		fmt.Println("Error al enviar recomendaciones:", err)
	} else {
		fmt.Println("Recomendaciones enviadas al servidor exitosamente.")
	}
}
