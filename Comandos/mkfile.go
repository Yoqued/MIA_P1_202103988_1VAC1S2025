package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	objs "MIA_P1_202103988_1VAC1S2025/Objetos"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type MKFILE struct {
	Path string
	R    bool
	Size int32
	Cont string
}

func ParseMkfile(tokens []string) ([]string, error) {
	cmd := &MKFILE{}
	var contador int
	for cont, token := range tokens {

		if token == "<newline>" {
			contador += 1
			break
		}

		if token == "-r" {
			cmd.R = true
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
			cmd.Path = value
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size < 0 {
				return tokens[cont:], errors.New("El tamaño debe ser un número entero positivo")
			}
			cmd.Size = int32(size)
		case "-cont":
			cmd.Cont = value
		default:
			return tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	if cmd.Path == "" {
		return tokens[contador:], errors.New("faltan parámetros requeridos: -path")
	}
	if cmd.Size == 0 {
		cmd.Size = 0
	}
	if cmd.Cont == "" {
		cmd.Cont = ""
	} else {
		contenido, err := LeerContenidoArchivo(cmd.Cont)
		if err != nil {
			return tokens[contador:], err
		}
		cmd.Cont = contenido
	}

	if cmd.Cont == "" && cmd.Size == 0 {
		return tokens[contador:], errors.New("No se puede generar el archivo porque no hay contenido")
	}

	// Obtener particion montada
	mbr, _ := objs.ReadMbr(util.UserOnline.Path)
	particion := mbr.BuscarIdParticion(util.UserOnline.Id)
	startPart := int64(binary.LittleEndian.Uint32(particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(util.UserOnline.Path, startPart)

	if cmd.Cont == "" {
		cmd.Cont = generateContent(int(cmd.Size))
	}

	err := createFile(cmd.Path, int(cmd.Size), cmd.Cont, superblock, util.UserOnline.Path, particion, cmd.R)
	if err != nil {
		return tokens[contador:], err
	}

	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Archivo creado en la ruta %s", cmd.Path))

	return tokens[contador:], nil
}

func createFile(filePath string, size int, content string, sb *estructuras.SUPERBLOCK, particionPath string, particion *objs.PARTICION, r bool) error {
	fmt.Println("\nCreando archivo:", filePath)

	parentDirs, destDir := util.GetParentDirectories(filePath)
	fmt.Println("\nDirectorios padres:", parentDirs)
	fmt.Println("Directorio destino:", destDir)

	// Obtener contenido por chunks
	chunks := util.SplitStringChunks(content)
	fmt.Println("\nChunks del contenido:", chunks)

	// Crear el archivo
	err := sb.CreateFile(particionPath, parentDirs, destDir, size, chunks, r)
	if err != nil {
		return fmt.Errorf("error al crear el archivo: %w", err)
	}

	// Serializar el superbloque
	err = sb.WriteToFile(particionPath, int(binary.LittleEndian.Uint32(particion.Start[:])))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	return nil
}

func generateContent(size int) string {
	content := ""
	for len(content) < size {
		content += "0123456789"
	}
	return content[:size] // Recorta la cadena al tamaño exacto
}

func LeerContenidoArchivo(ruta string) (string, error) {
	// Verificar si el archivo existe
	if _, err := os.Stat(ruta); os.IsNotExist(err) {
		return "", fmt.Errorf("el archivo no existe: %s", ruta)
	}

	// Leer el archivo
	contenido, err := ioutil.ReadFile(ruta)
	if err != nil {
		return "", fmt.Errorf("error al leer el archivo: %v", err)
	}

	// Retornar el contenido como string
	return string(contenido), nil
}
