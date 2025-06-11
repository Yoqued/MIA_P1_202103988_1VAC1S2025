package util

import (
	objs "MIA_P1_202103988_1VAC1S2025/Objetos"
	"bytes"
	"encoding/binary"
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

var Respuestas []string

type GRUPO struct {
	GNames     []string
	UNames     []string
	GroupCount int
	UserCount  int
}

type USER struct {
	User      string
	Pass      string
	Id        string
	Grp       string
	Particion *objs.PARTICION
	Path      string
	UId       int
	GId       int
}

var UserOnline USER

var GruposId = make(map[string]GRUPO)

func BuscarGId(userCadena string, Groups []string) int {
	var name string
	if strings.Contains(userCadena, ",") {
		elementos := strings.Split(userCadena, ",")
		name = elementos[2]
	} else {
		name = strings.TrimSpace(userCadena)
	}

	for i, nombre := range Groups {
		if nombre == name {
			return i + 1
		}
	}
	return -1
}

func UsuarioExiste(userCadena string, Names []string) bool {
	var name string
	if strings.Contains(userCadena, ",") {
		elementos := strings.Split(userCadena, ",")
		name = elementos[3]
	} else {
		name = strings.TrimSpace(userCadena)
	}

	for _, nombre := range Names {
		if nombre == name {
			return true
		}
	}
	return false
}

// Función para agregar un nuevo nombre al grupo en el mapa GruposId
func UpdateUser(groupId string, name string) {
	// Recuperas la estructura GRUPO actual del mapa
	grupo := GruposId[groupId]

	// Actualizar
	grupo.UNames = append(grupo.UNames, name)
	grupo.UserCount += 1
	GruposId[groupId] = grupo
}

// First devuelve el primer elemento de un slice
func First[T any](slice []T) (T, error) {
	if len(slice) == 0 {
		var zero T
		return zero, errors.New("el slice está vacío")
	}
	return slice[0], nil
}

// RemoveElement elimina un elemento de un slice en el índice dado
func RemoveElement[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice // Índice fuera de rango, devolver el slice original
	}
	return append(slice[:index], slice[index+1:]...)
}

// GetParentDirectories obtiene las carpetas padres y el directorio de destino
func GetParentDirectories(path string) ([]string, string) {
	// Normalizar el path
	path = filepath.Clean(path)

	// Dividir el path en sus componentes
	components := strings.Split(path, string(filepath.Separator))

	// Lista para almacenar las rutas de las carpetas padres
	var parentDirs []string

	// Construir las rutas de las carpetas padres, excluyendo la última carpeta
	for i := 1; i < len(components)-1; i++ {
		parentDirs = append(parentDirs, components[i])
	}

	// La última carpeta es la carpeta de destino
	destDir := components[len(components)-1]

	return parentDirs, destDir
}

// splitStringIntoChunks divide una cadena en partes de tamaño chunkSize y las almacena en una lista
func SplitStringChunks(s string) []string {
	var chunks []string
	for i := 0; i < len(s); i += 64 {
		end := i + 64
		if end > len(s) {
			end = len(s)
		}
		chunks = append(chunks, s[i:end])
	}
	return chunks
}

func GetMBR_Partitions(path string) (*objs.MBR, []objs.PARTICION_CONV, error) {
	mbr, err := objs.ReadMbr(path)
	if err != nil {
		return nil, nil, err
	}

	particions := ConvertPartition(mbr.Particiones[:])
	return mbr, particions, nil
}

func ConvertPartition(particiones []objs.PARTICION) []objs.PARTICION_CONV {
	partitions := []objs.PARTICION_CONV{}
	for _, particion := range particiones {
		partition := objs.PARTICION_CONV{}
		partition.Name = string(bytes.Trim(particion.Name[:], "\x00"))
		partition.Status = string(particion.Status[:])
		partition.Tipo = string(particion.Tipo[:])
		partition.Fit = string(particion.Fit[:])
		partition.Start = int(binary.LittleEndian.Uint32(particion.Start[:]))
		partition.Size = float32(binary.LittleEndian.Uint32(particion.Size[:]))
		partition.Correlative = string(bytes.Trim(particion.Correlative[:], "\x00"))
		partition.Id = string(particion.Id[:])
		partitions = append(partitions, partition)
	}
	return partitions
}

func ConvertToRegex(filtro string) (*regexp.Regexp, error) {
	// Reemplazar los caracteres especiales ? y * por su equivalente en regex
	// ? -> . (un solo carácter)
	// * -> .* (cero o más caracteres)
	regexStr := strings.ReplaceAll(filtro, "?", ".")
	regexStr = strings.ReplaceAll(regexStr, "*", ".*")

	// Compilar la expresión regular
	return regexp.Compile("^" + regexStr + "$")
}
