package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	objs "MIA_P1_202103988_1VAC1S2025/Objetos"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"
)

type MKFS struct {
	id   string
	tipo string
	fs   string
}

func ParseMkfs(tokens []string) (*MKFS, []string, error) {
	cmd := &MKFS{}
	var contador int
	for cont, token := range tokens {

		if token == "<newline>" {
			contador += 1
			break
		}

		// Divide cada token en clave y valor usando "=" como delimitador
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			return nil, tokens[cont:], fmt.Errorf("formato de parámetro inválido: %s", token)
		}
		key, value := strings.ToLower(parts[0]), parts[1]

		// Key representa las palabras claves de cada atributo
		switch key {
		case "-id":
			cmd.id = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "FULL" {
				return nil, tokens[cont:], errors.New("tipo no aceptado")
			}
			cmd.tipo = value
		case "-fs":
			value = strings.ToUpper(value)
			if value != "2FS" && value != "3FS" {
				return nil, tokens[cont:], errors.New("Formato de fs incorrecto, unicamente se puede ext2 o ext3")
			}
			cmd.fs = value
		default:
			return nil, tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	if cmd.id == "" {
		return nil, tokens[contador:], errors.New("faltan parámetros requeridos: -id")
	}
	if cmd.fs == "" {
		cmd.fs = "2FS"
	}

	// Buscar la ruta del disco donde se encuentra la particion
	path, err := objs.BuscarPathID(cmd.id)
	if err != nil {
		return nil, tokens[contador:], err
	}

	mbr, err := objs.ReadMbr(path)
	if err != nil {
		return nil, tokens[contador:], err
	}

	particion := mbr.BuscarIdParticion(cmd.id)
	if particion == nil {
		return nil, tokens[contador:], fmt.Errorf("no se eoncontro particion con id: %s", cmd.id)
	}

	//Escribir el ext2 en la particion
	crearSuperBloque(path, particion, cmd)
	util.GruposId[cmd.id] = util.GRUPO{
		GNames:     []string{"root"},
		UNames:     []string{"root"},
		GroupCount: 0,
		UserCount:  0,
	}

	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Sistema de archivos ext2 creado con exit en id %s", cmd.id))

	return cmd, tokens[contador:], nil // Devuelve el comando MKDISK creado
}

func crearSuperBloque(path string, particion *objs.PARTICION, cmd *MKFS) error {
	n_structures := estructuras.CalcularNumEstructuras(int(binary.LittleEndian.Uint32(particion.Size[:])))
	bm_inodeStart := int(binary.LittleEndian.Uint32(particion.Start[:])) + binary.Size(estructuras.SUPERBLOCK{})
	bm_blockStart := bm_inodeStart + n_structures
	inodo_start := bm_blockStart + (3 * n_structures)
	block_start := inodo_start + (binary.Size(estructuras.INODO{}) * n_structures)
	superblock := estructuras.SUPERBLOCK{
		Filesystem:       2,
		CountInodos:      0,
		CountBlocks:      0,
		FreeInodos:       int32(n_structures),
		FreeBlocks:       int32(n_structures * 3),
		MontTime:         float32(time.Now().Unix()),
		UMontTime:        float32(time.Now().Add(-24 * time.Hour).Unix()),
		CountMount:       int32(1),
		Magic:            0xEF53,
		InodeSize:        int32(binary.Size(estructuras.INODO{})),
		BlockSize:        int32(binary.Size(estructuras.BLOCK_CARPETA{})),
		FirstFreeInode:   int32(inodo_start),
		FirstFreeBlock:   int32(block_start),
		InodeBitmapStart: int32(bm_inodeStart),
		BlockBitmapStart: int32(bm_blockStart),
		InodeStart:       int32(inodo_start),
		BlockStart:       int32(block_start),
	}
	err := superblock.CreateBitMaps(path)
	if err != nil {
		return err
	}

	if cmd.fs == "2FS" {
		err := superblock.CreateUsersFile(path)
		if err != nil {
			return err
		}
	} else {
		err := superblock.CreateUsersFileExt3(path, int64(int32(binary.LittleEndian.Uint32(particion.Start[:]))+int32(binary.Size(estructuras.SUPERBLOCK{}))))
		if err != nil {
			return err
		}
	}

	err = superblock.WriteToFile(path, int(binary.LittleEndian.Uint32(particion.Start[:])))
	if err != nil {
		return err
	}

	return nil
}
