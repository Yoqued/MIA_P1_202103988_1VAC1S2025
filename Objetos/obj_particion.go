package objs

type PARTICION struct {
	Name        [16]byte // El nombre de la particion
	Status      [1]byte  // Indica si esta montada o no
	Tipo        [1]byte  // Indica si es primaria o extendida
	Fit         [1]byte  // Tipo de Ajuste de la particion
	Start       [4]byte  // En que byte inicia la particion
	Size        [4]byte  // Contiene el tamaño de la particion
	Correlative [4]byte  // Indica el correlativo de la particion
	Id          [4]byte  // Indica el ID de la particion
}

type PARTICION_CONV struct {
	Name        string
	Status      string
	Tipo        string
	Fit         string
	Start       int
	Size        float32
	Correlative string
	Id          string
}

func (p *PARTICION) Clear() {
	copy(p.Name[:], make([]byte, len(p.Name)))
	copy(p.Status[:], make([]byte, len(p.Status)))
	copy(p.Tipo[:], make([]byte, len(p.Tipo)))
	copy(p.Fit[:], make([]byte, len(p.Fit)))
	copy(p.Start[:], make([]byte, len(p.Start)))
	copy(p.Size[:], make([]byte, len(p.Size)))
	copy(p.Correlative[:], make([]byte, len(p.Correlative)))
	copy(p.Id[:], make([]byte, len(p.Id)))
}
