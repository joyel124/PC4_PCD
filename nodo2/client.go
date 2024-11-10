package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"strings"
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

// Función para recibir fragmentos desde el servidor
func receiveFragment(conn net.Conn) (DataFragment, error) {
	var fragment DataFragment
	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(&fragment)
	return fragment, err
}

// Función para enviar el resultado procesado al servidor
func sendResult(conn net.Conn, result float64) error {
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(result)
}

// Función para enviar mensaje de finalización al servidor
func sendCompletionMessage(conn net.Conn) error {
	completionMessage := "finished"
	encoder := gob.NewEncoder(conn)
	return encoder.Encode(completionMessage)
}

// Procesa un fragmento de datos, calculando el promedio de los ratings
func processFragment(fragment DataFragment) float64 {
	var sum float64
	for _, value := range fragment.Data {
		sum += value.Rating
	}
	if len(fragment.Data) == 0 {
		return 0 // Evita división por cero si el fragmento está vacío
	}
	return sum / float64(len(fragment.Data)) // Promedio de los datos
}

// Función para mostrar el progreso de procesamiento
func showProgress(current, total int) {
	progress := float64(current) / float64(total) * 100
	fmt.Printf("\rProgreso: [%-50s] %.2f%%", stringProgress(int(progress)), progress)
}

// Genera una cadena de progreso visual
func stringProgress(progress int) string {
	barLength := 50
	filled := (progress * barLength) / 100
	return fmt.Sprintf("%s%s", strings.Repeat("=", filled), strings.Repeat("-", barLength-filled))
}

// Función para conectar al servidor
func connectToServer() net.Conn {
	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", "172.20.0.5:9002") // Cambia la IP si es necesario
		if err != nil {
			fmt.Println("Error al conectar con el servidor. Reintentando en 2 segundos...")
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Println("Conectado al servidor")
		return conn
	}
}

func main() {
	var totalFragments int // Contador total estimado de fragmentos

	for {
		conn := connectToServer()
		fmt.Println("Esperando fragmentos del servidor...")

		fragmentCounter := 0 // Contador de fragmentos recibidos en esta conexión

		for {
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))

			// Recibe un fragmento del servidor
			fragment, err := receiveFragment(conn)
			if err != nil {
				if err == io.EOF {
					// Manejar cierre de la conexión si es por EOF
					fmt.Println("\nConexión cerrada por el servidor. Intentando reconectar...")
					break
				}
				// Imprimir el fragmento que causó el error
				fmt.Printf("\nFragmento que causó el error: %v\n", fragment)
				fmt.Println("\nError al recibir fragmento:", err)
				break
			}

			fragmentCounter++ // Aumenta el contador de fragmentos
			totalFragments++  // Mantiene un recuento total de fragmentos

			// Procesa el fragmento y muestra el progreso
			result := processFragment(fragment)
			showProgress(fragmentCounter, totalFragments)
			fmt.Printf("\nResultado procesado para fragmento %d: %f\n", fragment.FragmentID, result)

			// Enviar el resultado al servidor
			if err := sendResult(conn, result); err != nil {
				fmt.Println("Error al enviar resultado:", err)
				break
			}
		}

		// Enviar mensaje de finalización después de procesar todos los fragmentos
		fmt.Println("Todos los fragmentos procesados. Enviando mensaje de finalización al servidor...")
		if err := sendCompletionMessage(conn); err != nil {
			fmt.Println("Error al enviar mensaje de finalización:", err)
		} else {
			fmt.Println("Mensaje de finalización enviado exitosamente.")
		}

		// Esperar la señal de fin de proceso desde el servidor
		var endSignal string
		decoder := gob.NewDecoder(conn)
		if err := decoder.Decode(&endSignal); err != nil {
			if err == io.EOF {
				// Si se recibe EOF, la conexión probablemente se cerró por el servidor
				fmt.Println("\nServidor cerró la conexión correctamente.")
			} else {
				fmt.Println("Error al recibir señal de fin:", err)
			}
			break
		}

		if endSignal == "FIN" {
			fmt.Println("Servidor ha indicado que ya no hay más fragmentos para procesar.")
		}

		conn.Close() // Cerrar la conexión después de completar el proceso
	}
}
