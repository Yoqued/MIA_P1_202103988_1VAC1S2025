package estructuras

import (
	"encoding/binary"
	"os"
)

// CreateBitMaps crea los Bitmaps de inodos y bloques en el archivo especificado
func (sb *SUPERBLOCK) CreateBitMaps(path string) error {
	// Escribir Bitmaps
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Bitmap de inodos
	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(int64(sb.InodeBitmapStart), 0)
	if err != nil {
		return err
	}

	// Crear un buffer de n '0'
	buffer := make([]byte, sb.FreeInodos)
	for i := range buffer {
		buffer[i] = '0'
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	// Bitmap de bloques
	// Mover el puntero del archivo a la posición especificada
	_, err = file.Seek(int64(sb.BlockBitmapStart), 0)
	if err != nil {
		return err
	}

	// Crear un buffer de n 'O'
	buffer = make([]byte, sb.FreeBlocks)
	for i := range buffer {
		buffer[i] = 'O'
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return err
	}

	return nil
}

func (sb *SUPERBLOCK) UpdateBitmapInodo(path string) error {
	// Abrir el archivo
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición del bitmap de inodos
	_, err = file.Seek(int64(sb.InodeBitmapStart)+int64(sb.CountInodos), 0)
	if err != nil {
		return err
	}

	// Escribir el bit en el archivo
	_, err = file.Write([]byte{'1'})
	if err != nil {
		return err
	}

	return nil
}

func (sb *SUPERBLOCK) UpdateBitmapBlock(path string) error {
	// Abrir el archivo
	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Mover el puntero del archivo a la posición del bitmap de bloques
	_, err = file.Seek(int64(sb.BlockBitmapStart)+int64(sb.CountBlocks), 0)
	if err != nil {
		return err
	}

	// Escribir el bit en el archivo
	_, err = file.Write([]byte{'X'})
	if err != nil {
		return err
	}

	return nil
}
