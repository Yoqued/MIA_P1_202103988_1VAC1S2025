package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	objs "MIA_P1_202103988_1VAC1S2025/Objetos"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func ParseLogin(tokens []string) ([]string, error) {
	cmd := &util.USER{}
	var contador int
	for cont, token := range tokens {

		if token == "<newline>" {
			contador += 1
			break
		}

		// Divide cada token en clave y valor usando "=" como delimitador
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			return tokens[cont:], fmt.Errorf("formato de parámetro inválido: %s", token)
		}
		key, value := strings.ToLower(parts[0]), parts[1]

		// Key representa las palabras claves de cada atributo
		switch key {
		case "-user":
			if len(value) > 10 {
				return tokens[cont:], errors.New("limite de caracteres, debe ser menos de 10")
			}
			cmd.User = value
		case "-pass":
			if len(value) > 10 {
				return tokens[cont:], errors.New("limite de caracteres, debe ser menos de 10")
			}
			cmd.Pass = value
		case "-id":
			cmd.Id = value
		default:
			return tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	if cmd.User == "" {
		return tokens[contador:], errors.New("faltan parámetros requeridos: -user")
	}
	if cmd.Pass == "" {
		return tokens[contador:], errors.New("faltan parametros requeridos: -pass")
	}
	if cmd.Id == "" {
		return tokens[contador:], errors.New("faltan parametros requeridos: -id")
	}

	if reflect.DeepEqual(util.UserOnline, util.USER{}) {
		_, err := Login(*cmd)
		if err != nil {
			return tokens[contador:], err
		}
		fmt.Println("Inicio de Sesion exitoso")
	} else {
		return tokens[contador:], errors.New("Ya hay una sesion activa, cierra sesion primero para abrir una nueva sesion")
	}

	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Inicio de Sesion exitoso con usuario %s", cmd.User))

	return tokens[contador:], nil // Devuelve el comando MKDISK creado
}

func Login(user util.USER) (bool, error) {
	var UserString string
	path, err := objs.BuscarPathID(user.Id)
	if err != nil {
		return false, err
	}
	mbr, _ := objs.ReadMbr(path)
	particion := mbr.BuscarIdParticion(user.Id)
	startPart := int64(binary.LittleEndian.Uint32(particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(path, startPart)

	err = superblock.ReadFile(path, []string{"users.txt"}, 0, &UserString)
	if err != nil {
		return false, err
	}
	fmt.Println("USUARIOS:", UserString)
	lineas := strings.Split(UserString, "\n")

	// Analizar el txt línea por línea
	for _, linea := range lineas {
		partes := strings.Split(linea, ",")
		fmt.Println("linea", linea)
		// Verificar si hay suficientes elementos
		if len(partes) >= 3 {
			if partes[1] == "U" {
				// Comparar user.user y user.pass con los valores correspondientes
				if partes[3] == user.User && partes[4] == user.Pass {
					fmt.Println("Autenticación exitosa para el usuario:", user.User)
					user.Particion = particion
					user.Path = path
					user.UId, _ = strconv.Atoi(partes[0])
					user.GId = util.BuscarGId(linea, util.GruposId[user.Id].GNames)
					util.UserOnline = user
					fmt.Println(util.UserOnline)
					return true, nil
				}
			}
		}
	}

	return false, errors.New("Autenticación fallida: usuario o contraseña incorrectos")
}
