package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Aryma-f4/uas-backend/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository struct {
	db          *sql.DB
	mongoClient *mongo.Client
	collection  *mongo.Collection
}

func NewAchievementRepository(db *sql.DB, mongoClient *mongo.Client) *AchievementRepository {
	collection := mongoClient.Database("achievement_db").Collection("achievements")
	return &AchievementRepository{
		db:         db,
		mongoClient: mongoClient,
		collection:  collection,
	}
}

func (r *AchievementRepository) CreateReference(ctx context.Context, ref *models.AchievementReference) error {
	query := `
		INSERT INTO achievement_references (id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		ref.ID, ref.StudentID, ref.MongoAchievementID, ref.Status,
	).Scan(&ref.CreatedAt, &ref.UpdatedAt)
}

func (r *AchievementRepository) CreateAchievement(ctx context.Context, achievement *models.Achievement) error {
	result, err := r.collection.InsertOne(ctx, achievement)
	if err != nil {
		return err
	}

	achievement.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *AchievementRepository) GetAchievementByID(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
	var achievement models.Achievement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&achievement)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("achievement not found")
		}
		return nil, err
	}
	return &achievement, nil
}

func (r *AchievementRepository) GetReferenceByID(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`

	ref := &models.AchievementReference{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
		&ref.CreatedAt, &ref.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("achievement reference not found")
		}
		return nil, err
	}

	return ref, nil
}

func (r *AchievementRepository) UpdateReferenceStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
		UPDATE achievement_references
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("achievement reference not found")
	}

	return nil
}

func (r *AchievementRepository) SubmitForVerification(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE achievement_references
		SET status = 'submitted', submitted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status = 'draft'
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("achievement must be in draft status")
	}

	return nil
}

func (r *AchievementRepository) VerifyAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error {
	query := `
		UPDATE achievement_references
		SET status = 'verified', verified_at = NOW(), verified_by = $1, updated_at = NOW()
		WHERE id = $2 AND status = 'submitted'
	`

	result, err := r.db.ExecContext(ctx, query, verifiedBy, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("achievement must be in submitted status")
	}

	return nil
}

func (r *AchievementRepository) RejectAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID, note string) error {
	query := `
		UPDATE achievement_references
		SET status = 'rejected', verified_at = NOW(), verified_by = $1, rejection_note = $2, updated_at = NOW()
		WHERE id = $3 AND status = 'submitted'
	`

	result, err := r.db.ExecContext(ctx, query, verifiedBy, note, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("achievement must be in submitted status")
	}

	return nil
}

func (r *AchievementRepository) GetStudentAchievements(ctx context.Context, studentID uuid.UUID, limit, offset int) ([]*models.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE student_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, studentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*models.AchievementReference
	for rows.Next() {
		ref := &models.AchievementReference{}
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		references = append(references, ref)
	}

	return references, nil
}

func (r *AchievementRepository) GetAdviseeAchievements(ctx context.Context, adviseeIDs []uuid.UUID, limit, offset int) ([]*models.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE student_id = ANY($1) AND status = 'submitted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, adviseeIDs, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*models.AchievementReference
	for rows.Next() {
		ref := &models.AchievementReference{}
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		references = append(references, ref)
	}

	return references, nil
}

func (r *AchievementRepository) DeleteAchievement(ctx context.Context, id primitive.ObjectID, referenceID uuid.UUID) error {
	// Soft delete in MongoDB
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"deletedAt": time.Now()}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Update PostgreSQL reference
	query := "UPDATE achievement_references SET status = 'deleted' WHERE id = $1"
	_, err = r.db.ExecContext(ctx, query, referenceID)
	return err
}

func (r *AchievementRepository) ListAll(ctx context.Context, limit, offset int) ([]*models.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var references []*models.AchievementReference
	for rows.Next() {
		ref := &models.AchievementReference{}
		err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&ref.SubmittedAt, &ref.VerifiedAt, &ref.VerifiedBy, &ref.RejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		references = append(references, ref)
	}

	return references, nil
}
