package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AchievementRepository struct {
	db         *sql.DB
	collection *mongo.Collection
}

func NewAchievementRepository(db *sql.DB, mongoDB *mongo.Database) *AchievementRepository {
	return &AchievementRepository{
		db:         db,
		collection: mongoDB.Collection("achievements"),
	}
}

// MongoDB Operations
func (r *AchievementRepository) CreateMongo(ctx context.Context, achievement *entity.Achievement) (primitive.ObjectID, error) {
	achievement.CreatedAt = time.Now()
	achievement.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, achievement)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *AchievementRepository) GetMongoByID(ctx context.Context, id primitive.ObjectID) (*entity.Achievement, error) {
	var achievement entity.Achievement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&achievement)
	if err != nil {
		return nil, err
	}
	return &achievement, nil
}

func (r *AchievementRepository) UpdateMongo(ctx context.Context, id primitive.ObjectID, achievement *entity.Achievement) error {
	achievement.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": achievement},
	)
	return err
}

func (r *AchievementRepository) DeleteMongo(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *AchievementRepository) ListMongo(ctx context.Context, filter bson.M, limit, offset int64) ([]*entity.Achievement, error) {
	opts := options.Find().SetLimit(limit).SetSkip(offset).SetSort(bson.M{"createdAt": -1})
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var achievements []*entity.Achievement
	if err := cursor.All(ctx, &achievements); err != nil {
		return nil, err
	}
	return achievements, nil
}

func (r *AchievementRepository) CountMongo(ctx context.Context, filter bson.M) (int64, error) {
	return r.collection.CountDocuments(ctx, filter)
}

