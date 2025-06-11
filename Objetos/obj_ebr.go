package objs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type EBR struct {
	Name   [16]byte // Nombre de la particion
	Status [1]byte  // Indica si esta montada o no
	Fit    [1]byte  // Tipo de ajuste (B, F, W)
	Start  [4]byte  // En que byte inicia la particion
	Size   [4]byte  // Contiene el tamaño de la particion
	Next   [4]byte  // Byte en el que esta el proximo ebr
}

// WriteToFile escribe la estructura del ebr segun el start
func (s *EBR) WriteToFile(path string, start int) error {
	// Abre el archivo binario en modo lectura/escritura, lo crea si no existe
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Mueve el puntero al byte correspondiente
	startoff := int64(start)
	_, err = file.Seek(startoff, 0)
	if err != nil {
		return fmt.Errorf("error al mover el puntero del archivo: %v", err)
	}

	// Escribe la estructura Ebr
	err = binary.Write(file, binary.LittleEndian, s)
	if err != nil {
		return fmt.Errorf("error al escribir la estructura Student: %v", err)
	}

	return nil
}

func (s *EBR) Deserialize(path string, offset int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Obtener el tamaño de la estructura FolderBlock
	fbSize := binary.Size(s)
	if fbSize <= 0 {
		return fmt.Errorf("invalid FolderBlock size: %d", fbSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura FolderBlock
	buffer := make([]byte, fbSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura FolderBlock
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, s)
	if err != nil {
		return err
	}

	return nil
}

func ReadEBRsFromFile(filename string, start int) (*EBR, error) {
	ebr := &EBR{}
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()
	currentOffset := int64(start)

	for {
		// Mueve el puntero del archivo a la posición del EBR actual
		_, err := file.Seek(currentOffset, 0)
		if err != nil {
			return nil, fmt.Errorf("error al mover el puntero del archivo: %v", err)
		}

		// Lee el EBR desde el archivo
		if err := binary.Read(file, binary.LittleEndian, ebr); err != nil {
			return nil, fmt.Errorf("error al leer el EBR: %v", err)
		}

		// Verifica si hay otro EBR
		nextOffset := int64(binary.LittleEndian.Uint32(ebr.Next[:]))
		if nextOffset == int64(0xFFFFFFFF) {
			break // No hay más EBRs, sal del bucle
		}

		currentOffset = nextOffset
	}

	return ebr, nil
}

func StartLogica(filename string, start int) (int, error) {
	var startL int
	ebr := &EBR{}
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()
	currentOffset := int64(start)

	startL = start + 30

	for {
		// Mueve el puntero del archivo a la posición del EBR actual
		_, err := file.Seek(currentOffset, 0)
		if err != nil {
			return 0, fmt.Errorf("error al mover el puntero del archivo: %v", err)
		}

		// Lee el EBR desde el archivo
		if err := binary.Read(file, binary.LittleEndian, ebr); err != nil {
			return 0, fmt.Errorf("error al leer el EBR: %v", err)
		}

		// Verifica si hay otro EBR
		nextOffset := int64(binary.LittleEndian.Uint32(ebr.Next[:]))
		if nextOffset == int64(0xFFFFFFFF) {
			break
		}

		startL = int(nextOffset) + 30
		currentOffset = nextOffset
	}
	fmt.Println("Startlogica:", startL)
	return startL, nil
}

func FindLogica(path string, start int, name string) bool {
	ebr := &EBR{}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	currentOffset := int64(start)

	for {
		// Mueve el puntero del archivo a la posición del EBR actual
		_, err := file.Seek(currentOffset, 0)
		if err != nil {
			return false
		}

		// Lee el EBR desde el archivo
		if err := binary.Read(file, binary.LittleEndian, ebr); err != nil {
			return false
		}

		partitionName := strings.TrimRight(string(ebr.Name[:]), "\x00")
		if partitionName == name {
			return true
		}

		nextOffset := int64(binary.LittleEndian.Uint32(ebr.Next[:]))
		if nextOffset == int64(0xFFFFFFFF) {
			break // No hay más EBRs, sal del bucle
		}

		currentOffset = nextOffset
	}

	return false
}
