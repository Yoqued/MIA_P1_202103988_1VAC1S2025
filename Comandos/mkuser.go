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

func ParseMkusr(tokens []string) ([]string, error) {
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
		case "-grp":
			if len(value) > 10 {
				return tokens[cont:], errors.New("limite de caracteres, debe ser menos de 10")
			}
			cmd.Grp = value
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
	if cmd.Grp == "" {
		return tokens[contador:], errors.New("faltan parametros requeridos: -grp")
	}
	if reflect.DeepEqual(util.UserOnline, util.USER{}) {
		return tokens[contador:], errors.New("Se necesita iniciar sesion para ejecutar este comando")
	}
	if util.UsuarioExiste(cmd.User, util.GruposId[util.UserOnline.Id].UNames) {
		return tokens[contador:], fmt.Errorf("El usuario \"%s\" ya existe", cmd.User)
	}
	if !groupExiste(cmd.Grp, util.GruposId[util.UserOnline.Id].GNames) {
		return tokens[contador:], fmt.Errorf("El grupo indicado \"%s\" no existe", cmd.Grp)
	}
	var UserString string

	startPart := int64(binary.LittleEndian.Uint32(util.UserOnline.Particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(util.UserOnline.Path, startPart)
	userCadena := strconv.Itoa(util.GruposId[util.UserOnline.Id].UserCount+2) + ",U," + cmd.Grp + "," + cmd.User + "," + cmd.Pass + "\n"

	err := superblock.ReadFile(util.UserOnline.Path, []string{"users.txt"}, 0, &UserString)
	if err != nil {
		return tokens[contador:], err
	}
	UserString += userCadena
	fmt.Println("mkUSERS", userCadena, "\n ----")

	err = estructuras.ModifyUserGroup(superblock, util.UserOnline.Path, UserString)
	if err != nil {
		return tokens[contador:], err
	}
	fmt.Println("Usuario Creado con exito!")
	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Usuario %s creado con exito!", cmd.User))

	//Guardar todos los cambios del superbloque
	superblock.WriteToFile(util.UserOnline.Path, int(startPart))
	util.UpdateUser(util.UserOnline.Id, cmd.User)

	return tokens[contador:], nil
}
