package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	Nim       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   int       `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Prestation struct {
	ID             int    `json:"id"`
	StudentID      int    `json:"student_id"`
	NamePrestation string `json:"name_prestation"`
	Juara          string `json:"juara"`
}

type StudentWithPrestation struct {
	Student
	Prestasi []Prestation `json:"prestasi"`
}

type CreatedStudentRequest struct {
	Nim   string  `json:"nim" validate:"required,nim"`
	Name  string  `json:"name" validate:"required,min=3,max=50"`
	Grade float64 `json:"grade" validate:"required,min=0,max=100"`
}

type ReplaceStudentRequest struct {
	Name     string  `json:"name" validate:"required,min=3,max=50"`
	Grade    float64 `json:"grade" validate:"required,min=0,max=100"`
	IsActive bool    `json:"is_active"`
}

type PatchStudentRequest struct {
	Name     *string  `json:"name,omitempty" validate:"omitnil,min=3,max=50"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitnil,min=0,max=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type StudentCursorQuery struct {
	Search   string
	IsActive *bool
	Limit    int
	After    *Cursor
}

//Amplop baku untuk semua respon
type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

type Meta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	Total     int `json:"total"`
	TotalPage int `json:"total_page"`
}

type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	MinGrade *float64
	MaxGrade *float64
}

func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
