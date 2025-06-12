package estructuras

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

type INODO struct {
	Uid        int32     // Id del usuario propietario
	Gid        int32     // Id del grupo que pertenece
	Size       int32     // Tamaño del archivo
	Lastime    float32   // Ultima fecha de lectura del inodo
	CreateTime float32   // Fecha que se creo el inodo
	ModifyTime float32   // Ultima fecha de modificacion inodo
	Blocks     [15]int32 // Array de direcciones de 15 bloques
	Tipo       [1]byte   // Indica si es archivo o carpeta
	Perm       [3]byte   // Permisos del archivo o carpeta
}

// Escribe la estructura del Inodo
func (ino *INODO) WriteToFile(path string, start int) error {
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

	err = binary.Write(file, binary.LittleEndian, ino)
	if err != nil {
		return fmt.Errorf("error al escribir la estructura Student: %v", err)
	}

	return nil
}

// Deserialize lee la estructura Inode desde un archivo binario en la posición especificada
func (ino *INODO) Deserialize(path string, offset int64) error {
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

	// Obtener el tamaño de la estructura Inode
	inodeSize := binary.Size(ino)
	if inodeSize <= 0 {
		return fmt.Errorf("invalid Inode size: %d", inodeSize)
	}

	// Leer solo la cantidad de bytes que corresponden al tamaño de la estructura Inode
	buffer := make([]byte, inodeSize)
	_, err = file.Read(buffer)
	if err != nil {
		return err
	}

	// Deserializar los bytes leídos en la estructura Inode
	reader := bytes.NewReader(buffer)
	err = binary.Read(reader, binary.LittleEndian, ino)
	if err != nil {
		return err
	}

	return nil
}

func (inode *INODO) Print() {
	atime := time.Unix(int64(inode.CreateTime), 0)
	ctime := time.Unix(int64(inode.Lastime), 0)
	mtime := time.Unix(int64(inode.ModifyTime), 0)

	fmt.Printf("I_uid: %d\n", inode.Uid)
	fmt.Printf("I_gid: %d\n", inode.Gid)
	fmt.Printf("I_size: %d\n", inode.Size)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.Blocks)
	fmt.Printf("I_type: %s\n", string(inode.Tipo[:]))
	fmt.Printf("I_perm: %s\n", string(inode.Perm[:]))
	fmt.Println("---")
}
