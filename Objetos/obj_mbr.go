package objs

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type MBR struct {
	Size        [4]byte
	Fecha       [4]byte
	Fit         [1]byte
	Signature   [4]byte
	Particiones [4]PARTICION
}

// WriteToFile escribe la estructura en el primer bloque libre de un archivo binario
func (mbr *MBR) WriteToFile(filepath string) error {
	// Abre el archivo binario en modo lectura/escritura, lo crea si no existe
	file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close() // Asegura que el archivo se cierre al finalizar

	// Mueve el puntero del archivo a la posición del bloque libre encontrado
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("error al mover el puntero del archivo: %v", err)
	}

	// Escribe la estructura Student en el archivo usando codificación Little Endian
	err = binary.Write(file, binary.LittleEndian, mbr)
	if err != nil {
		return fmt.Errorf("error al escribir la estructura Student: %v", err)
	}

	return nil // Retorna nil si no hubo errores
}

func ReadMbr(path string) (*MBR, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	mbr := &MBR{}
	if err := binary.Read(file, binary.LittleEndian, mbr); err != nil {
		return nil, err
	}

	return mbr, nil
}

func (m *MBR) AgregarParticion(part PARTICION) error {
	for i, particion := range m.Particiones {
		fmt.Println("for")
		if IsEmptyByte(particion.Name[:]) {
			copy(part.Correlative[:], strconv.Itoa(i+1))
			fmt.Println("Particion Agregada")
			m.Particiones[i] = part
			return nil
		}
	}
	return fmt.Errorf("no hay particiones vacías disponibles")
}

func (m *MBR) BuscarParticion(name string) *PARTICION {
	for i, particion := range m.Particiones {
		// Convertir el array de bytes a string y eliminar los ceros nulos
		partitionName := strings.TrimRight(string(particion.Name[:]), "\x00")
		if partitionName == name {
			return &m.Particiones[i]
		}
	}
	return nil
}

func (m *MBR) BuscarIdParticion(id string) *PARTICION {
	for i, particion := range m.Particiones {
		// Convertir el array de bytes a string y eliminar los ceros nulos
		partitionID := strings.TrimRight(string(particion.Id[:]), "\x00")
		if partitionID == id {
			return &m.Particiones[i]
		}
	}
	return nil
}

func (m *MBR) PosicionParticion(name string, path string) int {
	for i, particion := range m.Particiones {
		// Convertir el array de bytes a string y eliminar los ceros nulos
		partitionName := strings.TrimRight(string(particion.Name[:]), "\x00")
		if string(particion.Tipo[:]) == "E" {
			if partitionName == name || FindLogica(path, int(binary.LittleEndian.Uint32(particion.Start[:])), name) {
				return 4
			}
		}
		if partitionName == name {
			return i
		}
	}
	return -1
}

func (m *MBR) VerificarExtendida() bool {
	for _, particion := range m.Particiones {
		if string(particion.Tipo[:]) == "E" {
			return true
		}
	}
	return false
}

func (m *MBR) StartExtendida() int {
	for _, particion := range m.Particiones {
		if string(particion.Tipo[:]) == "E" {
			start := int(binary.LittleEndian.Uint32(particion.Start[:]))
			fmt.Println("startExtendida:", start)
			return start
		}
	}
	return 0
}

func findFreeBlock(file *os.File, blockSize int) (int64, error) {
	buffer := make([]byte, blockSize) // Crea un buffer del tamaño del bloque
	var offset int64

	for {
		// Lee un bloque del archivo en la posición actual de offset
		_, err := file.ReadAt(buffer, offset)
		if err != nil {
			break // Si hay un error (EOF), sale del bucle
		}

		isFree := true
		// Verifica si el bloque está libre (todos los bytes son 0)
		for _, b := range buffer {
			if b != 0 {
				isFree = false
				break
			}
		}

		if isFree {
			return offset, nil // Retorna el offset del bloque libre encontrado
		}

		offset += int64(blockSize) // Incrementa el offset para leer el siguiente bloque
	}

	return offset, nil // Retorna el offset del final del archivo si no se encontró un bloque libre
}
