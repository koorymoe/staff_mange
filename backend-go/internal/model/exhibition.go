package model

import (
	"time"

	"github.com/lib/pq"
)

// Exhibition معرض تجاري تحضره الشركة (وحدة التقنيين → إدارة المعارض) — أي تقني
// أو إداري يقدر يضيفه، لكن الترشيح (من يروح فعلياً) حصراً بيد المدير. بعد
// انتهاء المعرض يُؤرشف وتُكتب/تُولَّد تقرير الزيارة.
type Exhibition struct {
	ID                   string         `db:"id" json:"id"`
	Title                string         `db:"title" json:"title"`
	Location             string         `db:"location" json:"location"`
	StartDate            string         `db:"startDate" json:"startDate"` // "YYYY-MM-DD"
	EndDate              string         `db:"endDate" json:"endDate"`
	Companies            pq.StringArray `db:"companies" json:"companies"`
	ProductsToShow       pq.StringArray `db:"productsToShow" json:"productsToShow"`
	NominatedEmployeeIDs pq.StringArray `db:"nominatedEmployeeIds" json:"nominatedEmployeeIds"`
	BusinessCardPhotos   pq.StringArray `db:"businessCardPhotos" json:"businessCardPhotos"`
	KeyFindings          *string        `db:"keyFindings" json:"keyFindings"` // اهم ما اكتُشف بالمعرض (يكتبه التقني)
	VisitReport          *string        `db:"visitReport" json:"visitReport"` // تقرير الزيارة المولَّد بالذكاء الصناعي
	Archived             bool           `db:"archived" json:"archived"`
	CreatedByID          string         `db:"createdById" json:"-"`
	CreatedAt            time.Time      `db:"createdAt" json:"createdAt"`

	CreatedBy         *EmployeeBrief  `db:"-" json:"createdBy"`
	NominatedEmployee []EmployeeBrief `db:"-" json:"nominatedEmployees"`
}

type CreateExhibitionRequest struct {
	Title          string   `json:"title"`
	Location       string   `json:"location"`
	StartDate      string   `json:"startDate"`
	EndDate        string   `json:"endDate"`
	Companies      []string `json:"companies"`
	ProductsToShow []string `json:"productsToShow"`
}

type NominateExhibitionRequest struct {
	EmployeeIDs []string `json:"employeeIds"`
}

type AddExhibitionPhotosRequest struct {
	PhotoUrls []string `json:"photoUrls"`
}

type SetExhibitionFindingsRequest struct {
	KeyFindings string `json:"keyFindings"`
}
