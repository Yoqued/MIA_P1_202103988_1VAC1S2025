package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

func ParseFind(tokens []string) ([]string, error) {
	var path, name string
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
			path = value
		case "-name":
			name = value
		default:
			return tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	if path == "" {
		return tokens[contador:], errors.New("faltan parámetros requeridos: -path")
	}
	if name == "" {
		return tokens[contador:], errors.New("faltan parámetros requeridos: -name")
	}

	var treeFile string

	startPart := int64(binary.LittleEndian.Uint32(util.UserOnline.Particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(util.UserOnline.Path, startPart)
	parentsDir, _ := util.GetParentDirectories(path)

	fmt.Println("leyendo find")
	err := superblock.FindFile(util.UserOnline.Path, 0, parentsDir, name, &treeFile, 0)
	if err != nil {
		fmt.Println("ERROR", err)
		return tokens[contador:], err
	}

	util.Respuestas = append(util.Respuestas, "Arbol Actual")
	util.Respuestas = append(util.Respuestas, path)
	util.Respuestas = append(util.Respuestas, treeFile)

	return tokens[contador:], nil // Devuelve el comando MKDISK creado
}
