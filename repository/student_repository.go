package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/airlangga/achievement-reporting/models"
	"github.com/google/uuid"
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) Create(ctx context.Context, student *models.Student) error {
	query := `
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at
	`

	return r.db.QueryRowContext(ctx, query,
		student.ID, student.UserID, student.StudentID, student.ProgramStudy, student.AcademicYear, student.AdvisorID,
	).Scan(&student.CreatedAt)
}

func (r *StudentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE user_id = $1
	`

	student := &models.Student{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy, &student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	return student, nil
}

func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE id = $1
	`

	student := &models.Student{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy, &student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("student not found")
		}
		return nil, err
	}

	return student, nil
}

func (r *StudentRepository) SetAdvisor(ctx context.Context, studentID, advisorID uuid.UUID) error {
	query := `
		UPDATE students
		SET advisor_id = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, advisorID, studentID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("student not found")
	}

	return nil
}

func (r *StudentRepository) GetAdvisees(ctx context.Context, advisorID uuid.UUID) ([]*models.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		WHERE advisor_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		student := &models.Student{}
		err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy, &student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

func (r *StudentRepository) List(ctx context.Context, limit, offset int) ([]*models.Student, error) {
	query := `
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []*models.Student
	for rows.Next() {
		student := &models.Student{}
		err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID, &student.ProgramStudy, &student.AcademicYear, &student.AdvisorID, &student.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}
