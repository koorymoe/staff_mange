package model

import "time"

type Employee struct {
	ID                   string    `db:"id" json:"id"`
	Name                 string    `db:"name" json:"name"`
	Certificate          *string   `db:"certificate" json:"certificate"`
	Position             *string   `db:"position" json:"position"`
	Phone                *string   `db:"phone" json:"phone"`
	Status               string    `db:"status" json:"status"`
	Role                 string    `db:"role" json:"role"`
	OnDuty               bool      `db:"onDuty" json:"onDuty"`
	Username             *string   `db:"username" json:"username"`
	Password             *string   `db:"password" json:"-"`
	HasDrivingLicense    bool      `db:"hasDrivingLicense" json:"hasDrivingLicense"`
	HasSafetyCertificate bool      `db:"hasSafetyCertificate" json:"hasSafetyCertificate"`
	Salary               *float64  `db:"salary" json:"salary"`
	Shift                *string   `db:"shift" json:"shift"`
	MonthlyLeaves        int       `db:"monthlyLeaves" json:"monthlyLeaves"`
	JobTitle             *string   `db:"jobTitle" json:"jobTitle"`
	LeaderSkillLevel     int       `db:"leaderSkillLevel" json:"leaderSkillLevel"`
	IsLeader             bool      `db:"isLeader" json:"isLeader"`
	IsTrainee            bool      `db:"isTrainee" json:"isTrainee"`
	CreatedAt            time.Time `db:"createdAt" json:"createdAt"`
}

type CreateEmployeeRequest struct {
	Name        string  `json:"name"`
	Certificate *string `json:"certificate"`
	Position    *string `json:"position"`
	Phone       *string `json:"phone"`
	Username    *string `json:"username"`
	Password    *string `json:"password"`
}

type UpdateEmployeeRequest struct {
	Name                 *string `json:"name"`
	Certificate          *string `json:"certificate"`
	Position             *string `json:"position"`
	Phone                *string `json:"phone"`
	Status               *string `json:"status"`
	Role                 *string `json:"role"`
	OnDuty               *bool   `json:"onDuty"`
	Username             *string `json:"username"`
	Password             *string `json:"password"`
	HasDrivingLicense    *bool   `json:"hasDrivingLicense"`
	HasSafetyCertificate *bool   `json:"hasSafetyCertificate"`
	IsLeader             *bool   `json:"isLeader"`
	IsTrainee            *bool   `json:"isTrainee"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
