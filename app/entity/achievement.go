package entity

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AchievementStatus string

const (
	StatusDraft     AchievementStatus = "draft"
	StatusSubmitted AchievementStatus = "submitted"
	StatusVerified  AchievementStatus = "verified"
	StatusRejected  AchievementStatus = "rejected"
)

type AchievementType string

const (
	TypeAcademic      AchievementType = "academic"
	TypeCompetition   AchievementType = "competition"
	TypeOrganization  AchievementType = "organization"
	TypePublication   AchievementType = "publication"
	TypeCertification AchievementType = "certification"
	TypeOther         AchievementType = "other"
)

// MongoDB Achievement Document
type Achievement struct {
	ID              primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	StudentID       uuid.UUID              `json:"student_id" bson:"studentId"`
	AchievementType AchievementType        `json:"achievement_type" bson:"achievementType"`
	Title           string                 `json:"title" bson:"title"`
	Description     string                 `json:"description" bson:"description"`
	Details         map[string]interface{} `json:"details" bson:"details"`
	Attachments     []Attachment           `json:"attachments" bson:"attachments"`
	Tags            []string               `json:"tags" bson:"tags"`
	Points          int                    `json:"points" bson:"points"`
	CreatedAt       time.Time              `json:"created_at" bson:"createdAt"`
	UpdatedAt       time.Time              `json:"updated_at" bson:"updatedAt"`
}

type Attachment struct {
	FileName   string    `json:"file_name" bson:"fileName"`
	FileURL    string    `json:"file_url" bson:"fileUrl"`
	FileType   string    `json:"file_type" bson:"fileType"`
	UploadedAt time.Time `json:"uploaded_at" bson:"uploadedAt"`
}

// PostgreSQL Achievement Reference
type AchievementReference struct {
	ID                 uuid.UUID         `json:"id"`
	StudentID          uuid.UUID         `json:"student_id"`
	MongoAchievementID string            `json:"mongo_achievement_id"`
	Status             AchievementStatus `json:"status"`
	SubmittedAt        *time.Time        `json:"submitted_at,omitempty"`
	VerifiedAt         *time.Time        `json:"verified_at,omitempty"`
	VerifiedBy         *uuid.UUID        `json:"verified_by,omitempty"`
	RejectionNote      string            `json:"rejection_note,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

type AchievementStatusHistory struct {
	ID               uuid.UUID         `json:"id"`
	AchievementRefID uuid.UUID         `json:"achievement_ref_id"`
	OldStatus        AchievementStatus `json:"old_status"`
	NewStatus        AchievementStatus `json:"new_status"`
	ChangedBy        uuid.UUID         `json:"changed_by"`
	Note             string            `json:"note"`
	CreatedAt        time.Time         `json:"created_at"`
}

// Combined Achievement Response
type AchievementResponse struct {
	ID              string                 `json:"id"`
	StudentID       uuid.UUID              `json:"student_id"`
	StudentName     string                 `json:"student_name,omitempty"`
	AchievementType AchievementType        `json:"achievement_type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Attachments     []Attachment           `json:"attachments"`
	Tags            []string               `json:"tags"`
	Points          int                    `json:"points"`
	Status          AchievementStatus      `json:"status"`
	SubmittedAt     *time.Time             `json:"submitted_at,omitempty"`
	VerifiedAt      *time.Time             `json:"verified_at,omitempty"`
	VerifiedBy      *uuid.UUID             `json:"verified_by,omitempty"`
	RejectionNote   string                 `json:"rejection_note,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Request DTOs
type CreateAchievementRequest struct {
	AchievementType AchievementType        `json:"achievement_type" validate:"required"`
	Title           string                 `json:"title" validate:"required"`
	Description     string                 `json:"description"`
	Details         map[string]interface{} `json:"details"`
	Tags            []string               `json:"tags"`
}

type UpdateAchievementRequest struct {
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

type RejectAchievementRequest struct {
	RejectionNote string `json:"rejection_note" validate:"required"`
}

type AchievementFilter struct {
	Status          string `query:"status"`
	AchievementType string `query:"type"`
	StudentID       string `query:"student_id"`
	StartDate       string `query:"start_date"`
	EndDate         string `query:"end_date"`
	Page            int    `query:"page"`
	Limit           int    `query:"limit"`
}
