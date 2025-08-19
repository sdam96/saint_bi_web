package models

// Stock representa la existencia de un producto en un inventario.
type Stock struct {
	CodSucu  *string  `json:"codsucu"`
	CodProd  *string  `json:"codprod"`
	CodUbic  *string  `json:"codubic"`
	PuestoI  *string  `json:"puestoi"`
	Existen  *float64 `json:"existen"`
	ExUnidad *float64 `json:"exunidad"`
	CantPed  *float64 `json:"cantped"`
	UnidPed  *float64 `json:"unidped"`
	CantCom  *float64 `json:"cantcom"`
	UnidCom  *float64 `json:"unidcom"`
	ID       *int     `json:"id"`
}
