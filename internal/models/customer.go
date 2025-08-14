package models

// Customer representa los datos de un cliente.
type Customer struct {
	CodClie    *string  `json:"codclie"`
	Descrip    *string  `json:"descrip"`
	ID3        *string  `json:"id3"`
	TipoID3    *int     `json:"tipoid3"`
	TipoID     *int     `json:"tipoid"`
	Activo     *int     `json:"activo"`
	DescOrder  *string  `json:"descorder"`
	Clase      *string  `json:"clase"`
	Represent  *string  `json:"represent"`
	Direc1     *string  `json:"direc1"`
	Direc2     *string  `json:"direc2"`
	Pais       *int     `json:"pais"`
	Estado     *int     `json:"estado"`
	Ciudad     *int     `json:"ciudad"`
	Municipio  *int     `json:"municipio"`
	ZipCode    *string  `json:"zipcode"`
	Telef      *string  `json:"telef"`
	Movil      *string  `json:"movil"`
	Email      *string  `json:"email"`
	Fax        *string  `json:"fax"`
	FechaE     *string  `json:"fechae"`
	CodZona    *string  `json:"codzona"`
	CodVend    *string  `json:"codvend"`
	CodConv    *string  `json:"codconv"`
	CodAlte    *string  `json:"codalte"`
	TipoCli    *int     `json:"tipocli"`
	TipoReg    *int     `json:"tiporeg"`
	TipoPVP    *int     `json:"tipopvp"`
	Observa    *string  `json:"observa"`
	EsMoneda   *int     `json:"esmoneda"`
	EsCredito  *int     `json:"escredito"`
	LimiteCred *float64 `json:"limitecred"`
	DiasCred   *int     `json:"diascred"`
	EsToleran  *int     `json:"estoleran"`
	DiasTole   *int     `json:"diastole"`
	IntMora    *int     `json:"*intmora"`
	Descto     *float64 `json:"descto"`
	Saldo      *float64 `json:"saldo"`
	PagosA     *float64 `json:"pagosa"`
	FechaUV    *string  `json:"fechauv"`
	MontoUV    *float64 `json:"montouv"`
	NumeroUV   *string  `json:"numerouv"`
	FechaUP    *string  `json:"fechaup"`
	MontoUP    *float64 `json:"montoup"`
	NumeroUP   *string  `json:"numeroup"`
	MontoMax   *float64 `json:"montomax"`
	MtoMaxCred *float64 `json:"mtomaxcred"`
	PromPago   *int     `json:"prompago"`
	RetenIVA   *float64 `json:"reteniva"`
	SaldoPtos  *int     `json:"saldoptos"`
	EsReten    *int     `json:"esreten"`
	DescripExt *string  `json:"descripext"`
	ID         *int     `json:"id"`
}
