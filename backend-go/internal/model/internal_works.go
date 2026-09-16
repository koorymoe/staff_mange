package model

import "time"

// إحصائية الأعمال داخل الشركة خلال شهر: شنو انخلص جوه، شكد انشتغل جوه،
// ومنو اشتغل — مقابل الشغل الي طلع للزبون.

type InternalWorkServiceRow struct {
	Name   string  `db:"name" json:"name"`
	Count  int     `db:"count" json:"count"`
	Amount float64 `db:"amount" json:"amount"`
}

type InternalWorkCrewRow struct {
	EmployeeName string `db:"employeeName" json:"employeeName"`
	Count        int    `db:"count" json:"count"`
}

type InternalWorkRow struct {
	Code        string     `db:"code" json:"code"`
	CompletedAt *time.Time `db:"completedAt" json:"completedAt"`
	ServiceName string     `db:"serviceName" json:"serviceName"`
	Amount      float64    `db:"amount" json:"amount"`
}

type InternalWorksReport struct {
	Month string `db:"-" json:"month"`

	InHouseCount  int     `db:"inHouseCount" json:"inHouseCount"`
	OnSiteCount   int     `db:"onSiteCount" json:"onSiteCount"`
	InHouseAmount float64 `db:"inHouseAmount" json:"inHouseAmount"`

	Services []InternalWorkServiceRow `db:"-" json:"services"`
	Crew     []InternalWorkCrewRow    `db:"-" json:"crew"`
	Works    []InternalWorkRow        `db:"-" json:"works"`
}
