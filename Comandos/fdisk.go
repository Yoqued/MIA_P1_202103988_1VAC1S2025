package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	objs "MIA_P1_202103988_1VAC1S2025/Objetos"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseFdisk(tokens []string) (*objs.PARTICION, []string, error) {
	cmd := &objs.PARTICION{}
	var unit, path, tipo, fit, delete string
	var contador, Sizes, add int
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
		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, tokens[cont:], errors.New("el tamaño debe ser un número entero positivo")
			}
			Sizes = size
		case "-unit":
			value = strings.ToUpper(value)
			// Verifica que la unidad sea "K" o "M" o "B"
			if value != "B" && value != "K" && value != "M" {
				return nil, tokens[cont:], errors.New("la unidad debe ser K o M")
			}
			unit = value
		case "-path":
			path = value
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, tokens[cont:], errors.New("el ajuste debe ser BF, FF o WF")
			}
			fit = string(value[0])
		case "-type":
			value = strings.ToUpper(value)
			// Verifica que la unidad sea "K" o "M"
			if value != "P" && value != "E" && value != "L" {
				return nil, tokens[cont:], errors.New("el tipo debe ser P, E o L")
			}
			tipo = value
			copy(cmd.Tipo[:], tipo)
		case "-name":
			fmt.Println(value)
			copy(cmd.Name[:], value)
		case "-delete":
			value = strings.ToUpper(value)
			// Verifica que sea fast o full
			if value != "FAST" && value != "FULL" {
				return nil, tokens[cont:], errors.New("el tipo de delete debe ser full o fast")
			}
			delete = value
		case "-add":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil {
				return nil, tokens[cont:], err
			}
			add = size
		default:
			return nil, tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}
	if delete != "" && add != 0 {
		return nil, tokens[contador:], errors.New("no se puede hacer un delete y un add a la vez")
	}
	if path == "" {
		return nil, tokens[contador:], errors.New("faltan parametros requeridos: -path")
	}
	if objs.IsEmptyByte(cmd.Name[:]) {
		return nil, tokens[contador:], errors.New("faltan parametros requeridos: -name")
	}
	if delete != "" {
		name := string(bytes.Trim(cmd.Name[:], "\x00"))
		err := deleteParticion(path, name, delete)
		if err != nil {
			return nil, tokens[contador:], err
		}
		return cmd, tokens[contador:], nil
	}

	if Sizes == 0 {
		return nil, tokens[contador:], errors.New("faltan parámetros requeridos: -size")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if unit == "" {
		unit = "K"
	}
	if fit == "" {
		fit = "W"
	}

	tamano, err := estructuras.ConvertToBytes(Sizes, unit)
	if err != nil {
		return nil, tokens[contador:], err
	}

	fmt.Println("TAMANOOO", tamano)

	disk, err := objs.BuscarDisco(path)
	if err != nil {
		return nil, tokens[contador:], err
	}
	fmt.Println("TAMAÑOS", disk.FreeSpace, tamano)
	if disk.FreeSpace < tamano {
		return nil, tokens[contador:], errors.New("Espacio Insuficiente en el disco")
	} else {
		disk.FreeSpace -= tamano
	}

	fmt.Println(tamano)
	binary.LittleEndian.PutUint32(cmd.Size[:], uint32(tamano))
	copy(cmd.Status[:], "0")
	copy(cmd.Fit[:], fit)

	mbr, err := objs.ReadMbr(path)
	fmt.Println("mbr", mbr)
	if err != nil {
		fmt.Println("---------------ERROR2-------------", err)
		return nil, tokens[contador:], err
	}
	binary.LittleEndian.PutUint32(cmd.Start[:], uint32(CalcularStart(mbr)))

	if tipo == "P" {
		err = mbr.AgregarParticion(*cmd)
		if err != nil {
			fmt.Println("---------------ERROR-------------")
			fmt.Println(err)
			return nil, tokens[contador:], err
		}
		mbr.WriteToFile(path)
	} else if tipo == "E" {
		if !mbr.VerificarExtendida() {
			ebr := &objs.EBR{}
			copy(ebr.Fit[:], cmd.Fit[:])
			binary.LittleEndian.PutUint32(ebr.Next[:], uint32(0xFFFFFFFF))
			err = mbr.AgregarParticion(*cmd)
			if err != nil {
				fmt.Println("---------------ERROR-------------")
				fmt.Println(err)
				return nil, tokens[contador:], err
			}
			mbr.WriteToFile(path)
			ebr.WriteToFile(path, int(binary.LittleEndian.Uint32(cmd.Start[:])))
		} else {
			return nil, tokens[contador:], errors.New("ya existe una particion extendida en este disco")
		}
	} else if tipo == "L" {
		if mbr.VerificarExtendida() {
			startExt := mbr.StartExtendida()
			ebr, err := objs.ReadEBRsFromFile(path, startExt)
			if err != nil {
				return nil, tokens[contador:], err
			}
			startLogica, err := objs.StartLogica(path, startExt)
			if err != nil {
				return nil, tokens[contador:], err
			}
			//Empezar a llenar el ebr con la info de la particion logica
			startNextEbr := startLogica + tamano
			copy(ebr.Status[:], "0")
			copy(ebr.Name[:], cmd.Name[:])
			binary.LittleEndian.PutUint32(ebr.Size[:], uint32(tamano))
			binary.LittleEndian.PutUint32(ebr.Start[:], uint32(startLogica))
			binary.LittleEndian.PutUint32(ebr.Next[:], uint32(startNextEbr))
			ebr.WriteToFile(path, startLogica-30)
			//Crear el siguiente ebr
			nextebr := &objs.EBR{}
			copy(nextebr.Fit[:], ebr.Fit[:])
			binary.LittleEndian.PutUint32(nextebr.Next[:], uint32(0xFFFFFFFF))
			nextebr.WriteToFile(path, startNextEbr)

		} else {
			return nil, tokens[contador:], errors.New("no existe una particion extendida en este disco para la logica")
		}
	}
	name := string(bytes.Trim(cmd.Name[:], "\x00"))
	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Particion %s creada con exito", name))
	return cmd, tokens[contador:], nil // Devuelve el comando MKDISK creado
}

