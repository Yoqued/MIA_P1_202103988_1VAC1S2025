package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

var BlockTypes = make(map[int]string)

// Bloque de CARPETA
type BLOCK_CARPETA struct {
	Content [4]BCONTENT // Array con el contenido de la carpeta
}

func (bc *BLOCK_CARPETA) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura FileBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, bc)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura FolderBlock desde un archivo binario en la posición especificada
func (bc *BLOCK_CARPETA) Deserialize(path string, offset int64) error {
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
	fbSize := binary.Size(bc)
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
	err = binary.Read(reader, binary.LittleEndian, bc)
	if err != nil {
		return err
	}

	return nil
}

func (bc *BLOCK_CARPETA) Print() {
	for i, content := range bc.Content {
		name := string(content.Name[:])
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.Inodo)
	}
}

type BCONTENT struct {
	Name  [12]byte // Nombre de la carpeta o archivo
	Inodo int32    // Apuntador hacia un inodo asociado
}

func (bc *BCONTENT) Print() {
	name := string(bc.Name[:])
	fmt.Printf("  name: %s\n", name)
	fmt.Printf("  inodo: %d\n", bc.Inodo)

}

// Bloque de Apuntadores
type BLOCK_APUNTADOR struct {
	Pointers [16]int32 // Array con los apuntadores a bloques (de archivo o carpeta)
}

// Constructor para BLOCK_APUNTADOR que inicializa todos los valores a -1
func NewBlockApuntador() *BLOCK_APUNTADOR {
	ba := &BLOCK_APUNTADOR{}
	for i := range ba.Pointers {
		ba.Pointers[i] = -1
	}
	return ba
}

func (ba *BLOCK_APUNTADOR) Serialize(path string, offset int64) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}

	// Serializar la estructura FileBlock directamente en el archivo
	err = binary.Write(file, binary.LittleEndian, ba)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura FolderBlock desde un archivo binario en la posición especificada
func (ba *BLOCK_APUNTADOR) Deserialize(path string, offset int64) error {
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
	fbSize := binary.Size(ba)
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
	err = binary.Read(reader, binary.LittleEndian, ba)
	if err != nil {
		return err
	}

	return nil
}

func (ba *BLOCK_APUNTADOR) Print() {
	for i, content := range ba.Pointers {
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %d\n", content)
	}
}
