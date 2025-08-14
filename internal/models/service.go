package models

// Service representa los datos de un servicio.
type Service struct {
	CodServ   string   `json:"codserv"`
	CodInst   int      `json:"codinst"`
	Descrip   string   `json:"descrip"`
	Descrip2  *string  `json:"descrip2"`
	Descrip3  *string  `json:"descrip3"`
	Clase     *string  `json:"clase"`
	Activo    int      `json:"activo"`
	Unidad    *string  `json:"unidad"`
	Precio1   float64  `json:"precio1"`
	PrecioI1  float64  `json:"precioi1"`
	PrecioR1  float64  `json:"precior1"`
	Precio2   float64  `json:"precio2"`
	PrecioI2  float64  `json:"precioi2"`
	PrecioR2  float64  `json:"precior2"`
	Precio3   float64  `json:"precio3"`
	PrecioI3  float64  `json:"precioi3"`
	PrecioR3  float64  `json:"precior3"`
	Costo     float64  `json:"costo"`
	EsExento  int      `json:"esexento"`
	EsReten   int      `json:"esreten"`
	EsPorCost int      `json:"esporcost"`
	UsaServ   int      `json:"usaserv"`
	Comision  float64  `json:"comision"`
	EsPorComi int      `json:"esporcomi"`
	FechaUV   *string  `json:"fechauv"`
	FechaUC   *string  `json:"fechauc"`
	EsImport  int      `json:"esimport"`
	EsVenta   int      `json:"esventa"`
	EsCompra  int      `json:"escompra"`
	ID        int      `json:"id"`
}