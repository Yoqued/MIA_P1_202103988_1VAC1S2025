package objs

import (
	"bytes"
	"errors"
	"fmt"
)

type DISK struct {
	Name      string
	Size      int
	Fit       string
	Path      string
	Id        []string
	Letter    string
	FreeSpace int
}

var Discos []DISK

func BuscarDisco(path string) (*DISK, error) {
	for i, disco := range Discos {
		if disco.Path == path {
			return &Discos[i], nil
		}
	}
	return nil, errors.New("Ruta del Disco NO Encontrada")
}

func BuscarPathID(idS string) (string, error) {
	for i, disco := range Discos {
		for _, id := range disco.Id {
			fmt.Println(idS, id)
			if len(idS) > 3 && len(id) > 3 {
				if string(id[3]) == string(idS[3]) {
					return Discos[i].Path, nil
				}
			} else {
				return "", errors.New("el ID no cumple con el formato establecido (4 caracteres)")
			}
		}
	}
	return "", errors.New("Id no encontrado")
}

func IsEmptyByte(arr []byte) bool {
	return bytes.Equal(arr, make([]byte, len(arr)))
}

func RemoveIdDisk(path string, id string) error {
	// Buscar el disco por su path
	disco, err := BuscarDisco(path)
	if err != nil {
		return err
	}

	// Buscar y eliminar el id en el array disco.Id
	for i, v := range disco.Id {
		if v == id {
			// Eliminar el id encontrado
			disco.Id = append(disco.Id[:i], disco.Id[i+1:]...)
			fmt.Printf("ID '%s' eliminado del disco '%s'\n", id, path)
			return nil
		}
	}

	return errors.New("ID no encontrado en el disco")
}