func CalcularStart(mbr *objs.MBR) int {
	var start = 153
	particiones := mbr.Particiones
	for _, particion := range particiones {
		size := int(binary.LittleEndian.Uint32(particion.Size[:]))
		startpart := int(binary.LittleEndian.Uint32(particion.Start[:]))
		fmt.Println(size, "+", startpart)
		if objs.IsEmptyByte(particion.Name[:]) {
			break
		}
		start = startpart + size
	}
	fmt.Println("Start", start)
	return start
}

func deleteParticion(path string, name string, tipo string) error {
	mbr, err := objs.ReadMbr(path)
	if err != nil {
		return err
	}

	particion := mbr.BuscarParticion(name)
	if particion == nil {
		return fmt.Errorf("la partición '%s' no fue encontrada", name)
	}

	if tipo == "FULL" {
		fmt.Println("ELIMINAR FULL")
		// Abrir el archivo en modo de escritura
		file, err := os.OpenFile(path, os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("no se pudo abrir el archivo: %v", err)
		}
		defer file.Close()

		// Moverse a la posición de inicio de la partición
		start := int64(binary.LittleEndian.Uint32(particion.Start[:]))
		_, err = file.Seek(start, 0)
		if err != nil {
			return fmt.Errorf("no se pudo mover a la posición de inicio de la partición: %v", err)
		}

		// Crear un buffer lleno de '\0' del tamaño de la partición
		size := int(binary.LittleEndian.Uint32(particion.Size[:]))
		zeroes := make([]byte, size)

		// Escribir el buffer en el archivo
		_, err = file.Write(zeroes)
		if err != nil {
			return fmt.Errorf("no se pudo sobrescribir la partición: %v", err)
		}
	}

	particion.Clear()
	err = mbr.WriteToFile(path)
	if err != nil {
		return err
	}

	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Particion %s removida con exito, con el protocolo %s", name, tipo))

	return nil
}
