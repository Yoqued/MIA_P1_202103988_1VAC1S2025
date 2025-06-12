package estructuras

import (
	util "MIA_P1_202103988_1VAC1S2025/Util"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type SUPERBLOCK struct {
	Filesystem       int32   // Numero que identifica el sistema
	CountInodos      int32   // Cantidad de inodos
	CountBlocks      int32   // Cantidad de bloques
	FreeBlocks       int32   // Cantidad de bloques libres
	FreeInodos       int32   // Cantidad de inodos libres
	MontTime         float32 // Ultima fecha que fue montado el sistema
	UMontTime        float32 // Ultima fecha que fue desmontado el sistema
	CountMount       int32   // Cuantas veces se ha montado el sistema
	Magic            int32   // Valor que identifica al sistema
	InodeSize        int32   // Tamaño del inodo
	BlockSize        int32   // Tamaño del Bloque
	FirstFreeInode   int32   // Primer Inodo libre
	FirstFreeBlock   int32   // Primer bloque libre
	InodeBitmapStart int32   // Inicio del bitmap de inodos
	BlockBitmapStart int32   // Inicio del bitmap de bloques
	InodeStart       int32   // Inicio de la tabla de inodos
	BlockStart       int32   // Inicio de la tabla de bloques
}

// Escribe la estructura del SuperBloque
func (sb *SUPERBLOCK) WriteToFile(path string, start int) error {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Mueve el puntero al byte correspondiente
	startoff := int64(start)
	_, err = file.Seek(startoff, 0)
	if err != nil {
		return fmt.Errorf("error al mover el puntero del archivo: %v", err)
	}

	err = binary.Write(file, binary.LittleEndian, sb)
	if err != nil {
		return fmt.Errorf("error al escribir la estructura Student: %v", err)
	}

	return nil
}

func (sb *SUPERBLOCK) CreateUsersFile(path string) error {
	// ----------- Creamos / -----------
	// Creamos el inodo raíz
	rootInode := &INODO{
		Uid:        1,
		Gid:        1,
		Size:       0,
		Lastime:    float32(time.Now().Unix()),
		CreateTime: float32(time.Now().Unix()),
		ModifyTime: float32(time.Now().Unix()),
		Blocks:     [15]int32{sb.CountBlocks, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Tipo:       [1]byte{'0'},
		Perm:       [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err := rootInode.WriteToFile(path, int(sb.FirstFreeInode))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInodo(path)
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.CountInodos++
	sb.FreeInodos--
	sb.FirstFreeInode += sb.InodeSize

	// Creamos el bloque del Inodo Raíz
	rootBlock := &BLOCK_CARPETA{
		Content: [4]BCONTENT{
			{Name: [12]byte{'.'}, Inodo: 0},
			{Name: [12]byte{'.', '.'}, Inodo: 0},
			{Name: [12]byte{'-'}, Inodo: -1},
			{Name: [12]byte{'-'}, Inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.FirstFreeBlock))
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	BlockTypes[int(sb.CountBlocks)] = "CARPETA"
	sb.CountBlocks++
	sb.FreeBlocks--
	sb.FirstFreeBlock += sb.BlockSize

	// Verificar el inodo raíz
	fmt.Println("\nInodo Raíz:")
	rootInode.Print()

	// Verificar el bloque de carpeta raíz
	fmt.Println("\nBloque de Carpeta Raíz:")
	rootBlock.Print()

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Deserializar el inodo raíz
	err = rootInode.Deserialize(path, int64(sb.InodeStart)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Actualizamos el inodo raíz
	rootInode.Lastime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = rootInode.WriteToFile(path, int(sb.InodeStart)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Deserializar el bloque de carpeta raíz
	err = rootBlock.Deserialize(path, int64(sb.BlockStart)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Actualizamos el bloque de carpeta raíz
	rootBlock.Content[2] = BCONTENT{Name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, Inodo: sb.CountInodos}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.BlockStart+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Creamos el inodo users.txt
	usersInode := &INODO{
		Uid:        1,
		Gid:        1,
		Size:       int32(len(usersText)),
		Lastime:    float32(time.Now().Unix()),
		CreateTime: float32(time.Now().Unix()),
		ModifyTime: float32(time.Now().Unix()),
		Blocks:     [15]int32{sb.CountBlocks, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Tipo:       [1]byte{'1'},
		Perm:       [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInodo(path)
	if err != nil {
		return err
	}

	// Serializar el inodo users.txt
	err = usersInode.WriteToFile(path, int(sb.FirstFreeInode))
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.CountInodos++
	sb.FreeInodos--
	sb.FirstFreeInode += sb.InodeSize

	// Creamos el bloque de users.txt
	usersBlock := &BLOCK_CONTARCHIVO{
		Content: [64]byte{},
	}
	// Copiamos el texto de usuarios en el bloque
	copy(usersBlock.Content[:], usersText)

	// Serializar el bloque de users.txt
	BlockTypes[int(sb.CountBlocks)] = "CONTARCHIVO"
	err = usersBlock.Serialize(path, int64(sb.FirstFreeBlock))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.CountBlocks++
	sb.FreeBlocks--
	sb.FirstFreeBlock += sb.BlockSize

	// Verificar el inodo raíz
	fmt.Println("\nInodo Raíz Actualizado:")
	rootInode.Print()

	// Verificar el bloque de carpeta raíz
	fmt.Println("\nBloque de Carpeta Raíz Actualizado:")
	rootBlock.Print()

	// Verificar el inodo users.txt
	fmt.Println("\nInodo users.txt:")
	usersInode.Print()

	// Verificar el bloque de users.txt
	fmt.Println("\nBloque de users.txt:")
	usersBlock.Print()

	return nil
}

func ReadSuperBlock(path string, offset int64) (*SUPERBLOCK, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Mover el puntero de lectura al offset especificado
	_, err = file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}

	sb := &SUPERBLOCK{}
	if err := binary.Read(file, binary.LittleEndian, sb); err != nil {
		return nil, err
	}

	return sb, nil
}

func (sb *SUPERBLOCK) Print() {
	// Convertir el tiempo de montaje a una fecha
	mountTime := time.Unix(int64(sb.MontTime), 0)
	// Convertir el tiempo de desmontaje a una fecha
	unmountTime := time.Unix(int64(sb.UMontTime), 0)

	fmt.Printf("Filesystem Type: %d\n", sb.Filesystem)
	fmt.Printf("Inodes Count: %d\n", sb.CountInodos)
	fmt.Printf("Blocks Count: %d\n", sb.CountBlocks)
	fmt.Printf("Free Inodes Count: %d\n", sb.FreeInodos)
	fmt.Printf("Free Blocks Count: %d\n", sb.FreeBlocks)
	fmt.Printf("Mount Time: %s\n", mountTime.Format(time.RFC3339))
	fmt.Printf("Unmount Time: %s\n", unmountTime.Format(time.RFC3339))
	fmt.Printf("Mount Count: %d\n", sb.CountMount)
	fmt.Printf("Magic: %d\n", sb.Magic)
	fmt.Printf("Inode Size: %d\n", sb.InodeSize)
	fmt.Printf("Block Size: %d\n", sb.BlockSize)
	fmt.Printf("First Inode: %d\n", sb.FirstFreeInode)
	fmt.Printf("First Block: %d\n", sb.FirstFreeBlock)
	fmt.Printf("Bitmap Inode Start: %d\n", sb.BlockBitmapStart)
	fmt.Printf("Bitmap Block Start: %d\n", sb.BlockBitmapStart)
	fmt.Printf("Inode Start: %d\n", sb.InodeStart)
	fmt.Printf("Block Start: %d\n", sb.BlockStart)
}

func (sb *SUPERBLOCK) IncreaseMount() {
	sb.CountMount = sb.CountMount + 1
}

// CreateFolder crea una carpeta en el sistema de archivos
func (sb *SUPERBLOCK) CreateFolder(path string, parentsDir []string, destDir string, createParents bool) error {
	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		_, err := sb.createFolderInInode(path, 0, parentsDir, destDir, createParents)
		return err
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.CountInodos; i++ {
		fmt.Println("ITERACION", i, "INODOS:", sb.CountInodos)
		create, err := sb.createFolderInInode(path, i, parentsDir, destDir, createParents)
		if err != nil {
			return err
		}
		if create {
			break
		}
	}

	return nil
}

// CreateFile crea un archivo en el sistema de archivos
func (sb *SUPERBLOCK) CreateFile(path string, parentsDir []string, destFile string, size int, cont []string, createParents bool) error {

	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		_, err := sb.createFileInInode(path, 0, parentsDir, destFile, size, cont, createParents)
		return err
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.CountInodos; i++ {
		fmt.Println("---Inodo-----", i)
		create, err := sb.createFileInInode(path, i, parentsDir, destFile, size, cont, createParents)
		if err != nil {
			return err
		}
		if create {
			break
		}
	}

	return nil
}

func (sb *SUPERBLOCK) ValidateSuperBlock() error {
	// Validar que el Filesystem sea 2 o 3
	if sb.Filesystem != 2 && sb.Filesystem != 3 {
		return errors.New("el valor de Filesystem debe ser 2 o 3")
	}

	// Validar que CountInodos y CountBlocks sean mayores o iguales a 0
	fmt.Println(sb.CountInodos, sb.Magic, sb.InodeBitmapStart, sb.InodeSize, sb.InodeStart)
	if sb.CountInodos < 0 && sb.CountBlocks < 0 || sb.InodeStart < 0 || sb.BlockStart < 0 {
		return errors.New("el valor de CountInodos debe ser mayor o igual a 0")
	}

	if sb.CountBlocks < 0 {
		return errors.New("el valor de CountBlocks debe ser mayor o igual a 0")
	}

	if sb.Magic != 61267 {
		return errors.New("Error de superbloque")
	}

	// Validar otros parámetros si es necesario (ejemplo: Magic, FirstFreeInode, etc.)

	return nil
}

// Operaciones para el sistema de archivos ext3

// Crear users.txt en nuestro sistema de archivos
func (sb *SUPERBLOCK) CreateUsersFileExt3(path string, journauling_start int64) error {
	// ----------- Creamos / -----------
	// Crear Journal
	journal := &JOURNAL{
		J_count: sb.CountInodos,
		J_content: Information{
			I_operation: [10]byte{'m', 'k', 'd', 'i', 'r'},
			I_path:      [32]byte{'/'},
			I_content:   [64]byte{},
			I_date:      float32(time.Now().Unix()),
		},
	}

	// Serializar el journal
	err := journal.Serialize(path, journauling_start)
	if err != nil {
		return err
	}

	// Creamos el inodo raíz
	rootInode := &INODO{
		Uid:        1,
		Gid:        1,
		Size:       0,
		Lastime:    float32(time.Now().Unix()),
		CreateTime: float32(time.Now().Unix()),
		ModifyTime: float32(time.Now().Unix()),
		Blocks:     [15]int32{sb.CountBlocks, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Tipo:       [1]byte{'0'},
		Perm:       [3]byte{'7', '7', '7'},
	}

	// Serializar el inodo raíz
	err = rootInode.WriteToFile(path, int(sb.FirstFreeInode))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInodo(path)
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.CountInodos++
	sb.FreeInodos--
	sb.FirstFreeInode += sb.InodeSize

	// Creamos el bloque del Inodo Raíz
	rootBlock := &BLOCK_CARPETA{
		Content: [4]BCONTENT{
			{Name: [12]byte{'.'}, Inodo: 0},
			{Name: [12]byte{'.', '.'}, Inodo: 0},
			{Name: [12]byte{'-'}, Inodo: -1},
			{Name: [12]byte{'-'}, Inodo: -1},
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.FirstFreeBlock))
	if err != nil {
		return err
	}

	// Actualizar el superbloque
	sb.CountBlocks++
	sb.FreeBlocks--
	sb.FirstFreeBlock += sb.BlockSize

	// ----------- Creamos /users.txt -----------
	usersText := "1,G,root\n1,U,root,root,123\n"

	// Crear Journal
	journalFile := &JOURNAL{
		J_count: sb.CountInodos,
		J_content: Information{
			I_operation: [10]byte{'m', 'k', 'f', 'i', 'l', 'e'},
			I_path:      [32]byte{'/', 'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'},
			I_content:   [64]byte{},
			I_date:      float32(time.Now().Unix()),
		},
	}
	// Copiamos el texto de usuarios en el journal
	copy(journalFile.J_content.I_content[:], usersText)

	// Serializar el journal
	err = journalFile.Serialize(path, journauling_start)
	if err != nil {
		return err
	}

	// Deserializar el inodo raíz
	err = rootInode.Deserialize(path, int64(sb.InodeStart+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Actualizamos el inodo raíz
	rootInode.Lastime = float32(time.Now().Unix())

	// Serializar el inodo raíz
	err = rootInode.WriteToFile(path, int(sb.InodeStart+0)) // 0 porque es el inodo raíz
	if err != nil {
		return err
	}

	// Deserializar el bloque de carpeta raíz
	err = rootBlock.Deserialize(path, int64(sb.BlockStart+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Actualizamos el bloque de carpeta raíz
	rootBlock.Content[2] = BCONTENT{Name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, Inodo: sb.CountInodos}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Serialize(path, int64(sb.BlockStart+0)) // 0 porque es el bloque de carpeta raíz
	if err != nil {
		return err
	}

	// Creamos el inodo users.txt
	usersInode := &INODO{
		Uid:        1,
		Gid:        1,
		Size:       int32(len(usersText)),
		Lastime:    float32(time.Now().Unix()),
		CreateTime: float32(time.Now().Unix()),
		ModifyTime: float32(time.Now().Unix()),
		Blocks:     [15]int32{sb.CountBlocks, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		Tipo:       [1]byte{'1'},
		Perm:       [3]byte{'7', '7', '7'},
	}

	// Actualizar el bitmap de inodos
	err = sb.UpdateBitmapInodo(path)
	if err != nil {
		return err
	}

	// Serializar el inodo users.txt
	err = usersInode.WriteToFile(path, int(sb.FirstFreeInode))
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.CountInodos++
	sb.FreeInodos--
	sb.FirstFreeInode += sb.InodeSize

	// Creamos el bloque de users.txt
	usersBlock := &BLOCK_CONTARCHIVO{
		Content: [64]byte{},
	}
	// Copiamos el texto de usuarios en el bloque
	copy(usersBlock.Content[:], usersText)

	// Serializar el bloque de users.txt
	err = usersBlock.Serialize(path, int64(sb.FirstFreeBlock))
	if err != nil {
		return err
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(path)
	if err != nil {
		return err
	}

	// Actualizamos el superbloque
	sb.CountBlocks++
	sb.FreeBlocks--
	sb.FirstFreeBlock += sb.BlockSize
	return nil
}

// createFolderInInode crea una carpeta en un inodo específico
func (sb *SUPERBLOCK) createFolderInInodeExt3(path string, inodeIndex int32, parentsDir []string, destDir string) error {
	// Crear un nuevo inodo
	inode := &INODO{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.InodeStart+(inodeIndex*sb.InodeSize)))
	if err != nil {
		return err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.Tipo[0] == '1' {
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.Blocks {
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
					//fmt.Println("---------LA ENCONTRÉ-------")
					// Si son las mismas, entonces entramos al inodo que apunta el bloque
					err := sb.createFolderInInodeExt3(path, content.Inodo, util.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
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
					return err
				}

				// HACER EL JOURNALING

				// Crear el inodo de la carpeta
				folderInode := &INODO{
					Uid:        1,
					Gid:        1,
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
					return err
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInodo(path)
				if err != nil {
					return err
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
					return err
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(path)
				if err != nil {
					return err
				}

				// Actualizar el superbloque
				sb.CountBlocks++
				sb.FreeBlocks--
				sb.FirstFreeBlock += sb.BlockSize

				return nil
			}
		}

	}
	return nil
}

// createFolderinode crea una carpeta en un inodo específico
func (sb *SUPERBLOCK) createFileInInodeExt3(path string, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string) error {
	// Crear un nuevo inodo
	inode := &INODO{}
	// Deserializar el inodo
	err := inode.Deserialize(path, int64(sb.InodeStart+(inodeIndex*sb.InodeSize)))
	if err != nil {
		return err
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.Tipo[0] == '1' {
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.Blocks {
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
				//fmt.Println("---------ESTOY  VISITANDO--------")

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
					err := sb.createFileInInodeExt3(path, content.Inodo, util.RemoveElement(parentsDir, 0), destFile, fileSize, fileContent)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
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
					return err
				}

				// HACER EL JOURNALING

				// Crear el inodo del archivo
				fileInode := &INODO{
					Uid:        1,
					Gid:        1,
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
						return err
					}

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(path)
					if err != nil {
						return err
					}

					// Actualizamos el superbloque
					sb.CountBlocks++
					sb.FreeBlocks--
					sb.FirstFreeBlock += sb.BlockSize
				}

				// Serializar el inodo de la carpeta
				err = fileInode.WriteToFile(path, int(sb.FirstFreeInode))
				if err != nil {
					return err
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInodo(path)
				if err != nil {
					return err
				}

				// Actualizar el superbloque
				sb.CountInodos++
				sb.FreeInodos--
				sb.FirstFreeInode += sb.InodeSize

				return nil
			}
		}

	}
	return nil
}
