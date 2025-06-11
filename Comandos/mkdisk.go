package comandos

import (
	estructuras "MIA_P1_202103988_1VAC1S2025/Estructuras"
	objs "MIA_P1_202103988_1VAC1S2025/Objetos"
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ParseMkdisk(tokens []string) (*objs.DISK, []string, error) {
	cmd := &objs.DISK{}
	var contador, Size int
	var unit string
	fmt.Println(tokens)
	for _, token := range tokens {

		if token == "<newline>" {
			contador += 1
			break
		}

		// Divide cada token en clave y valor usando "=" como delimitador
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			return nil, tokens[contador:], fmt.Errorf("formato de parámetro inválido: %s", token)
		}
		key, value := strings.ToLower(parts[0]), parts[1]

		// Key representa las palabras claves de cada atributo
		key = strings.ToLower(key)
		switch key {

		case "-size":
			// Convierte el valor del tamaño a un entero
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return nil, tokens[contador:], errors.New("el tamaño debe ser un número entero positivo")
			}
			Size = size
		case "-unit":
			// Verifica que la unidad sea "K" o "M"
			value = strings.ToUpper(value)
			if value != "K" && value != "M" {
				return nil, tokens[contador:], errors.New("la unidad debe ser K o M")
			}
			unit = value
		case "-path":
			cmd.Path = value
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return nil, tokens[contador:], errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.Fit = value
		default:
			return nil, tokens[contador:], fmt.Errorf("parámetro desconocido: %s", key)
		}
		contador += 1
	}

	if Size == 0 {
		return nil, tokens[contador:], errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.Path == "" {
		return nil, tokens[contador:], errors.New("faltan parámetros requeridos: -path")
	}

	// Si no se proporcionó la unidad, se establece por defecto a "M"
	if unit == "" {
		unit = "M"
	}
	if cmd.Fit == "" {
		cmd.Fit = "FF"
	}

	// Llama a la función CreateBinaryFile del paquete disk para crear el archivo binario
	cmd.Size, _ = estructuras.ConvertToBytes(Size, unit)
	cmd.FreeSpace, _ = estructuras.ConvertToBytes(Size, unit)
	err := estructuras.CreateBinaryFile(cmd.Path, cmd.Size)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, tokens[contador:], err
	}
	cmd.Name = filepath.Base(cmd.Path)
	objs.Discos = append(objs.Discos, *cmd)

	CrearMbr(cmd)
	fmt.Println("Se creo el disco")
	util.Respuestas = append(util.Respuestas, fmt.Sprintf("Disco %s creado con exito!", cmd.Path))
	return cmd, tokens[contador:], nil // Devuelve el comando MKDISK creado
}

func CrearMbr(cmd *objs.DISK) error {
	mbr := &objs.MBR{}

	binary.LittleEndian.PutUint32(mbr.Size[:], uint32(cmd.Size))

	//Ingresa la fecha y hora actual
	now := time.Now()
	yearBase := 2000
	daysSinceBase := (now.Year()-yearBase)*365 + now.YearDay()
	secondsInDay := now.Hour()*3600 + now.Minute()*60 + now.Second()
	dayFraction := float32(secondsInDay) / 86400.0
	fechaHora := float32(daysSinceBase) + dayFraction
	// Convertir a bytes
	binary.LittleEndian.PutUint32(mbr.Fecha[:], math.Float32bits(fechaHora))

	// Generar y asignar un signature aleatorio
	binary.LittleEndian.PutUint32(mbr.Signature[:], GenerateDiskID())

	//Primer caracter de fit
	copy(mbr.Fit[:], string(cmd.Fit[0]))

	if err := mbr.WriteToFile(cmd.Path); err != nil {
		fmt.Println("Error al escribir el archivo:", err)
		return err
	}

	fmt.Println("Mbr creado con exito")

	return nil
}

func GenerateDiskID() uint32 {
	// Inicializa la semilla para el generador de números aleatorios usando la hora actual
	rand.Seed(time.Now().UnixNano())

	// Genera un número aleatorio de 32 bits (uint32)
	id := rand.Uint32()
	return id
}
