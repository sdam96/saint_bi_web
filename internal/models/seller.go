package models

// Seller representa un vendedor.
type Seller struct {
	CodVend   *string  `json:"codvend"`
	Descrip   *string  `json:"descrip"`
	TipoID3   *int     `json:"tipoid3"`
	TipoID    *int     `json:"tipoid"`
	ID3       *string  `json:"id3"`
	DescOrder *string  `json:"descorder"`
	Clase     *string  `json:"clase"`
	Direc1    *string  `json:"direc1"`
	Direc2    *string  `json:"direc2"`
	Telef     *string  `json:"telef"`
	Movil     *string  `json:"movil"`
	Email     *string  `json:"email"`
	FechaUV   *string  `json:"fechauv"`
	FechaUC   *string  `json:"fechauc"`
	EsComiPV  *float64 `json:"escomipv"`
	EsComiTV  *float64 `json:"escomitv"`
	EsComiTC  *float64 `json:"escomitc"`
	EsComiTU  *float64 `json:"escomitu"`
	EsComiDT  *float64 `json:"escomidt"`
	EsComiUT  *float64 `json:"escomiut"`
	EsComiTM  *float64 `json:"escomitm"`
	Activo    *int     `json:"activo"`
	ID        *int     `json:"id"`
}
