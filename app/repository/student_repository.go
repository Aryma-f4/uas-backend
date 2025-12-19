package repository

import (
	"context"
	"database/sql"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/google/uuid"
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, u.email, s.program_study, 
		       s.academic_year, s.advisor_id, s.created_at,
		       COALESCE(lu.full_name, '') as advisor_name
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users lu ON l.user_id = lu.id
		WHERE s.id = $1
	`
	student := &entity.Student{}
	var advisorID sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.FullName, &student.Email,
		&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt, &student.AdvisorName,
	)
	if err != nil {
		return nil, err
	}
	if advisorID.Valid {
		uid, _ := uuid.Parse(advisorID.String)
		student.AdvisorID = &uid
	}
	return student, nil
}

func (r *StudentRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*entity.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, u.email, s.program_study, 
		       s.academic_year, s.advisor_id, s.created_at,
		       COALESCE(lu.full_name, '') as advisor_name
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users lu ON l.user_id = lu.id
		WHERE s.user_id = $1
	`
	student := &entity.Student{}
	var advisorID sql.NullString
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&student.ID, &student.UserID, &student.StudentID, &student.FullName, &student.Email,
		&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt, &student.AdvisorName,
	)
	if err != nil {
		return nil, err
	}
	if advisorID.Valid {
		uid, _ := uuid.Parse(advisorID.String)
		student.AdvisorID = &uid
	}
	return student, nil
}

func (r *StudentRepository) Create(ctx context.Context, student *entity.Student) error {
	query := `
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, advisor_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		student.ID, student.UserID, student.StudentID, student.ProgramStudy, student.AcademicYear, student.AdvisorID,
	)
	return err
}

func (r *StudentRepository) UpdateAdvisor(ctx context.Context, studentID, advisorID uuid.UUID) error {
	query := `UPDATE students SET advisor_id = $2 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, studentID, advisorID)
	return err
}

func (r *StudentRepository) List(ctx context.Context, limit, offset int) ([]*entity.Student, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM students`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, u.email, s.program_study, 
		       s.academic_year, s.advisor_id, s.created_at,
		       COALESCE(lu.full_name, '') as advisor_name
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users lu ON l.user_id = lu.id
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []*entity.Student
	for rows.Next() {
		student := &entity.Student{}
		var advisorID sql.NullString
		if err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID, &student.FullName, &student.Email,
			&student.ProgramStudy, &student.AcademicYear, &advisorID, &student.CreatedAt, &student.AdvisorName,
		); err != nil {
			return nil, 0, err
		}
		if advisorID.Valid {
			uid, _ := uuid.Parse(advisorID.String)
			student.AdvisorID = &uid
		}
		students = append(students, student)
	}

	return students, total, nil
}

func (r *StudentRepository) GetByAdvisorID(ctx context.Context, advisorID uuid.UUID, limit, offset int) ([]*entity.Student, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM students WHERE advisor_id = $1`
	if err := r.db.QueryRowContext(ctx, countQuery, advisorID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT s.id, s.user_id, s.student_id, u.full_name, u.email, s.program_study, 
		       s.academic_year, s.advisor_id, s.created_at,
		       COALESCE(lu.full_name, '') as advisor_name
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users lu ON l.user_id = lu.id
		WHERE s.advisor_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, advisorID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []*entity.Student
	for rows.Next() {
		student := &entity.Student{}
		var advID sql.NullString
		if err := rows.Scan(
			&student.ID, &student.UserID, &student.StudentID, &student.FullName, &student.Email,
			&student.ProgramStudy, &student.AcademicYear, &advID, &student.CreatedAt, &student.AdvisorName,
		); err != nil {
			return nil, 0, err
		}
		if advID.Valid {
			uid, _ := uuid.Parse(advID.String)
			student.AdvisorID = &uid
		}
		students = append(students, student)
	}

	return students, total, nil
}
