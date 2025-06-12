package estructuras

import (
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"errors"
	"fmt"
	"strings"
)

func (sb *SUPERBLOCK) RenameFile(path string, fileParts []string, inodoId int32, name string, newname string) error {
	rootInodo := &INODO{}
	err := rootInodo.Deserialize(path, int64(sb.InodeStart+inodoId*sb.InodeSize))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de carpeta raíz: %v", err)
	}

	var usersInodoNum int32 = -1
	encontrado := false

	// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
	for i, blockIndex := range rootInodo.Blocks {
		if blockIndex == -1 || encontrado {
			break
		}

		block := &BLOCK_CARPETA{}

		err := block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
		if err != nil {
			return err
		}

		for indexContent := 2; indexContent < len(block.Content); indexContent++ {
			content := block.Content[indexContent]

			if len(fileParts) != 0 {

				if content.Inodo == -1 {
					break
				}

				switch i {
				case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
					if string(rootInodo.Tipo[:]) == "0" {
						filename := strings.TrimRight(string(content.Name[:]), "\x00")

						if filename == fileParts[0] {
							usersInodoNum = content.Inodo
							fmt.Println("USERINODO", usersInodoNum)
							fileParts = fileParts[1:]
							err = sb.RenameFile(path, fileParts, usersInodoNum, name, newname)
							if err != nil {
								return err
							}
							encontrado = true
							break
						}
					}
				}

			} else {
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				content.Print()

				if contentName == name {
					for i := range content.Name {
						content.Name[i] = '\x00'
					}
					copy(content.Name[:], []byte(newname))
					fmt.Println("NOmbre cambiado", newname)

					// Actualizar el bloque
					content.Print()
					block.Content[indexContent] = content

					// Serializar el bloque
					err = block.Serialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
					if err != nil {
						return err
					}
					return nil
				}

			}
		}
	}
	if !encontrado {
		return errors.New("Archivo no encontrado")
	}

	return nil
}

// createFolderinode crea una carpeta en un inodo específico
func (sb *SUPERBLOCK) EditFile(path string, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string) (bool, error) {
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
		if blockIndex == -1 {
			break
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
				// Si el contenido está vacío, salir
				if content.Inodo == -1 {
					break
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
					encontrado, err = sb.EditFile(path, content.Inodo, util.RemoveElement(parentsDir, 0), destFile, fileSize, fileContent)
					if err != nil {
						return false, err
					}
					return encontrado, nil
				}
			} else {

				contentName := strings.Trim(string(content.Name[:]), "\x00 ")

				if contentName == destFile {
					// Buscar el inodo al que apunta el bloque (inodo del txt)

					inodeTxt := &INODO{}
					err := inodeTxt.Deserialize(path, int64(sb.InodeStart+(content.Inodo*sb.InodeSize)))
					if err != nil {
						return false, err
					}

					//Recorrer los bloques de texto que contiene el inodeTxt

					for i := 0; i < len(inodeTxt.Blocks); i++ {
						startBlock := sb.BlockStart + (inodeTxt.Blocks[i] * sb.BlockSize)
						fileBlock := &BLOCK_CONTARCHIVO{}
						fileBlock.Deserialize(path, int64(startBlock))

						if inodeTxt.Blocks[i] != -1 {
							for i := range fileBlock.Content {
								fileBlock.Content[i] = '\x00'
							}
							err = fileBlock.Serialize(path, int64(startBlock))
							if err != nil {
								return false, err
							}
						}

						if len(fileContent) > 0 {

							if inodeTxt.Blocks[i] != -1 {

								copy(fileBlock.Content[:], fileContent[i])
								fileContent = fileContent[1:]

								err = fileBlock.Serialize(path, int64(startBlock))
								if err != nil {
									return false, err
								}
							} else {
								// Actualizamos el inodo del archivo
								inodeTxt.Blocks[i] = sb.CountBlocks

								// Creamos el bloque del archivo
								fileBlock := &BLOCK_CONTARCHIVO{
									Content: [64]byte{},
								}
								// Copiamos el texto de usuarios en el bloque
								copy(fileBlock.Content[:], fileContent[i])
								fileContent = fileContent[1:]

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
						}
					}
					return true, nil
				}

			}
		}

	}
	if !encontrado {
		return false, errors.New("Archivo no encontrado")
	}
	return false, nil
}

