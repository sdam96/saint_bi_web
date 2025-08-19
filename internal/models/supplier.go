package models

// Supplier representa los datos de un proveedor.
type Supplier struct {
	CodProv   *string  `json:"codprov"`
	Descrip   *string  `json:"descrip"`
	TipoPrv   *int     `json:"tipoprv"`
	TipoID3   *int     `json:"tipoid3"`
	TipoID    *int     `json:"tipoid"`
	ID3       *string  `json:"id3"`
	DescOrder *string  `json:"descorder"`
	Clase     *string  `json:"clase"`
	Activo    *int     `json:"activo"`
	Represent *string  `json:"represent"`
	Direc1    *string  `json:"direc1"`
	Direc2    *string  `json:"direc2"`
	Pais      *int     `json:"pais"`
	Estado    *int     `json:"estado"`
	Ciudad    *int     `json:"ciudad"`
	Municipio *int     `json:"municipio"`
	ZipCode   *string  `json:"zipcode"`
	Telef     *string  `json:"telef"`
	Movil     *string  `json:"movil"`
	Fax       *string  `json:"fax"`
	Email     *string  `json:"email"`
	FechaE    *string  `json:"fechae"`
	EsReten   *int     `json:"esreten"`
	RetenISLR *float64 `json:"retenislr"`
	DiasCred  *int     `json:"diascred"`
	Observa   *string  `json:"observa"`
	EsMoneda  *int     `json:"esmoneda"`
	Saldo     *float64 `json:"saldo"`
	MontoMax  *float64 `json:"montomax"`
	PagosA    *int     `json:"pagosa"`
	PromPago  *int     `json:"prompago"`
	RetenIVA  *float64 `json:"reteniva"`
	FechaUC   *string  `json:"fechauc"`
	MontoUC   *float64 `json:"montouc"`
	NumeroUC  *string  `json:"numerouc"`
	FechaUP   *string  `json:"fechaup"`
	MontoUP   *float64 `json:"montoup"`
	NumeroUP  *string  `json:"numeroup"`
	PorctRet  *float64 `json:"porctret"`
	ID        *int     `json:"id"`
}
