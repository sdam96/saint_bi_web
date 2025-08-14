package models

// Server representa los datos de un servidor.
type Server struct {
	CodMeca   string   `json:"codmeca"`
	Descrip   string   `json:"descrip"`
	TipoID3   int      `json:"tipoid3"`
	TipoID    int      `json:"tipoid"`
	ID3       *string  `json:"id3"`
	DescOrder *string  `json:"descorder"`
	Clase     *string  `json:"clase"`
	Activo    int      `json:"activo"`
	Direc1    *string  `json:"direc1"`
	Direc2    *string  `json:"direc2"`
	Telef     *string  `json:"telef"`
	Movil     *string  `json:"movil"`
	Email     *string  `json:"email"`
	DesComi   int      `json:"descomi"`
	Monto     float64  `json:"monto"`
	PorctUtil float64  `json:"porctutil"`
	ZipCode   *string  `json:"zipcode"`
	ID        int      `json:"id"`
}