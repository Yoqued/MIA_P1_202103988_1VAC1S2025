package estructuras

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateBinaryFile crea un archivo binario con el tamaño y la unidad especificados
func CreateBinaryFile(path string, sizeInBytes int) error {

	err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		fmt.Println("Error creating directories:", err)
		return err
	}

	// Crea el archivo en la ruta especificada
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close() // Asegura que el archivo se cierre al finalizar

	// Escribe los bytes en el archivo
	return writeToFile(file, sizeInBytes)
}

// convertToBytes convierte el tamaño y la unidad a bytes
func ConvertToBytes(size int, unit string) (int, error) {
	switch unit {
	case "B":
		return size, nil
	case "K":
		return size * 1024, nil // Convierte kilobytes a bytes
	case "M":
		return size * 1024 * 1024, nil // Convierte megabytes a bytes
	default:
		return 0, fmt.Errorf("unidad desconocida: %s", unit)
	}
}

// writeToFile escribe los bytes en el archivo
func writeToFile(file *os.File, sizeInBytes int) error {
	buffer := make([]byte, 1024*1024) // Crea un buffer de 1 MB
	for sizeInBytes > 0 {
		writeSize := len(buffer)
		if sizeInBytes < writeSize {
			writeSize = sizeInBytes // Ajusta el tamaño de escritura si es menor que el buffer
		}
		if _, err := file.Write(buffer[:writeSize]); err != nil {
			return err // Devuelve un error si la escritura falla
		}
		sizeInBytes -= writeSize // Resta el tamaño escrito del tamaño total
	}
	fmt.Println("Archivo creado con éxito!")
	return nil
}
