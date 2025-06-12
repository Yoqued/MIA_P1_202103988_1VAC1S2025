package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type BLOCK_CONTARCHIVO struct {
	Content [64]byte // Array con el contenido del archivo
}

// Serialize escribe la estructura FileBlock en un archivo binario en la posición especificada
func (fb *BLOCK_CONTARCHIVO) Serialize(path string, offset int64) error {
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
	err = binary.Write(file, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

// Deserialize lee la estructura FolderBlock desde un archivo binario en la posición especificada
func (fb *BLOCK_CONTARCHIVO) Deserialize(path string, offset int64) error {
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
	fbSize := binary.Size(fb)
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
	err = binary.Read(reader, binary.LittleEndian, fb)
	if err != nil {
		return err
	}

	return nil
}

func ReadCont_Archivo(path string, offset int64) (*BLOCK_CONTARCHIVO, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Mover el puntero de lectura al offset especificado
	_, err = file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	ca := &BLOCK_CONTARCHIVO{}
	if err := binary.Read(file, binary.LittleEndian, ca); err != nil {
		return nil, err
	}

	return ca, nil
}

// PrintContent prints the content of B_content as a string
func (fb *BLOCK_CONTARCHIVO) Print() {
	fmt.Printf("%s", fb.Content)
}