func (r *AchievementRepository) GetStatsMongo(ctx context.Context, studentIDs []uuid.UUID) (map[string]int, error) {
	// Convert UUIDs to interface slice for $in query
	ids := make([]interface{}, len(studentIDs))
	for i, id := range studentIDs {
		ids[i] = id
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"studentId": bson.M{"$in": ids}}}},
		{{Key: "$group", Value: bson.M{
			"_id":   "$achievementType",
			"count": bson.M{"$sum": 1},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats[result.ID] = result.Count
	}
	return stats, nil
}

// PostgreSQL Operations (Achievement References)
func (r *AchievementRepository) CreateReference(ctx context.Context, ref *entity.AchievementReference) error {
	query := `
		INSERT INTO achievement_references (id, student_id, mongo_achievement_id, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, ref.ID, ref.StudentID, ref.MongoAchievementID, ref.Status)
	return err
}

func (r *AchievementRepository) GetReferenceByMongoID(ctx context.Context, mongoID string) (*entity.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1
	`
	ref := &entity.AchievementReference{}
	var submittedAt, verifiedAt sql.NullTime
	var verifiedBy sql.NullString
	var rejectionNote sql.NullString

	err := r.db.QueryRowContext(ctx, query, mongoID).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
		&submittedAt, &verifiedAt, &verifiedBy, &rejectionNote,
		&ref.CreatedAt, &ref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if submittedAt.Valid {
		ref.SubmittedAt = &submittedAt.Time
	}
	if verifiedAt.Valid {
		ref.VerifiedAt = &verifiedAt.Time
	}
	if verifiedBy.Valid {
		id, _ := uuid.Parse(verifiedBy.String)
		ref.VerifiedBy = &id
	}
	if rejectionNote.Valid {
		ref.RejectionNote = rejectionNote.String
	}

	return ref, nil
}

func (r *AchievementRepository) UpdateReferenceStatus(ctx context.Context, mongoID string, status entity.AchievementStatus, verifiedBy *uuid.UUID, rejectionNote string) error {
	var query string
	var args []interface{}

	switch status {
	case entity.StatusSubmitted:
		query = `UPDATE achievement_references SET status = $1, submitted_at = NOW(), updated_at = NOW() WHERE mongo_achievement_id = $2`
		args = []interface{}{status, mongoID}
	case entity.StatusVerified:
		query = `UPDATE achievement_references SET status = $1, verified_at = NOW(), verified_by = $2, updated_at = NOW() WHERE mongo_achievement_id = $3`
		args = []interface{}{status, verifiedBy, mongoID}
	case entity.StatusRejected:
		query = `UPDATE achievement_references SET status = $1, rejection_note = $2, updated_at = NOW() WHERE mongo_achievement_id = $3`
		args = []interface{}{status, rejectionNote, mongoID}
	default:
		query = `UPDATE achievement_references SET status = $1, updated_at = NOW() WHERE mongo_achievement_id = $2`
		args = []interface{}{status, mongoID}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *AchievementRepository) DeleteReference(ctx context.Context, mongoID string) error {
	query := `DELETE FROM achievement_references WHERE mongo_achievement_id = $1`
	_, err := r.db.ExecContext(ctx, query, mongoID)
	return err
}

func (r *AchievementRepository) ListReferences(ctx context.Context, studentID *uuid.UUID, status string, limit, offset int) ([]*entity.AchievementReference, int, error) {
	var total int
	var countQuery, listQuery string
	var countArgs, listArgs []interface{}

	if studentID != nil && status != "" {
		countQuery = `SELECT COUNT(*) FROM achievement_references WHERE student_id = $1 AND status = $2`
		countArgs = []interface{}{studentID, status}
		listQuery = `
			SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
			FROM achievement_references WHERE student_id = $1 AND status = $2
			ORDER BY created_at DESC LIMIT $3 OFFSET $4
		`
		listArgs = []interface{}{studentID, status, limit, offset}
	} else if studentID != nil {
		countQuery = `SELECT COUNT(*) FROM achievement_references WHERE student_id = $1`
		countArgs = []interface{}{studentID}
		listQuery = `
			SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
			FROM achievement_references WHERE student_id = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`
		listArgs = []interface{}{studentID, limit, offset}
	} else if status != "" {
		countQuery = `SELECT COUNT(*) FROM achievement_references WHERE status = $1`
		countArgs = []interface{}{status}
		listQuery = `
			SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
			FROM achievement_references WHERE status = $1
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`
		listArgs = []interface{}{status, limit, offset}
	} else {
		countQuery = `SELECT COUNT(*) FROM achievement_references`
		countArgs = []interface{}{}
		listQuery = `
			SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
			FROM achievement_references
			ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`
		listArgs = []interface{}{limit, offset}
	}

	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var refs []*entity.AchievementReference
	for rows.Next() {
		ref := &entity.AchievementReference{}
		var submittedAt, verifiedAt sql.NullTime
		var verifiedBy, rejectionNote sql.NullString

		if err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&submittedAt, &verifiedAt, &verifiedBy, &rejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if submittedAt.Valid {
			ref.SubmittedAt = &submittedAt.Time
		}
		if verifiedAt.Valid {
			ref.VerifiedAt = &verifiedAt.Time
		}
		if verifiedBy.Valid {
			id, _ := uuid.Parse(verifiedBy.String)
			ref.VerifiedBy = &id
		}
		if rejectionNote.Valid {
			ref.RejectionNote = rejectionNote.String
		}

		refs = append(refs, ref)
	}

	return refs, total, nil
}

func (r *AchievementRepository) ListReferencesByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, status string, limit, offset int) ([]*entity.AchievementReference, int, error) {
	if len(studentIDs) == 0 {
		return []*entity.AchievementReference{}, 0, nil
	}

	// Build placeholders for IN clause
	placeholders := ""
	args := make([]interface{}, len(studentIDs))
	for i, id := range studentIDs {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "$" + string(rune('1'+i))
		args[i] = id
	}

	var total int
	countQuery := `SELECT COUNT(*) FROM achievement_references WHERE student_id IN (` + placeholders + `)`
	if status != "" {
		countQuery += ` AND status = $` + string(rune('1'+len(studentIDs)))
		args = append(args, status)
	}

	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Reset args for list query
	args = make([]interface{}, len(studentIDs))
	for i, id := range studentIDs {
		args[i] = id
	}

	listQuery := `
		SELECT id, student_id, mongo_achievement_id, status, submitted_at, verified_at, verified_by, rejection_note, created_at, updated_at
		FROM achievement_references WHERE student_id IN (` + placeholders + `)`
	
	argIndex := len(studentIDs) + 1
	if status != "" {
		listQuery += ` AND status = $` + string(rune('0'+argIndex))
		args = append(args, status)
		argIndex++
	}
	
	listQuery += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+argIndex)) + ` OFFSET $` + string(rune('0'+argIndex+1))
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var refs []*entity.AchievementReference
	for rows.Next() {
		ref := &entity.AchievementReference{}
		var submittedAt, verifiedAt sql.NullTime
		var verifiedBy, rejectionNote sql.NullString

		if err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status,
			&submittedAt, &verifiedAt, &verifiedBy, &rejectionNote,
			&ref.CreatedAt, &ref.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if submittedAt.Valid {
			ref.SubmittedAt = &submittedAt.Time
		}
		if verifiedAt.Valid {
			ref.VerifiedAt = &verifiedAt.Time
		}
		if verifiedBy.Valid {
			id, _ := uuid.Parse(verifiedBy.String)
			ref.VerifiedBy = &id
		}
		if rejectionNote.Valid {
			ref.RejectionNote = rejectionNote.String
		}

		refs = append(refs, ref)
	}

	return refs, total, nil
}

