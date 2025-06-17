package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

type RMUSR struct {
	name string
}

func ParseRmusr(tokens []string) ([]string, error) {
	cmd := &RMUSR{}
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
			cmd.name = value
		default:
			return tokens[cont:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	if cmd.name == "" {
		return tokens[contador:], errors.New("faltan parámetros requeridos: -user")
	}
	if reflect.DeepEqual(util.UserOnline, util.USER{}) {
		return tokens[contador:], errors.New("Se necesita iniciar sesion para ejecutar este comando")
	}
	if !util.UsuarioExiste(cmd.name, util.GruposId[util.UserOnline.Id].UNames) {
		return tokens[contador:], errors.New("El usuario indicado no existe")
	}
	var UserString string

	// Extraer el superbloque de la particion
	startPart := int64(binary.LittleEndian.Uint32(util.UserOnline.Particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(util.UserOnline.Path, startPart)
	superblock.ReadFile(util.UserOnline.Path, []string{"users.txt"}, 0, &UserString)
	lineas := strings.Split(UserString, "\n")

	var newLines []string
	// Analizar el txt línea por línea
	for _, linea := range lineas {
		partes := strings.Split(linea, ",")
		if len(partes) >= 3 {
			if partes[1] == "U" {
				if partes[3] == cmd.name {
					partes[0] = "0"
				}
			}
		}
		newLines = append(newLines, strings.Join(partes, ","))
	}
	newUserString := strings.Join(newLines, "\n")
	fmt.Println("Nueva cadena a guardar:", newUserString)
	err := estructuras.ModifyUserGroup(superblock, util.UserOnline.Path, newUserString)
	if err != nil {
		return tokens[contador:], err
	}
	fmt.Println("Usuario removido con exito")
	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Usuario %s removido con exito!", cmd.name))

	//Guardar todos los cambios del superbloque
	superblock.WriteToFile(util.UserOnline.Path, int(startPart))

	return tokens[contador:], nil
}
