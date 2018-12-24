package models

// PrevPayment model
type PrevPayment struct {
	ID           int        `json:"id"`
	PhysicalOpID int        `json:"physical_op_id"`
	Year         int        `json:"year"`
	Value        int64      `json:"value"`
	Descript     NullString `json:"descript"`
}

// PrevPayments embeddes an array of PrevPayment for json export.
type PrevPayments struct {
	PrevPayments []PrevPayment `json:"PrevPayment"`
}
