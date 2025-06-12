package estructuras

import (
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

type SHOWFILES struct {
	Name string
	Type string
}

func CalcularNumEstructuras(sizePartition int) int {
	// tamaño_particion = sizeOf(superblock) + n + 3 * n + n * sizeOf(inodos) + 3 * n * sizeOf(block)
	// n := float64(sizePartition-sizeSuperblock) / float64(4+sizeInode + 3*sizeBlock)
	n := float64(sizePartition-binary.Size(SUPERBLOCK{})) / float64(4+binary.Size(INODO{})+3*binary.Size(BLOCK_CARPETA{}))
	return int(math.Floor(n))
}

// Funcion para agregar alguna modificacion del contenido de users.txt

func ModifyUserGroup(sb *SUPERBLOCK, path string, cadena string) error {
	// Deserializar el bloque de carpeta raíz para encontrar el inodo de users.txt
	rootBlock := &BLOCK_CARPETA{}
	err := rootBlock.Deserialize(path, int64(sb.BlockStart))
	if err != nil {
		return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
	}

	// Buscar el inodo asociado a users.txt
	var usersInodoNum int32 = -1
	for _, content := range rootBlock.Content {
		filename := strings.TrimRight(string(content.Name[:]), "\x00")
		if filename == "users.txt" {
			usersInodoNum = content.Inodo
			break
		}
	}

	if usersInodoNum == -1 {
		return fmt.Errorf("users.txt no encontrado en la carpeta raíz")
	}

	// Leer el inodo de users.txt
	usersInode := &INODO{}
	offsetUser := sb.InodeStart + (usersInodoNum * sb.InodeSize)
	err = usersInode.Deserialize(path, int64(offsetUser))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de users.txt: %v", err)
	}

	// Variable para manejar el bloque actual
	var usersBlock *BLOCK_CONTARCHIVO
	var StartBlockuser int32
	newSize := len(cadena)

	// Rellenar espacio con los bloques ya utilizados
	for i := 0; i < len(usersInode.Blocks); i++ {
		if usersInode.Blocks[i] == -1 {
			break
		}
		StartBlockuser = sb.BlockStart + (usersInode.Blocks[i] * sb.BlockSize)
		usersBlock, err = ReadCont_Archivo(path, int64(StartBlockuser))
		if err != nil {
			return fmt.Errorf("error al leer el bloque de users.txt: %v", err)
		}
		if i < 12 {
			if len(cadena) > 64 {
				copy(usersBlock.Content[:], (cadena)[:64])
				cadena = (cadena)[64:]
			} else {
				// Limpiar el contenido del bloque antes de copiar
				for i := range usersBlock.Content {
					usersBlock.Content[i] = '\x00'
				}
				copy(usersBlock.Content[:], cadena)
				cadena = ""
			}
			usersBlock.Serialize(path, int64(StartBlockuser))
		} else if i == 12 {
			fmt.Println("Rellenando Bloque Indirecto Simple")
			err := sb.handleRecursivePointer(path, &cadena, &usersInode.Blocks[i], 1)
			if err != nil {
				return err
			}
			fmt.Println("Salio de indirecto simple", cadena)
		} else if i == 13 {
			err := sb.handleRecursivePointer(path, &cadena, &usersInode.Blocks[i], 2)
			if err != nil {
				return err
			}
		} else if i == 14 { // Apuntador Indirecto Triple
			err := sb.handleRecursivePointer(path, &cadena, &usersInode.Blocks[i], 3)
			if err != nil {
				return err
			}
		}
	}

	// En caso aun quede cadena por guardar se tendran que quedar nuevos bloques
	err = usersInode.Deserialize(path, int64(offsetUser))
	fmt.Println("ANTES DE FOR", usersInode)
	for i := 0; i < len(usersInode.Blocks); i++ {
		if usersInode.Blocks[i] != -1 {
			continue
		}
		if len(cadena) > 0 {
			// Apuntador Directo
			fmt.Println(cadena, "index", i)
			if i < 12 {
				fmt.Println("Creación de nuevo Bloque", cadena)
				newBlock := &BLOCK_CONTARCHIVO{}
				if len(cadena) > 64 {
					copy(usersBlock.Content[:], (cadena)[:64])
					cadena = (cadena)[64:]
				} else {
					copy(newBlock.Content[:], cadena)
					cadena = ""
				}

				fmt.Println("NUEVA CADENA", cadena)

				// Asignar el nuevo bloque a un apuntador directo del inodo
				usersInode.Blocks[i] = sb.CountBlocks
				BlockTypes[int(sb.CountBlocks)] = "CONTARCHIVO"
				err = newBlock.Serialize(path, int64(sb.FirstFreeBlock))
				if err != nil {
					return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
				}
				sb.UpdateBitmapBlock(path)
				sb.CountBlocks++
				sb.FreeBlocks--
				sb.FirstFreeBlock += sb.BlockSize

			} else if i == 12 { // Apuntador Indirecto Simple
				fmt.Println("Creación de nuevo Bloque Indirecto Simple")
				err := sb.handleRecursivePointer(path, &cadena, &usersInode.Blocks[i], 1)
				if err != nil {
					return err
				}
			} else if i == 13 { // Apuntador Indirecto Doble
				fmt.Println("Creación de nuevo Bloque Indirecto Doble")
				err := sb.handleRecursivePointer(path, &cadena, &usersInode.Blocks[i], 2)
				if err != nil {
					return err
				}
			} else if i == 14 { // Apuntador Indirecto Triple
				fmt.Println("Creación de nuevo Bloque Indirecto Triple")
				err := sb.handleRecursivePointer(path, &cadena, &usersInode.Blocks[i], 3)
				if err != nil {
					return err
				}
			}
		}
	}

	// Actualizar el tamaño del archivo en el inodo
	usersInode.Size = int32(newSize)
	usersInode.ModifyTime = float32(time.Now().Unix())
	// Serializar el inodo actualizado
	err = usersInode.WriteToFile(path, int(sb.InodeStart+(usersInodoNum*sb.InodeSize)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo de users.txt: %v", err)
	}
	return nil
}

func (sb *SUPERBLOCK) handleRecursivePointer(path string, cadena *string, pointer *int32, depth int) error {
	fmt.Println("ENTRO A RECURSIVE POINTER")
	// Caso base: escribir en un bloque de contenido de archivo
	if depth == 0 {
		if *pointer == -1 {
			fmt.Println("CREANDO NUEVO BLOQUE", sb.CountBlocks)
			newBlock := &BLOCK_CONTARCHIVO{}
			if len(*cadena) > 64 {
				copy(newBlock.Content[:], (*cadena)[:64])
				*cadena = (*cadena)[64:]
			} else {
				copy(newBlock.Content[:], *cadena)
				*cadena = ""
			}

			// Serializar el bloque de contenido
			err := newBlock.Serialize(path, int64(sb.FirstFreeBlock))
			if err != nil {
				return fmt.Errorf("error al serializar bloque de contenido: %v", err)
			}
			*pointer = sb.CountBlocks
			BlockTypes[int(sb.CountBlocks)] = "CONTARCHIVO"
			sb.UpdateBitmapBlock(path)
			sb.CountBlocks++
			sb.FreeBlocks--
			sb.FirstFreeBlock += sb.BlockSize
			return nil
		} else {
			fmt.Println("Rellenando bloque existente", *pointer)
			fmt.Println(*cadena)
			newBlock := &BLOCK_CONTARCHIVO{}
			if len(*cadena) > 64 {
				copy(newBlock.Content[:], (*cadena)[:64])
				*cadena = (*cadena)[64:]
			} else {
				copy(newBlock.Content[:], *cadena)
				*cadena = ""
			}

			// Serializar el bloque de contenido
			err := newBlock.Serialize(path, int64(sb.BlockStart+(*pointer*sb.BlockSize)))
			if err != nil {
				return fmt.Errorf("error al serializar bloque de contenido: %v", err)
			}
			return nil
		}
	}

	// Si el bloque de apuntadores no existe, crearlo
	if *pointer == -1 {
		newBlockap := NewBlockApuntador()
		err := newBlockap.Serialize(path, int64(sb.FirstFreeBlock))
		if err != nil {
			return fmt.Errorf("error al actualizar el bloque de users.txt: %v", err)
		}

		*pointer = sb.CountBlocks
		BlockTypes[int(sb.CountBlocks)] = "APUNTADOR"
		sb.UpdateBitmapBlock(path)
		sb.CountBlocks++
		sb.FreeBlocks--
		sb.FirstFreeBlock += sb.BlockSize
	}

	// Caso recursivo: manejar un bloque de apuntadores
	newBlockap := &BLOCK_APUNTADOR{}

	err := newBlockap.Deserialize(path, int64(sb.BlockStart+(*pointer*sb.BlockSize)))
	if err != nil {
		return err
	}

	// Iterar sobre los apuntadores del bloque
	for i := 0; i < len(newBlockap.Pointers); i++ {
		if len(*cadena) > 0 {
			// Llamada recursiva para bajar un nivel de profundidad
			err := sb.handleRecursivePointer(path, cadena, &newBlockap.Pointers[i], depth-1)
			if err != nil {
				return err
			}
		}
	}

	fmt.Println("SE acabaron los bloques ahora a serializar")
	// Serializar el bloque de apuntadores actualizado
	err = newBlockap.Serialize(path, int64(sb.BlockStart+(*pointer*sb.BlockSize)))
	if err != nil {
		return fmt.Errorf("error al serializar bloque de apuntadores: %v", err)
	}

	return nil
}

func (sb *SUPERBLOCK) ReadFile(path string, fileParts []string, inodoId int32, cadenaFinal *string) error {
	fmt.Println("Empieza READFILE\n")
	rootInodo := &INODO{}
	err := rootInodo.Deserialize(path, int64(sb.InodeStart+inodoId*sb.InodeSize))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de carpeta raíz: %v", err)
	}

	rootBlock := &BLOCK_CARPETA{}
	var StartBlockfile int32
	var usersInodoNum int32 = -1
	encontrado := false

	for i := 0; i < len(rootInodo.Blocks); i++ {
		if rootInodo.Blocks[i] == -1 || encontrado {
			break
		}
		fmt.Println(rootInodo.Blocks)
		fmt.Println("Bloque INodo", rootInodo.Blocks[i])
		StartBlockfile = sb.BlockStart + (rootInodo.Blocks[i] * sb.BlockSize)

		switch i {
		case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:

			if string(rootInodo.Tipo[:]) == "0" {
				err = rootBlock.Deserialize(path, int64(StartBlockfile))
				if err != nil {
					return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
				}

				for _, content := range rootBlock.Content {
					filename := strings.TrimRight(string(content.Name[:]), "\x00")
					fmt.Println("FILEPARTS", fileParts, filename)
					if filename == fileParts[0] {
						usersInodoNum = content.Inodo
						fmt.Println("USERINODO", usersInodoNum)
						fileParts = fileParts[1:]
						err = sb.ReadFile(path, fileParts, usersInodoNum, cadenaFinal)
						if err != nil {
							return err
						}
						encontrado = true
						break
					}
				}
			} else if string(rootInodo.Tipo[:]) == "1" {
				// Final de la ruta "a.txt"
				fmt.Println("ENTRO A ESTE IF DE TIPO 1 CON", i)
				// Calcular el offset del bloque actual
				blockOffset := sb.BlockStart + (rootInodo.Blocks[i] * sb.BlockSize)

				// Leer el bloque de datos
				usersBlock, err := ReadCont_Archivo(path, int64(blockOffset))
				if err != nil {
					return fmt.Errorf("error al leer el bloque %d de users.txt: %v", i, err)
				}
				// Convertir el contenido del bloque a string y agregarlo a usersText
				*cadenaFinal += strings.TrimRight(string(usersBlock.Content[:]), "\x00")
				fmt.Println(*cadenaFinal)
			}
		case 12:
			err := sb.ReadIndirectBlock(path, rootInodo.Blocks[i], 1, cadenaFinal)
			if err != nil {
				return err
			}
		case 13:
			err := sb.ReadIndirectBlock(path, rootInodo.Blocks[i], 2, cadenaFinal)
			if err != nil {
				return err
			}
		case 14:
			err := sb.ReadIndirectBlock(path, rootInodo.Blocks[i], 3, cadenaFinal)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sb *SUPERBLOCK) ReadIndirectBlock(path string, blockPointer int32, depth int, cadenaFinal *string) error {
	if depth == 0 {
		// Caso base: leer un bloque de contenido de archivo
		blockOffset := sb.BlockStart + (blockPointer * sb.BlockSize)
		usersBlock, err := ReadCont_Archivo(path, int64(blockOffset))
		if err != nil {
			return fmt.Errorf("error al leer el bloque de archivo: %v", err)
		}
		// Convertir el contenido del bloque a string y agregarlo a cadenaFinal
		*cadenaFinal += strings.TrimRight(string(usersBlock.Content[:]), "\x00")
		return nil
	}

	// Caso recursivo: leer un bloque de apuntadores
	blockOffset := sb.BlockStart + (blockPointer * sb.BlockSize)
	blockApuntador := NewBlockApuntador()
	err := blockApuntador.Deserialize(path, int64(blockOffset))
	if err != nil {
		return fmt.Errorf("error al leer el bloque de apuntadores: %v", err)
	}

	// Recorrer los apuntadores del bloque de apuntadores
	for i := 0; i < len(blockApuntador.Pointers); i++ {
		if blockApuntador.Pointers[i] == -1 {
			break
		}
		// Llamada recursiva con un nivel de profundidad menor
		err := sb.ReadIndirectBlock(path, blockApuntador.Pointers[i], depth-1, cadenaFinal)
		if err != nil {
			return err
		}
	}

	return nil
}

// createFolderInInode crea una carpeta en un inodo específico
func (sb *SUPERBLOCK) createFolderInInode(path string, inodeIndex int32, parentsDir []string, destDir string, createParents bool) (bool, error) {
	// Crear un nuevo inodo
	inode := &INODO{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.InodeStart+(inodeIndex*sb.InodeSize)))
	if err != nil {
		return false, err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.Tipo[0] == '1' {
		return false, nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	var encontrado = false
	for blockIdx := 0; blockIdx < len(inode.Blocks); blockIdx++ {
		blockIndex := inode.Blocks[blockIdx]

		// Si el bloque no existe (es decir, tiene un valor de -1), creamos un nuevo bloque de carpeta
		fmt.Println("bloques", inode.Blocks)
		if blockIndex == -1 && !encontrado && createParents {
			fmt.Println("CREANDO BLOQUE CARPETA EN INODO")
			newBlock := &BLOCK_CARPETA{
				Content: [4]BCONTENT{
					{Name: [12]byte{'.'}, Inodo: inodeIndex},
					{Name: [12]byte{'.', '.'}, Inodo: inodeIndex}, // El segundo es el padre
					{Name: [12]byte{'-'}, Inodo: -1},
					{Name: [12]byte{'-'}, Inodo: -1},
				},
			}

			// Asignar el nuevo bloque al inodo
			inode.Blocks[blockIdx] = sb.CountBlocks
			BlockTypes[int(sb.CountBlocks)] = "CARPETA"

			// Serializar el nuevo bloque en el archivo
			err := newBlock.Serialize(path, int64(sb.FirstFreeBlock))
			if err != nil {
				return false, fmt.Errorf("Error al serializar nuevo bloque de carpeta: %v", err)
			}

			// Actualizar el bitmap de bloques y el superbloque
			sb.UpdateBitmapBlock(path)
			sb.CountBlocks++
			sb.FreeBlocks--
			sb.FirstFreeBlock += sb.BlockSize

			// Actualizar el inodo con el nuevo bloque
			err = inode.WriteToFile(path, int(sb.InodeStart+(inodeIndex*sb.InodeSize)))
			if err != nil {
				return false, fmt.Errorf("Error al actualizar el inodo con nuevo bloque: %v", err)
			}

			blockIndex = inode.Blocks[blockIdx]
		}

		fmt.Println("ENTRO", parentsDir, destDir)

		// Crear un nuevo bloque de carpeta
		block := &BLOCK_CARPETA{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return false, err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.Content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.Content[indexContent]
			// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.Inodo == -1 {
					// Si `createParents` es true, crear la carpeta padre
					if createParents {
						// Crear la carpeta padre
						fmt.Println("Creando la carpeta padre:", parentsDir[0])
						_, err := sb.createFolderInInode(path, inodeIndex, []string{}, parentsDir[0], createParents)
						if err != nil {
							return false, err
						}
						// **Deserializar de nuevo el inodo y el bloque después de la llamada recursiva**
						err = inode.Deserialize(path, int64(sb.InodeStart+(inodeIndex*sb.InodeSize)))
						if err != nil {
							return false, err
						}

						err = block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
						if err != nil {
							return false, err
						}

						content = block.Content[indexContent]

					} else {
						// Si `createParents` es false, devolver un error o detenerse
						return false, fmt.Errorf("La carpeta padre '%s' no existe y no se permite su creación automatica", parentsDir[0])
					}
				}

				// Obtenemos la carpeta padre más cercana

				parentDir, err := util.First(parentsDir)
				if err != nil {
					return false, err
				}
				fmt.Println("Despues de if", parentsDir, parentDir)
				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				fmt.Println("COMPARANDO", contentName, parentDirName)
				if strings.EqualFold(contentName, parentDirName) {
					fmt.Println("Coincide", parentsDir)
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					encontrado, err = sb.createFolderInInode(path, content.Inodo, util.RemoveElement(parentsDir, 0), destDir, createParents)
					fmt.Println("Salio de llamada revursiva")
					if err != nil {
						return false, err
					}
					return true, nil
				}
			} else {
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				fmt.Println("Llegamos a destino")
				content := block.Content[indexContent]
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				fmt.Println(contentName, content.Inodo)
				if content.Inodo != -1 {
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.Name[:], destDir)
				content.Inodo = sb.CountInodos

				// Actualizar el bloque
				block.Content[indexContent] = content

				// Serializar el bloque
				err = block.Serialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
				if err != nil {
					return false, err
				}

				// Crear el inodo de la carpeta
				folderInode := &INODO{
					Uid:        int32(util.UserOnline.UId),
					Gid:        int32(util.UserOnline.GId),
					Size:       0,
					Lastime:    float32(time.Now().Unix()),
					CreateTime: float32(time.Now().Unix()),
					ModifyTime: float32(time.Now().Unix()),
					Blocks:     [15]int32{sb.CountBlocks, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					Tipo:       [1]byte{'0'},
					Perm:       [3]byte{'6', '6', '4'},
				}

				// Serializar el inodo de la carpeta
				err = folderInode.WriteToFile(path, int(sb.FirstFreeInode))
				if err != nil {
					return false, err
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInodo(path)
				if err != nil {
					return false, err
				}

				// Actualizar el superbloque
				sb.CountInodos++
				sb.FreeInodos--
				sb.FirstFreeInode += sb.InodeSize

				// Crear el bloque de la carpeta
				folderBlock := &BLOCK_CARPETA{
					Content: [4]BCONTENT{
						{Name: [12]byte{'.'}, Inodo: content.Inodo},
						{Name: [12]byte{'.', '.'}, Inodo: inodeIndex},
						{Name: [12]byte{'-'}, Inodo: -1},
						{Name: [12]byte{'-'}, Inodo: -1},
					},
				}

				// Serializar el bloque de la carpeta
				err = folderBlock.Serialize(path, int64(sb.FirstFreeBlock))
				if err != nil {
					return false, err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return false, err
				}

				// Actualizar el superbloque
				BlockTypes[int(sb.CountBlocks)] = "CARPETA"
				sb.CountBlocks++
				sb.FreeBlocks--
				sb.FirstFreeBlock += sb.BlockSize
				fmt.Println("Guardado")
				return true, nil
			}
		}

	}
	return false, nil
}

// createFolderinode crea una carpeta en un inodo específico
func (sb *SUPERBLOCK) createFileInInode(path string, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string, createParents bool) (bool, error) {
	// Crear un nuevo inodo
	inode := &INODO{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.InodeStart+(inodeIndex*sb.InodeSize)))
	if err != nil {
		return false, err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.Tipo[0] == '1' {
		return false, nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	var encontrado = false
	for blockIdx := 0; blockIdx < len(inode.Blocks); blockIdx++ {
		blockIndex := inode.Blocks[blockIdx]
		// Si el bloque no existe, salir
		if blockIndex == -1 && !encontrado {
			fmt.Println("CREANDO BLOQUE CARPETA EN INODO")
			newBlock := &BLOCK_CARPETA{
				Content: [4]BCONTENT{
					{Name: [12]byte{'.'}, Inodo: inodeIndex},
					{Name: [12]byte{'.', '.'}, Inodo: inodeIndex}, // El segundo es el padre
					{Name: [12]byte{'-'}, Inodo: -1},
					{Name: [12]byte{'-'}, Inodo: -1},
				},
			}

			// Asignar el nuevo bloque al inodo
			inode.Blocks[blockIdx] = sb.CountBlocks
			BlockTypes[int(sb.CountBlocks)] = "CARPETA"

			// Serializar el nuevo bloque en el archivo
			err := newBlock.Serialize(path, int64(sb.FirstFreeBlock))
			if err != nil {
				return false, fmt.Errorf("Error al serializar nuevo bloque de carpeta: %v", err)
			}

			// Actualizar el bitmap de bloques y el superbloque
			sb.UpdateBitmapBlock(path)
			sb.CountBlocks++
			sb.FreeBlocks--
			sb.FirstFreeBlock += sb.BlockSize

			// Actualizar el inodo con el nuevo bloque
			err = inode.WriteToFile(path, int(sb.InodeStart+(inodeIndex*sb.InodeSize)))
			if err != nil {
				return false, fmt.Errorf("Error al actualizar el inodo con nuevo bloque: %v", err)
			}

			blockIndex = inode.Blocks[blockIdx]
		}

		fmt.Println("ENTRO FILE", parentsDir, destFile)
		// Crear un nuevo bloque de carpeta
		block := &BLOCK_CARPETA{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return false, err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.Content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.Content[indexContent]

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				//fmt.Println("---------ESTOY  VISITANDO--------")

				// Si el contenido está vacío, salir
				if content.Inodo == -1 {
					if createParents {
						fmt.Println("DIRECTORIOS PADRES ANTES", parentsDir)
						parentsDirFolders := append([]string(nil), parentsDir[:len(parentsDir)-1]...)
						destFolder := parentsDir[len(parentsDir)-1]
						fmt.Println("-------------", parentsDirFolders, destFolder)
						sb.createFolderInInode(path, inodeIndex, parentsDirFolders, destFolder, true)
						// **Deserializar de nuevo el inodo y el bloque después de la llamada recursiva**
						err = inode.Deserialize(path, int64(sb.InodeStart+(inodeIndex*sb.InodeSize)))
						if err != nil {
							return false, err
						}

						err = block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
						if err != nil {
							return false, err
						}

						content = block.Content[indexContent]
						fmt.Println("DIRECTORIOS PADRES DESPUES", parentsDir)
					} else {
						break
					}
				}

				// Obtenemos la carpeta padre más cercana
				parentDir, err := util.First(parentsDir)
				if err != nil {
					return false, err
				}
				fmt.Println("Despues de if File", parentsDir, parentDir)

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				// Convertir parentDir a string y eliminar los caracteres nulos
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				fmt.Println("COmparando", contentName, parentDirName)
				if strings.EqualFold(contentName, parentDirName) {
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					encontrado, err = sb.createFileInInode(path, content.Inodo, util.RemoveElement(parentsDir, 0), destFile, fileSize, fileContent, createParents)
					if err != nil {
						return false, err
					}
					return encontrado, nil
				}
			} else {
				//fmt.Println("---------ESTOY  CREANDO--------")

				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				fmt.Println("Else", contentName)
				fmt.Printf("inode.Blocks: %v\n", inode.Blocks)
				if content.Inodo != -1 {
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.Name[:], []byte(destFile))
				content.Inodo = sb.CountInodos

				// Actualizar el bloque
				block.Content[indexContent] = content

				// Serializar el bloque
				err = block.Serialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
				if err != nil {
					return false, err
				}

				// Crear el inodo del archivo
				fileInode := &INODO{
					Uid:        int32(util.UserOnline.UId),
					Gid:        int32(util.UserOnline.GId),
					Size:       int32(fileSize),
					Lastime:    float32(time.Now().Unix()),
					CreateTime: float32(time.Now().Unix()),
					ModifyTime: float32(time.Now().Unix()),
					Blocks:     [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					Tipo:       [1]byte{'1'},
					Perm:       [3]byte{'6', '6', '4'},
				}

				// Crear el bloques del archivo
				for i := 0; i < len(fileContent); i++ {
					// Actualizamos el inodo del archivo
					fileInode.Blocks[i] = sb.CountBlocks

					// Creamos el bloque del archivo
					fileBlock := &BLOCK_CONTARCHIVO{
						Content: [64]byte{},
					}
					// Copiamos el texto de usuarios en el bloque
					copy(fileBlock.Content[:], fileContent[i])

					// Serializar el bloque de users.txt
					err = fileBlock.Serialize(path, int64(sb.FirstFreeBlock))
					if err != nil {
						return false, err
					}

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(path)
					if err != nil {
						return false, err
					}

					// Actualizamos el superbloque
					BlockTypes[int(sb.CountBlocks)] = "CONTARCHIVO"
					sb.CountBlocks++
					sb.FreeBlocks--
					sb.FirstFreeBlock += sb.BlockSize
				}

				// Serializar el inodo de la carpeta
				err = fileInode.WriteToFile(path, int(sb.FirstFreeInode))
				if err != nil {
					return false, err
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInodo(path)
				if err != nil {
					return false, err
				}

				// Actualizar el superbloque
				sb.CountInodos++
				sb.FreeInodos--
				sb.FirstFreeInode += sb.InodeSize

				return true, nil
			}
		}

	}
	if !encontrado {
		return false, errors.New("Archivo no encontrado")
	}
	return false, nil
}

// Funcion que muestra lo que hay en el directorio actual
func (sb *SUPERBLOCK) ShowFiles(path string, fileParts []string, inodoId int32, files *[]SHOWFILES) error {
	rootInodo := &INODO{}
	err := rootInodo.Deserialize(path, int64(sb.InodeStart+inodoId*sb.InodeSize))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de carpeta raíz: %v", err)
	}

	rootBlock := &BLOCK_CARPETA{}
	var StartBlockfile int32
	var usersInodoNum int32 = -1
	encontrado := false

	// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
	fmt.Println(len(fileParts), fileParts)
	if len(fileParts) != 0 {
		for i := 0; i < len(rootInodo.Blocks); i++ {
			if rootInodo.Blocks[i] == -1 || encontrado {
				break
			}
			fmt.Println("Bloque INodo", rootInodo.Blocks[i])
			StartBlockfile = sb.BlockStart + (rootInodo.Blocks[i] * sb.BlockSize)

			switch i {
			case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:

				if string(rootInodo.Tipo[:]) == "0" {
					err = rootBlock.Deserialize(path, int64(StartBlockfile))
					if err != nil {
						return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
					}

					for _, content := range rootBlock.Content {
						filename := strings.TrimRight(string(content.Name[:]), "\x00")
						fmt.Println("FILEPARTS", fileParts, filename)
						if filename == fileParts[0] {
							usersInodoNum = content.Inodo
							fmt.Println("USERINODO", usersInodoNum)
							fileParts = fileParts[1:]
							err = sb.ShowFiles(path, fileParts, usersInodoNum, files)
							if err != nil {
								return err
							}
							encontrado = true
							break
						}
					}
				}
			}
		}
	} else {

		for i := 0; i < len(rootInodo.Blocks); i++ {
			if rootInodo.Blocks[i] == -1 {
				break
			}

			StartBlockfile = sb.BlockStart + (rootInodo.Blocks[i] * sb.BlockSize)

			switch i {
			case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
				err = rootBlock.Deserialize(path, int64(StartBlockfile))
				if err != nil {
					return fmt.Errorf("error al leer el bloque de carpeta raíz: %v", err)
				}

				for _, content := range rootBlock.Content {
					filename := strings.TrimRight(string(content.Name[:]), "\x00")

					if filename == "-" || filename == "." || filename == ".." {
						continue
					}

					if content.Inodo == -1 {
						continue
					}

					if strings.HasSuffix(filename, ".txt") {
						*files = append(*files, SHOWFILES{Name: filename, Type: "archivo"})
					} else {
						*files = append(*files, SHOWFILES{Name: filename, Type: "carpeta"})
					}
				}
			}
		}
	}

	return nil
}
