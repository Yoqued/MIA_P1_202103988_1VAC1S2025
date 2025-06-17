package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type CAT struct {
	filen []string
}

func ParseCat(tokens []string) ([]string, error, string) {
	cmd := &CAT{}
	var contador int
	var salida string
	er := regexp.MustCompile(`^-file\d+$`)
	for cont, token := range tokens {

		if token == "<newline>" {
			contador += 1
			break
		}

		// Divide cada token en clave y valor usando "=" como delimitador
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			return tokens[cont:], fmt.Errorf("formato de parámetro inválido: %s", token), ""
		}
		key, value := strings.ToLower(parts[0]), parts[1]

		if er.MatchString(key) {
			cmd.filen = append(cmd.filen, value)
		} else {
			return tokens[cont:], fmt.Errorf("formato no valido"), ""
		}
		contador += 1
	}
	if reflect.DeepEqual(util.UserOnline, util.USER{}) {
		return tokens[contador:], errors.New("Se necesita iniciar sesion para ejecutar este comando"), ""
	}

	startPart := int64(binary.LittleEndian.Uint32(util.UserOnline.Particion.Start[:]))
	superblock, _ := estructuras.ReadSuperBlock(util.UserOnline.Path, startPart)

	for i := 0; i < len(cmd.filen); i++ {
		fmt.Println("ruta actual:", cmd.filen[i])
		fileParts := strings.Split(cmd.filen[i], "/")
		fileParts = fileParts[1:]
		err := superblock.ReadFile(util.UserOnline.Path, fileParts, 0, &salida)
		if err != nil {
			return tokens[contador:], err, ""
		}
	}
	fmt.Println("SALIDA DE CONSOLA:")
	fmt.Println(salida)

	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Contenido de los archivos %s con comando CAT:", cmd.filen[:]))

	return tokens[contador:], nil, salida // Devuelve el comando MKDISK creado
}
