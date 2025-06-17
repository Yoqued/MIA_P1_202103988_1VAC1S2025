package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	objs "MIA_P1_202103988_1VAC1S2025/Objetos"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

type MKDIR struct {
	path string // Path del directorio
	p    bool   // Opción -p (crea directorios padres si no existen)
}

func ParseMkdir(tokens []string) ([]string, error) {
	cmd := &MKDIR{}
	var contador int
	for cont, token := range tokens {

		if token == "<newline>" {
			contador += 1
			break
		}

		if token == "-p" {
			cmd.p = true
			contador += 1
			continue
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
			cmd.path = value
		default:
			return tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	if cmd.path == "" {
		return tokens[contador:], errors.New("faltan parámetros requeridos: -path")
	}

	// Obtener particion montada
	fmt.Println(util.UserOnline)
	mbr, _ := objs.ReadMbr(util.UserOnline.Path)
	particion := mbr.BuscarIdParticion(util.UserOnline.Id)
	startPart := int64(binary.LittleEndian.Uint32(particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(util.UserOnline.Path, startPart)

	err := createDirectory(cmd.path, superblock, util.UserOnline.Path, particion, cmd.p)
	if err != nil {
		return tokens[contador:], err
	}

	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Carpeta creada en la ruta %s", cmd.path))

	return tokens[contador:], nil
}

func createDirectory(dirPath string, sb *estructuras.SUPERBLOCK, partitionPath string, particion *objs.PARTICION, createParents bool) error {
	fmt.Println("\nCreando directorio:", dirPath)

	parentDirs, destDir := util.GetParentDirectories(dirPath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// Crear el directorio segun el path proporcionado
	fmt.Println("VA ENTRAR")
	err := sb.CreateFolder(partitionPath, parentDirs, destDir, createParents)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	// Serializar el superbloque
	err = sb.WriteToFile(partitionPath, int(binary.LittleEndian.Uint32(particion.Start[:])))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}
