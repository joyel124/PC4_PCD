package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
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
	for userID, ratings := range data.Ratings {
		if userID != targetUser {
			sim := cosineSimilarity(data.Ratings[targetUser], ratings)
			if sim > maxSim {
				maxSim = sim
				similarUser = userID
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
		// Recomienda solo las películas que el usuario objetivo aún no ha calificado
		if _, rated := data.Ratings[targetUser][movieID]; !rated && rating >= 4.0 { // Recomendamos calificaciones >= 4
			recommendations = append(recommendations, movieID)
		}
	}
	return recommendations
}

func main() {
	// Cargar datos de ratings desde el archivo `movies_data.csv`
	data, err := loadNetflixData("movies_data.csv")
	if err != nil {
		log.Fatalf("Error al cargar datos: %v", err)
	}

	targetUser := 1488844 // Ejemplo de usuario objetivo
	recommendations := generateRecommendations(data, targetUser)

	fmt.Printf("Recomendaciones para el usuario %d:\n", targetUser)
	for _, movieID := range recommendations {
		fmt.Printf("Película ID: %d\n", movieID)
	}
}
