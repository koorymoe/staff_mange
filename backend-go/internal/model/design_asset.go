package model

import "time"

// ═══ معرض التصاميم ═══
//
// «معرض تصاميم مستقل» — ما ينربط بحجز، لأن المصممة ما تشوف الحجوزات.
type DesignAsset struct {
	ID             string     `db:"id" json:"id"`
	Title          string     `db:"title" json:"title"`
	Category       string     `db:"category" json:"category"`
	Notes          *string    `db:"notes" json:"notes"`
	FileKey        string     `db:"fileKey" json:"fileKey"`
	FileType       *string    `db:"fileType" json:"fileType"`
	UploadedByID   *string    `db:"uploadedById" json:"uploadedById"`
	UploadedByName string     `db:"uploadedByName" json:"uploadedByName"`
	CreatedAt      time.Time  `db:"createdAt" json:"createdAt"`
	ArchivedAt     *time.Time `db:"archivedAt" json:"archivedAt"`
}

// DesignCategoryLabels التصنيفات — **بيانات لا كود**: التوسعة بسطر
// هنا، مو بشرط جديد بكل شاشة. نفس مبدأ كل تصنيفات النظام.
var DesignCategoryLabels = map[string]string{
	"LOGO":         "شعار",
	"BANNER":       "بنر",
	"SOCIAL":       "سوشيال ميديا",
	"PRINT":        "طباعة",
	"PRESENTATION": "عرض تقديمي",
	"OTHER":        "غير ذلك",
}

type CreateDesignAssetRequest struct {
	Title    string  `json:"title"`
	Category string  `json:"category"`
	Notes    *string `json:"notes"`
	FileKey  string  `json:"fileKey"`
	FileType *string `json:"fileType"`
}
