package model

import "time"

type Student struct {
	ID int `json:"id"`
	Nim string `json:"nim"`
	Name string `json:"name"`
	Grade float64 `json:"grade"`
	IsActive bool `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type CreatedStudentRequest struct{
	Nim	string	`json:"nim"` 
	Name string	`json:"name"` 
	Grade float64	`json:"grade"` 
}

type ReplaceStudentRequest struct{
	Name string `json:"name"`
	Grade float64 `json:"grade"`
	IsActive bool `json:"is_active"`
}

type PatchStudentRequest struct{
	Name *string `json:"name,omitempty"`
	Grade *float64 `json:"grade,omitempty"`
	IsActive *bool `json:"is_active,omitempty"`
}

//Amplop baku untuk semua respon
type WebResponse struct{
	Success bool `json:"success"`
	Message string `json:"message"`
	Data any `json:"data,omitempty"`
	Meta *Meta `json:"meta,omitempty"`
	Errors any `json:"errors,omitempty"`
}

type Meta struct{
	Page int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	TotalPage int `json:"total_page"`
}

type ListQuery struct{
	Page		int
	Limit		int
	Search		string
	Sort		string
	Order		string
	IsActive	*bool
	MinGrade 	*float64
	MaxGrade 	*float64
}

func (q ListQuery) Offset() int { 
	return (q.Page - 1) * q.Limit
}