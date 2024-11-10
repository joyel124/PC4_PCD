package main

import (
	"encoding/csv"
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// Estructura que representa un rating en el dataset de recomendaciones
type Rating struct {
	UserID  int
	MovieID int
	Rating  float64
}

// Fragmento de datos que se enviará a cada nodo
type DataFragment struct {
	FragmentID int
	Data       []Rating
}

// Perfil de usuario
type UserProfile struct {
	UserID         int
	PreferredItems []int
}

// Variables globales
var dataset []Rating // Dataset de ratings cargado desde el CSV
var pendingFragments = make(map[int]DataFragment)
var mu sync.Mutex // Mutex para proteger el acceso a los fragmentos pendientes
var nextFragmentID = 1

// Función para cargar el dataset de ratings desde el CSV
func loadDatasetFromCSV(filePath string) ([]Rating, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data []Rating
	reader := csv.NewReader(file)
	_, _ = reader.Read() // Omitimos la primera línea (encabezado)

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		movieID, _ := strconv.Atoi(record[0])          // Movie_Id
		userID, _ := strconv.Atoi(record[1])           // Cust_Id
		rating, _ := strconv.ParseFloat(record[3], 64) // Rating como float64

		data = append(data, Rating{
			UserID:  userID,
			MovieID: movieID,
			Rating:  rating,
		})
	}
	return data, nil
}

// Función que divide el dataset en fragmentos
func createFragments(data []Rating, fragmentSize int) []DataFragment {
	var fragments []DataFragment
	for i := 0; i < len(data); i += fragmentSize {
		end := i + fragmentSize
		if end > len(data) {
			end = len(data)
		}
		fragment := DataFragment{
			FragmentID: nextFragmentID,
			Data:       data[i:end],
		}
		fragments = append(fragments, fragment)
		nextFragmentID++
	}
	fmt.Printf("Dataset dividido en %d fragmentos\n", len(fragments))
	return fragments
}

// Enviar fragmento al nodo
func sendFragment(conn net.Conn, fragment DataFragment) error {
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(fragment)
}

// Recibir resultado del nodo
func receiveResults(conn net.Conn, fragmentID int) (float64, error) {
	var avgRating float64
	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(&avgRating)
	if err == nil {
		mu.Lock()
		delete(pendingFragments, fragmentID)
		mu.Unlock()
	}
	return avgRating, err
}

// Función para reasignar un fragmento a otro nodo en caso de error
func reassignTaskToAnotherNode(fragment DataFragment) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("Reasignando fragmento %d a otro nodo\n", fragment.FragmentID)
	pendingFragments[fragment.FragmentID] = fragment
}

// Función que maneja la conexión con el nodo
func handleNodeConnection(conn net.Conn) {
	defer conn.Close()

	// Timeout de 10 segundos en la conexión
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	for {
		// Obtener un fragmento pendiente para enviar
		mu.Lock()
		var fragmentToAssign DataFragment
		for _, fragment := range pendingFragments {
			fragmentToAssign = fragment
			break
		}
		mu.Unlock()

		// Si no hay fragmentos pendientes, enviar mensaje de fin de proceso
		if fragmentToAssign.FragmentID == 0 && len(pendingFragments) == 0 {
			// Enviar el mensaje de fin de proceso
			fmt.Println("No hay más fragmentos para enviar. Enviando señal de fin.")
			endSignal := "FIN"
			encoder := gob.NewEncoder(conn)
			if err := encoder.Encode(&endSignal); err != nil {
				fmt.Println("Error al enviar señal de fin:", err)
			}
			return // Salir después de enviar el mensaje de fin
		}

		// Enviar el fragmento y recibir resultados
		fmt.Printf("Enviando fragmento %d al nodo\n", fragmentToAssign.FragmentID)
		// Imprimir el fragmento para depuración
		fmt.Printf("Fragmento %d: %v\n", fragmentToAssign.FragmentID, fragmentToAssign.Data)
		if err := sendFragment(conn, fragmentToAssign); err != nil {
			fmt.Println("Error al enviar fragmento:", err)
			reassignTaskToAnotherNode(fragmentToAssign)
			return
		}

		avgRating, err := receiveResults(conn, fragmentToAssign.FragmentID)
		if err != nil {
			fmt.Println("Error al recibir resultado del nodo:", err)
			reassignTaskToAnotherNode(fragmentToAssign)
			return
		}

		fmt.Printf("Promedio de rating recibido del nodo para fragmento %d: %f\n", fragmentToAssign.FragmentID, avgRating)
	}
}

func main() {
	// Iniciar servidor
	var err error
	listener, err := net.Listen("tcp", "172.20.0.5:9002")
	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Servidor escuchando en el puerto 9002")

	// Cargar dataset de ratings y títulos de películas
	dataset, err = loadDatasetFromCSV("/var/my-data/dataset.csv")
	if err != nil {
		fmt.Println("Error al cargar dataset:", err)
		return
	}

	fmt.Println("Dataset cargado correctamente")

	// Crear fragmentos del dataset y guardarlos en pendientes
	fragments := createFragments(dataset, 10) // Dividimos en fragmentos de tamaño 10

	// Limitar el número de fragmentos para pruebas (ej., solo los primeros 5)
	testFragments := fragments
	if len(fragments) > 5 {
		testFragments = fragments[:5] // Solo los primeros 5 fragmentos
	}

	// Guardar los fragmentos en el mapa de pendientes
	mu.Lock()
	for _, fragment := range testFragments {
		pendingFragments[fragment.FragmentID] = fragment
	}
	mu.Unlock()

	// Escuchar por conexiones entrantes
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error al aceptar conexión:", err)
			continue
		}
		fmt.Println("Nodo conectado")

		// Manejar cada conexión en una goroutine
		go handleNodeConnection(conn)
	}
}
