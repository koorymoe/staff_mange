package model

import "time"

type Project struct {
	ID           string    `db:"id" json:"id"`
	Code         string    `db:"code" json:"code"`
	Name         string    `db:"name" json:"name"`
	Rep          *string   `db:"rep" json:"rep"`
	Phone        *string   `db:"phone" json:"phone"`
	Location     *string   `db:"location" json:"location"`
	WorkType     *string   `db:"workType" json:"workType"`
	RefPerson    *string   `db:"refPerson" json:"refPerson"`
	Stage        string    `db:"stage" json:"stage"`
	Price        *string   `db:"price" json:"price"`
	Staff        *string   `db:"staff" json:"staff"`
	Time         *string   `db:"time" json:"time"`
	Task         *string   `db:"task" json:"task"`
	Priority     string    `db:"priority" json:"priority"`
	DeliveryDate *string   `db:"deliveryDate" json:"deliveryDate"`
	Survey       *string   `db:"survey" json:"survey"`
	SentToGroup  bool      `db:"sentToGroup" json:"sentToGroup"`
	CreatedAt    time.Time `db:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time `db:"updatedAt" json:"updatedAt"`
}

type ProjectStats struct {
	Contact  int `json:"اتصال"`
	Survey   int `json:"كشف"`
	Price    int `json:"سعر"`
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
	Name         string  `json:"name"`
	Rep          *string `json:"rep"`
	Phone        *string `json:"phone"`
	Location     *string `json:"location"`
	WorkType     *string `json:"workType"`
	RefPerson    *string `json:"refPerson"`
	Priority     *string `json:"priority"`
	DeliveryDate *string `json:"deliveryDate"`
}

type UpdateProjectRequest struct {
	Name         *string `json:"name"`
	Rep          *string `json:"rep"`
	Phone        *string `json:"phone"`
	Location     *string `json:"location"`
	WorkType     *string `json:"workType"`
	RefPerson    *string `json:"refPerson"`
	Stage        *string `json:"stage"`
	Price        *string `json:"price"`
	Staff        *string `json:"staff"`
	Time         *string `json:"time"`
	Task         *string `json:"task"`
	Priority     *string `json:"priority"`
	DeliveryDate *string `json:"deliveryDate"`
	Survey       *string `json:"survey"`
	SentToGroup  *bool   `json:"sentToGroup"`
}