// Remover Archivo o Carpeta
func (sb *SUPERBLOCK) RemoveFile(path string, fileParts []string, inodoId int32, dest string) error {
	rootInodo := &INODO{}
	err := rootInodo.Deserialize(path, int64(sb.InodeStart+inodoId*sb.InodeSize))
	if err != nil {
		return fmt.Errorf("error al leer el inodo de carpeta raíz: %v", err)
	}

	var usersInodoNum int32 = -1
	encontrado := false

	// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
	for i, blockIndex := range rootInodo.Blocks {
		if blockIndex == -1 || encontrado {
			break
		}

		block := &BLOCK_CARPETA{}

		err := block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
		if err != nil {
			return err
		}

		for indexContent := 2; indexContent < len(block.Content); indexContent++ {
			content := block.Content[indexContent]

			if len(fileParts) != 0 {

				if content.Inodo == -1 {
					break
				}

				switch i {
				case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
					if string(rootInodo.Tipo[:]) == "0" {
						filename := strings.TrimRight(string(content.Name[:]), "\x00")

						if filename == fileParts[0] {
							usersInodoNum = content.Inodo
							fmt.Println("USERINODO", usersInodoNum)
							fileParts = fileParts[1:]
							err = sb.RemoveFile(path, fileParts, usersInodoNum, dest)
							if err != nil {
								return err
							}
							encontrado = true
							break
						}
					}
				}

			} else {
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")

				if contentName == dest {
					for i := range content.Name {
						content.Name[i] = '\x00'
					}
					content.Inodo = -1

					// Actualizar el bloque
					content.Print()
					block.Content[indexContent] = content

					// Serializar el bloque
					err = block.Serialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
					if err != nil {
						return err
					}
					return nil
				}

			}
		}
	}
	if !encontrado {
		return errors.New("Archivo no encontrado")
	}

	return nil
}

// Buscar una carpeta o archivo y obtener el inodo al que apunta
func (sb *SUPERBLOCK) SearchFileInodo(path string, inodoId int32, parentsDir []string, dest string) (int32, error) {
	rootInodo := &INODO{}
	err := rootInodo.Deserialize(path, int64(sb.InodeStart+inodoId*sb.InodeSize))
	if err != nil {
		return 0, fmt.Errorf("error al leer el inodo de carpeta raíz: %v", err)
	}

	var usersInodoNum int32 = -1
	var InodoId int32
	encontrado := false

	// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
	for i, blockIndex := range rootInodo.Blocks {
		if blockIndex == -1 {
			break
		}

		block := &BLOCK_CARPETA{}

		err := block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
		if err != nil {
			return 0, err
		}

		for indexContent := 2; indexContent < len(block.Content); indexContent++ {
			content := block.Content[indexContent]

			if len(parentsDir) != 0 {

				if content.Inodo == -1 {
					continue
				}

				switch i {
				case 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11:
					if string(rootInodo.Tipo[:]) == "0" {
						filename := strings.TrimRight(string(content.Name[:]), "\x00")

						if filename == parentsDir[0] {
							usersInodoNum = content.Inodo
							fmt.Println("USERINODO", usersInodoNum)
							parentsDir = parentsDir[1:]
							InodoId, err = sb.SearchFileInodo(path, usersInodoNum, parentsDir, dest)
							if err != nil {
								return 0, err
							}
							encontrado = true
							break
						}
					}
				}

			} else {
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				content.Print()

				if dest == contentName {
					return content.Inodo, nil
				}

			}
		}
	}
	if !encontrado {
		return 0, errors.New("Archivo no encontrado")
	}

	return InodoId, nil
}

