package repository

import (
	"context"

	"github.com/Aryma-f4/uas-backend/app/entity"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockAchievementRepository adalah mock implementation untuk testing
type MockAchievementRepository struct {
	achievements      map[primitive.ObjectID]*entity.Achievement
	references        map[uuid.UUID]*entity.AchievementReference
	statusHistory     map[uuid.UUID][]*entity.AchievementStatusHistory
}

// NewMockAchievementRepository membuat instance mock repository untuk testing
func NewMockAchievementRepository() *MockAchievementRepository {
	return &MockAchievementRepository{
		achievements:  make(map[primitive.ObjectID]*entity.Achievement),
		references:    make(map[uuid. UUID]*entity.AchievementReference),
		statusHistory: make(map[uuid.UUID][]*entity.AchievementStatusHistory),
	}
}

// MongoDB Operations - Mock Implementation

func (m *MockAchievementRepository) CreateMongo(ctx context.Context, achievement *entity.Achievement) (primitive.ObjectID, error) {
	id := primitive.NewObjectID()
	achievement.CreatedAt = primitive.NewDateTimeFromTime(ctx.Value("now").(interface{}).(time.Time))
	achievement.UpdatedAt = achievement.CreatedAt
	m.achievements[id] = achievement
	return id, nil
}

func (m *MockAchievementRepository) GetMongoByID(ctx context.Context, id primitive.ObjectID) (*entity.Achievement, error) {
	if ach, ok := m.achievements[id]; ok {
		return ach, nil
	}
	return nil, nil // Return nil jika tidak ditemukan (sesuai pola MongoDB)
}

func (m *MockAchievementRepository) UpdateMongo(ctx context.Context, id primitive.ObjectID, achievement *entity.Achievement) error {
	if _, ok := m.achievements[id]; ok {
		m.achievements[id] = achievement
		return nil
	}
	return nil
}

func (m *MockAchievementRepository) DeleteMongo(ctx context.Context, id primitive. ObjectID) error {
	delete(m.achievements, id)
	return nil
}

func (m *MockAchievementRepository) ListMongo(ctx context.Context, filter bson.M, limit, offset int64) ([]*entity.Achievement, error) {
	var results []*entity.Achievement
	for _, ach := range m.achievements {
		results = append(results, ach)
	}
	return results, nil
}

func (m *MockAchievementRepository) CountMongo(ctx context.Context, filter bson.M) (int64, error) {
	return int64(len(m.achievements)), nil
}

func (m *MockAchievementRepository) GetStatsMongo(ctx context.Context, studentIDs []uuid.UUID) (map[string]int, error) {
	stats := make(map[string]int)
	for _, ach := range m.achievements {
		stats[ach.AchievementType]++
	}
	return stats, nil
}

// PostgreSQL Operations - Mock Implementation

func (m *MockAchievementRepository) CreateReference(ctx context.Context, ref *entity.AchievementReference) error {
	m.references[ref.ID] = ref
	return nil
}

func (m *MockAchievementRepository) GetReferenceByMongoID(ctx context.Context, mongoID string) (*entity.AchievementReference, error) {
	for _, ref := range m.references {
		if ref.MongoAchievementID == mongoID {
			return ref, nil
		}
	}
	return nil, nil
}

func (m *MockAchievementRepository) UpdateReferenceStatus(ctx context.Context, mongoID string, status entity.AchievementStatus, verifiedBy *uuid.UUID, rejectionNote string) error {
	for _, ref := range m.references {
		if ref.MongoAchievementID == mongoID {
			ref.Status = status
			if verifiedBy != nil {
				ref.VerifiedBy = verifiedBy
			}
			if rejectionNote != "" {
				ref.RejectionNote = rejectionNote
			}
			return nil
		}
	}
	return nil
}

func (m *MockAchievementRepository) DeleteReference(ctx context.Context, mongoID string) error {
	for id, ref := range m.references {
		if ref.MongoAchievementID == mongoID {
			delete(m.references, id)
			return nil
		}
	}
	return nil
}

func (m *MockAchievementRepository) ListReferences(ctx context.Context, studentID *uuid.UUID, status string, limit, offset int) ([]*entity.AchievementReference, int, error) {
	var results []*entity.AchievementReference
	for _, ref := range m.references {
		if studentID != nil && ref.StudentID != *studentID {
			continue
		}
		if status != "" && ref.Status != status {
			continue
		}
		results = append(results, ref)
	}
	return results, len(results), nil
}

func (m *MockAchievementRepository) ListReferencesByStudentIDs(ctx context.Context, studentIDs []uuid.UUID, status string, limit, offset int) ([]*entity.AchievementReference, int, error) {
	var results []*entity.AchievementReference
	for _, ref := range m.references {
		found := false
		for _, id := range studentIDs {
			if ref.StudentID == id {
				found = true
				break
			}
		}
		if ! found {
			continue
		}
		if status != "" && ref.Status != status {
			continue
		}
		results = append(results, ref)
	}
	return results, len(results), nil
}

func (m *MockAchievementRepository) AddStatusHistory(ctx context.Context, history *entity.AchievementStatusHistory) error {
	m.statusHistory[history.AchievementRefID] = append(m.statusHistory[history.AchievementRefID], history)
	return nil
}

func (m *MockAchievementRepository) GetStatusHistory(ctx context.Context, achievementRefID uuid.UUID) ([]*entity.AchievementStatusHistory, error) {
	return m.statusHistory[achievementRefID], nil
}

func (m *MockAchievementRepository) GetStatistics(ctx context.Context, studentID *uuid.UUID) (*entity.StatisticsResponse, error) {
	stats := &entity.StatisticsResponse{
		ByType:    make(map[string]int),
		ByStatus: make(map[string]int),
	}
	return stats, nil
}
