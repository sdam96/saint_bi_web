package models

// Agreement representa un convenio.
type Agreement struct {
	CodConv string  `json:"codconv"`
	Descrip string  `json:"descrip"`
	Autori  *string `json:"autori"`
	ID      int     `json:"id"`
}