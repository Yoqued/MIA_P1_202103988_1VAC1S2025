package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type MKGRP struct {
	name string
}

func ParseMkgrp(tokens []string) ([]string, error) {
	cmd := &MKGRP{}
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
		case "-name":
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
	if groupExiste(cmd.name, util.GruposId[util.UserOnline.Id].GNames) {
		return tokens[contador:], errors.New("Este grupo ya existe, intenta otro nombre")
	}
	var UserString string

	startPart := int64(binary.LittleEndian.Uint32(util.UserOnline.Particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(util.UserOnline.Path, startPart)
	groupCadena := strconv.Itoa(util.GruposId[util.UserOnline.Id].GroupCount+2) + ",G," + cmd.name + "\n"

	err := superblock.ReadFile(util.UserOnline.Path, []string{"users.txt"}, 0, &UserString)
	if err != nil {
		return tokens[contador:], err
	}
	UserString += groupCadena
	fmt.Println("mkGRP", groupCadena, "\n ----")

	err = estructuras.ModifyUserGroup(superblock, util.UserOnline.Path, UserString)
	if err != nil {
		return tokens[contador:], err
	}
	fmt.Println("Grupo Creado con exito!")
	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Grupo %s creado con exito!", cmd.name))

	//Guardar todos los cambios del superbloque
	superblock.WriteToFile(util.UserOnline.Path, int(startPart))
	UpdateGroup(util.UserOnline.Id, cmd.name)

	return tokens[contador:], nil // Devuelve el comando MKDISK creado
}

func groupExiste(groupCadena string, Names []string) bool {
	var name string
	if strings.Contains(groupCadena, ",") {
		elementos := strings.Split(groupCadena, ",")
		name = strings.TrimSpace(elementos[2])
	} else {
		name = strings.TrimSpace(groupCadena)
	}

	for _, nombre := range Names {
		if nombre == name {
			return true
		}
	}
	return false
}

// Función para agregar un nuevo nombre al grupo en el mapa GruposId
func UpdateGroup(groupId string, newName string) {
	// Recuperas la estructura GRUPO actual del mapa
	grupo := util.GruposId[groupId]

	// Actualizar
	grupo.GNames = append(grupo.GNames, newName)
	grupo.GroupCount += 1
	util.GruposId[groupId] = grupo
}
