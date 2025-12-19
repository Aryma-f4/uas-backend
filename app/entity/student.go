package entity

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	StudentID    string    `json:"student_id"`
	FullName     string    `json:"full_name,omitempty"`
	Email        string    `json:"email,omitempty"`
	ProgramStudy string    `json:"program_study"`
	AcademicYear string    `json:"academic_year"`
	AdvisorID    *uuid.UUID `json:"advisor_id,omitempty"`
	AdvisorName  string    `json:"advisor_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Lecturer struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	LecturerID string    `json:"lecturer_id"`
	FullName   string    `json:"full_name,omitempty"`
	Email      string    `json:"email,omitempty"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"created_at"`
}

type UpdateAdvisorRequest struct {
	AdvisorID uuid.UUID `json:"advisor_id" validate:"required"`
}