// createFolderInInode crea una carpeta en un inodo específico
func (sb *SUPERBLOCK) CopyFile(path string, inodeIndex int32, parentsDir []string, copyname string, inodoId int32) (bool, error) {
	// Crear un nuevo inodo
	inode := &INODO{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.InodeStart+(inodeIndex*sb.InodeSize)))
	if err != nil {
		return false, err
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	var encontrado = false
	for blockIdx := 0; blockIdx < len(inode.Blocks); blockIdx++ {
		blockIndex := inode.Blocks[blockIdx]

		// Si el bloque no existe (es decir, tiene un valor de -1), creamos un nuevo bloque de carpeta
		fmt.Println("bloques", inode.Blocks)
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
				if content.Inodo == -1 {
					continue
				}

				// Obtenemos la carpeta padre más cercana
				parentDir, err := util.First(parentsDir)
				if err != nil {
					return false, err
				}

				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")

				fmt.Println("COMPARANDO", contentName, parentDirName)
				if strings.EqualFold(contentName, parentDirName) {
					fmt.Println("Coincide", parentsDir)
					encontrado, err = sb.CopyFile(path, content.Inodo, util.RemoveElement(parentsDir, 0), copyname, inodoId)
					if err != nil {
						return false, err
					}
					return true, nil
				}
			} else {
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				content := block.Content[indexContent]
				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				fmt.Println(contentName, content.Inodo)
				if content.Inodo != -1 {
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.Name[:], []byte(copyname))
				content.Inodo = inodoId

				// Actualizar el bloque
				block.Content[indexContent] = content

				// Serializar el bloque
				err = block.Serialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize)))
				if err != nil {
					return false, err
				}

				return true, nil
			}
		}

	}
	return false, nil
}

func (sb *SUPERBLOCK) FindFile(path string, inodoId int32, parentsDir []string, Filtro string, treeFile *string, depth int) error {
	// Crear un nuevo inodo
	inode := &INODO{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.InodeStart+(inodoId*sb.InodeSize)))
	if err != nil {
		return err
	}

	// Convertir el filtro en una expresión regular
	regexFiltro, err := util.ConvertToRegex(Filtro)
	if err != nil {
		return fmt.Errorf("error al convertir el filtro en regex: %v", err)
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for blockIdx := 0; blockIdx < len(inode.Blocks); blockIdx++ {
		blockIndex := inode.Blocks[blockIdx]
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &BLOCK_CARPETA{}

		// Deserializar el bloque
		err := block.Deserialize(path, int64(sb.BlockStart+(blockIndex*sb.BlockSize))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.Content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.Content[indexContent]

			// Sí las carpetas padre no están vacías debereamos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.Inodo == -1 {
					break
				}

				// Obtenemos la carpeta padre más cercana
				parentDir, err := util.First(parentsDir)
				if err != nil {
					return err
				}

				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					err = sb.FindFile(path, content.Inodo, util.RemoveElement(parentsDir, 0), Filtro, treeFile, depth)
					if err != nil {
						return err
					}
					return nil
				}
			} else {

				if content.Inodo == -1 {
					break
				}

				contentName := strings.Trim(string(content.Name[:]), "\x00 ")
				indentation := strings.Repeat("    ", depth)

				// Verificar si el archivo/carpeta no tiene la extensión .txt (para identificar carpetas)
				if !strings.HasSuffix(contentName, ".txt") {
					// Considerar que esto es una carpeta
					*treeFile += indentation + "|_ " + contentName + " (carpeta)\n"
				} else {
					// Si es un archivo .txt, verificar si contiene el filtro
					if regexFiltro.MatchString(contentName) {
						*treeFile += indentation + "|_ " + contentName + " (archivo)\n"
					}
				}

				// llamamos recursivamente para seguir buscando en las subcarpetas.
				if !strings.HasSuffix(contentName, ".txt") {
					err = sb.FindFile(path, content.Inodo, []string{}, Filtro, treeFile, depth+1)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
