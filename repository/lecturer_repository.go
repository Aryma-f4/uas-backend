package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Aryma-f4/uas-backend/models"
	"github.com/google/uuid"
)

type LecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
	return &LecturerRepository{db: db}
}

func (r *LecturerRepository) Create(ctx context.Context, lecturer *models.Lecturer) error {
	query := `
		INSERT INTO lecturers (id, user_id, lecturer_id, department)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at
	`

	return r.db.QueryRowContext(ctx, query,
		lecturer.ID, lecturer.UserID, lecturer.LecturerID, lecturer.Department,
	).Scan(&lecturer.CreatedAt)
}

func (r *LecturerRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Lecturer, error) {
	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE user_id = $1
	`

	lecturer := &models.Lecturer{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("lecturer not found")
		}
		return nil, err
	}

	return lecturer, nil
}

func (r *LecturerRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Lecturer, error) {
	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		WHERE id = $1
	`

	lecturer := &models.Lecturer{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("lecturer not found")
		}
		return nil, err
	}

	return lecturer, nil
}

func (r *LecturerRepository) List(ctx context.Context, limit, offset int) ([]*models.Lecturer, error) {
	query := `
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lecturers []*models.Lecturer
	for rows.Next() {
		lecturer := &models.Lecturer{}
		err := rows.Scan(
			&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.Department, &lecturer.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		lecturers = append(lecturers, lecturer)
	}

	return lecturers, nil
}
