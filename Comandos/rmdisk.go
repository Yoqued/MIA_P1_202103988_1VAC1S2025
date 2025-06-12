package comandos

import (
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Función para parsear el comando RMDISK
func ParseRmdisk(tokens []string) ([]string, error) {
	var path string
	var contador int
	for cont, token := range tokens {

		if token == "<newline>" {
			contador += 1
			break
		}

		// Divide cada token en clave y valor usando "=" como delimitador
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			return tokens[cont:], fmt.Errorf("formato de parámetro inválido: %s", token)
		}
		key, value := strings.ToLower(parts[0]), parts[1]

		// Key representa las palabras claves de cada atributo
		switch key {
		case "-path":
			// Verificar si el archivo existe
			if _, err := os.Stat(value); os.IsNotExist(err) {
				return tokens[cont:], fmt.Errorf("la ruta indicada no existe")
			}
			path = value
		default:
			return tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	// Verificar si el parámetro -path se proporcionó
	if path == "" {
		return tokens[contador:], errors.New("faltan parámetros requeridos: -path")
	}

	// Eliminar el archivo
	err := os.Remove(path)
	if err != nil {
		return tokens[contador:], fmt.Errorf("error al eliminar el archivo: %v", err)
	}

	fmt.Printf("Archivo en la ruta '%s' eliminado con éxito.\n", path)
	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Disco %s Removido con exito!", path))

	return tokens[contador:], nil
}
