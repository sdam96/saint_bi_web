package models

// InventoryOperation representa una operación de inventario.
type InventoryOperation struct {
	ID         *int     `json:"id"`
	CodSucu    *string  `json:"codsucu"`
	TipoOpi    *string  `json:"tipoopi"`
	NumeroD    *string  `json:"numerod"`
	CodEsta    *string  `json:"codesta"`
	CodUsua    *string  `json:"codusua"`
	CodUbic    *string  `json:"codubic"`
	CodUbic2   *string  `json:"codubic2"`
	Signo      *int     `json:"signo"`
	FechaT     *string  `json:"fechat"`
	OTipo      *string  `json:"otipo"`
	ONumero    *string  `json:"onumero"`
	Autori     *string  `json:"autori"`
	Respon     *string  `json:"respon"`
	UsoMat     *string  `json:"usomat"`
	OrdenC     *string  `json:"ordenc"`
	Monto      *float64 `json:"monto"`
	FechaE     *string  `json:"fechae"`
	FechaV     *string  `json:"fechav"`
	CodOper    *string  `json:"codoper"`
	UsoInterno *string  `json:"usointerno"`
	CodClie    *string  `json:"codclie"`
	NroTurno   *int     `json:"nroturno"`
	Notas1     *string  `json:"notas1"`
	Notas2     *string  `json:"notas2"`
	Notas3     *string  `json:"notas3"`
	Notas4     *string  `json:"notas4"`
	Notas5     *string  `json:"notas5"`
	Notas6     *string  `json:"notas6"`
	Notas7     *string  `json:"notas7"`
	Notas8     *string  `json:"notas8"`
	Notas9     *string  `json:"notas9"`
	Notas10    *string  `json:"notas10"`
	CreatedAt  *string  `json:"createdat"`
	UpdatedAt  *string  `json:"updatedat"`
	NroUnico   *int     `json:"nrounico"`
}
