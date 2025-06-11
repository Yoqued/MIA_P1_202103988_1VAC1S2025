package main

import (
	analizar "MIA_P1_202103988_1VAC1S2025/Analizador" // Importación del paquete analizar
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Solicitar el nombre base del archivo (sin extensión)
	fmt.Print("Ingrese el nombre del archivo (sin extensión): ")
	var nombreBase string
	fmt.Scanln(&nombreBase)

	// Buscar archivos que coincidan
	archivosEncontrados, err := buscarArchivosPorNombre(nombreBase)
	if err != nil {
		log.Fatalf("Error al buscar archivos: %v", err)
	}

	if len(archivosEncontrados) == 0 {
		log.Fatalf("No se encontraron archivos con el nombre '%s'", nombreBase)
	}

	// Mostrar resultados
	fmt.Printf("\nArchivos encontrados:\n")
	for i, archivo := range archivosEncontrados {
		fmt.Printf("%d. %s\n", i+1, archivo)
	}

	// Leer el primer archivo encontrado
	contenido, err := ioutil.ReadFile(archivosEncontrados[0])
	if err != nil {
		log.Fatalf("Error al leer el archivo: %v", err)
	}

	fmt.Printf("\n=== Contenido de '%s' ===\n\n", archivosEncontrados[0])
	fmt.Println(string(contenido))

	// Llamar a la función AnalizarTexto del paquete analizar
	resultado, errores := analizar.AnalizarTexto(string(contenido))
	if len(errores) > 0 {
		fmt.Println("\n=== Errores encontrados ===")
		for _, err := range errores {
			fmt.Printf("- %v\n", err)
		}
	}

	fmt.Println("\n=== Resultado del análisis ===")
	fmt.Println(resultado)
}

// Busca archivos cuyo nombre base coincida (sin considerar extensión)
func buscarArchivosPorNombre(nombreBase string) ([]string, error) {
	var coincidencias []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Obtener el nombre del archivo sin extensión
			base := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

			// Comparar sin importar mayúsculas/minúsculas
			if strings.EqualFold(base, nombreBase) {
				coincidencias = append(coincidencias, path)
			}
		}
		return nil
	})

	return coincidencias, err
}
