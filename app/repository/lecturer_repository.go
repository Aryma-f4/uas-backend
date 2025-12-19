package repository

import (
	"context"
	"database/sql"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/google/uuid"
)

type LecturerRepository struct {
	db *sql.DB
}

func NewLecturerRepository(db *sql.DB) *LecturerRepository {
	return &LecturerRepository{db: db}
}

func (r *LecturerRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Lecturer, error) {
	query := `
		SELECT l.id, l.user_id, l.lecturer_id, u.full_name, u.email, l.department, l.created_at
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE l.id = $1
	`
	lecturer := &entity.Lecturer{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.FullName, &lecturer.Email,
		&lecturer.Department, &lecturer.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return lecturer, nil
}

func (r *LecturerRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Lecturer, error) {
	query := `
		SELECT l.id, l.user_id, l.lecturer_id, u.full_name, u.email, l.department, l.created_at
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE l.user_id = $1
	`
	lecturer := &entity.Lecturer{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.FullName, &lecturer.Email,
		&lecturer.Department, &lecturer.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return lecturer, nil
}

func (r *LecturerRepository) Create(ctx context.Context, lecturer *entity.Lecturer) error {
	query := `
		INSERT INTO lecturers (id, user_id, lecturer_id, department)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, lecturer.ID, lecturer.UserID, lecturer.LecturerID, lecturer.Department)
	return err
}

func (r *LecturerRepository) List(ctx context.Context, limit, offset int) ([]*entity.Lecturer, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM lecturers`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT l.id, l.user_id, l.lecturer_id, u.full_name, u.email, l.department, l.created_at
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		ORDER BY l.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []*entity.Lecturer
	for rows.Next() {
		lecturer := &entity.Lecturer{}
		if err := rows.Scan(
			&lecturer.ID, &lecturer.UserID, &lecturer.LecturerID, &lecturer.FullName, &lecturer.Email,
			&lecturer.Department, &lecturer.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, lecturer)
	}

	return lecturers, total, nil
}
