package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// RawJSON يمرر بيانات jsonb كما هي (بدون تحويلها لسترنغ Go عادي)، يستخدم لعمود
// survey اللي يخزن مصفوفة إجابات استمارة الكشف الفني الـ17 سؤال بصيغة JSON array.
type RawJSON []byte

func (r *RawJSON) Scan(src any) error {
	if src == nil {
		*r = nil
		return nil
	}
	switch v := src.(type) {
	case []byte:
		*r = append(RawJSON{}, v...)
		return nil
	case string:
		*r = RawJSON(v)
		return nil
	default:
		return fmt.Errorf("unsupported Scan type %T for RawJSON", src)
	}
}

func (r RawJSON) Value() (driver.Value, error) {
	if len(r) == 0 {
		return nil, nil
	}
	return []byte(r), nil
}

func (r RawJSON) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return r, nil
}

func (r *RawJSON) UnmarshalJSON(data []byte) error {
	*r = append((*r)[0:0], data...)
	return nil
}

type Project struct {
	ID           string   `db:"id" json:"id"`
	Code         string   `db:"code" json:"code"`
	Name         string   `db:"name" json:"name"`
	Rep          *string  `db:"rep" json:"rep"`
	Phone        *string  `db:"phone" json:"phone"`
	Location     *string  `db:"location" json:"location"`
	MapLatitude  *float64 `db:"mapLatitude" json:"mapLatitude"`
	MapLongitude *float64 `db:"mapLongitude" json:"mapLongitude"`
	WorkType     *string  `db:"workType" json:"workType"`
	RefPerson    *string  `db:"refPerson" json:"refPerson"`
	Stage        string   `db:"stage" json:"stage"`
	Price        *string  `db:"price" json:"price"`
	Staff        *string  `db:"staff" json:"staff"`
	Time         *string  `db:"time" json:"time"`
	Task         *string  `db:"task" json:"task"`
	Priority     string   `db:"priority" json:"priority"`
	DeliveryDate *string  `db:"deliveryDate" json:"deliveryDate"`
	Survey       *RawJSON `db:"survey" json:"survey"`
	BookingID    *string  `db:"bookingId" json:"bookingId"`
	// العقد: يترفع كـPDF (base64) قبل التوقيع وبعده — كلاهما اختياري ومرتبط
	// بنفس المشروع.
	ContractPdfBase64       *string `db:"contractPdfBase64" json:"contractPdfBase64"`
	SignedContractPdfBase64 *string `db:"signedContractPdfBase64" json:"signedContractPdfBase64"`
	// علمان محسوبان بالاستعلام — القائمة ترجّعهما بدل ملفات الـPDF الثقيلة نفسها.
	HasContract           bool      `db:"hasContract" json:"hasContract"`
	HasSignedContract     bool      `db:"hasSignedContract" json:"hasSignedContract"`
	ResponsibleEmployeeID *string   `db:"responsibleEmployeeId" json:"responsibleEmployeeId"`
	SurveyorEmployeeID    *string   `db:"surveyorEmployeeId" json:"surveyorEmployeeId"`
	CreatedAt             time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt             time.Time `db:"updatedAt" json:"updatedAt"`
}

type ProjectStats struct {
	Contact  int `json:"اتصال"`
	Survey   int `json:"كشف"`
	Price    int `json:"سعر"`
	Contract int `json:"عقد"`
	Execute  int `json:"تنفيذ"`
	Done     int `json:"مكتمل"`
	Rejected int `json:"مرفوض"`
}

type ProjectListResponse struct {
	Projects []Project    `json:"projects"`
	Stats    ProjectStats `json:"stats"`
	Stages   []string     `json:"stages"`
}

type CreateProjectRequest struct {
	Name                  string   `json:"name"`
	Rep                   *string  `json:"rep"`
	Phone                 *string  `json:"phone"`
	Location              *string  `json:"location"`
	MapLatitude           *float64 `json:"mapLatitude"`
	MapLongitude          *float64 `json:"mapLongitude"`
	WorkType              *string  `json:"workType"`
	RefPerson             *string  `json:"refPerson"`
	Priority              *string  `json:"priority"`
	DeliveryDate          *string  `json:"deliveryDate"`
	BookingID             *string  `json:"bookingId"`
	ResponsibleEmployeeID *string  `json:"responsibleEmployeeId"`
	SurveyorEmployeeID    *string  `json:"surveyorEmployeeId"`
}

type UpdateProjectRequest struct {
	Name                    *string  `json:"name"`
	Rep                     *string  `json:"rep"`
	Phone                   *string  `json:"phone"`
	Location                *string  `json:"location"`
	MapLatitude             *float64 `json:"mapLatitude"`
	MapLongitude            *float64 `json:"mapLongitude"`
	WorkType                *string  `json:"workType"`
	RefPerson               *string  `json:"refPerson"`
	Stage                   *string  `json:"stage"`
	Price                   *string  `json:"price"`
	Staff                   *string  `json:"staff"`
	Time                    *string  `json:"time"`
	Task                    *string  `json:"task"`
	Priority                *string  `json:"priority"`
	DeliveryDate            *string  `json:"deliveryDate"`
	Survey                  *RawJSON `json:"survey"`
	ContractPdfBase64       *string  `json:"contractPdfBase64"`
	SignedContractPdfBase64 *string  `json:"signedContractPdfBase64"`
	ResponsibleEmployeeID   *string  `json:"responsibleEmployeeId"`
	SurveyorEmployeeID      *string  `json:"surveyorEmployeeId"`
	// أعلام حذف العقد — COALESCE ما يقدر يصفّر عمود، فنحتاج علم صريح.
	// الحذف محصور بالأدمن (الراوت DELETE محمي بـrequireAdmin).
	ClearContract       bool `json:"-"`
	ClearSignedContract bool `json:"-"`
}