func (r *AchievementRepository) AddStatusHistory(ctx context.Context, history *entity.AchievementStatusHistory) error {
	query := `
		INSERT INTO achievement_status_history (id, achievement_ref_id, old_status, new_status, changed_by, note)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		history.ID, history.AchievementRefID, history.OldStatus, history.NewStatus, history.ChangedBy, history.Note,
	)
	return err
}

func (r *AchievementRepository) GetStatusHistory(ctx context.Context, achievementRefID uuid.UUID) ([]*entity.AchievementStatusHistory, error) {
	query := `
		SELECT id, achievement_ref_id, old_status, new_status, changed_by, note, created_at
		FROM achievement_status_history
		WHERE achievement_ref_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, achievementRefID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*entity.AchievementStatusHistory
	for rows.Next() {
		h := &entity.AchievementStatusHistory{}
		var oldStatus sql.NullString
		if err := rows.Scan(&h.ID, &h.AchievementRefID, &oldStatus, &h.NewStatus, &h.ChangedBy, &h.Note, &h.CreatedAt); err != nil {
			return nil, err
		}
		if oldStatus.Valid {
			h.OldStatus = entity.AchievementStatus(oldStatus.String)
		}
		history = append(history, h)
	}
	return history, nil
}

func (r *AchievementRepository) GetStatistics(ctx context.Context, studentID *uuid.UUID) (*entity.StatisticsResponse, error) {
	stats := &entity.StatisticsResponse{
		ByType:   make(map[string]int),
		ByStatus: make(map[string]int),
	}

	var statusQuery string
	var args []interface{}

	if studentID != nil {
		statusQuery = `
			SELECT status, COUNT(*) as count
			FROM achievement_references
			WHERE student_id = $1
			GROUP BY status
		`
		args = []interface{}{studentID}
	} else {
		statusQuery = `
			SELECT status, COUNT(*) as count
			FROM achievement_references
			GROUP BY status
		`
		args = []interface{}{}
	}

	rows, err := r.db.QueryContext(ctx, statusQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		stats.ByStatus[status] = count
		stats.TotalAchievements += count
		
		switch status {
		case "verified":
			stats.TotalVerified = count
		case "submitted":
			stats.TotalPending = count
		case "rejected":
			stats.TotalRejected = count
		}
	}

	return stats, nil
}
