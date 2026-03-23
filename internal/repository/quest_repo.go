package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
)

// QuestRepository provides quest CRUD operations
type QuestRepository struct {
	repo *Repository
}

// NewQuestRepository creates a new QuestRepository
func NewQuestRepository(repo *Repository) *QuestRepository {
	return &QuestRepository{repo: repo}
}

// Create inserts a new quest
func (qr *QuestRepository) Create(quest *domain.Quest) error {
	var dueDate interface{}
	if quest.DueDate != nil {
		dueDate = quest.DueDate.Format("2006-01-02")
	}

	_, err := qr.repo.db.Exec(
		`INSERT INTO quests (id, title, status, due_date, created_at, completed_at) 
		 VALUES (?, ?, ?, ?, ?, ?)`,
		quest.ID,
		quest.Title,
		string(quest.Status),
		dueDate,
		quest.CreatedAt.Format(time.RFC3339),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create quest: %w", err)
	}
	return nil
}

// GetByID retrieves a quest by ID
func (qr *QuestRepository) GetByID(id string) (*domain.Quest, error) {
	var quest domain.Quest
	var dueDate, completedAt sql.NullString

	err := qr.repo.db.QueryRow(
		`SELECT id, title, status, due_date, created_at, completed_at 
		 FROM quests WHERE id = ?`, id,
	).Scan(
		&quest.ID,
		&quest.Title,
		&quest.Status,
		&dueDate,
		&quest.CreatedAt,
		&completedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quest: %w", err)
	}

	if dueDate.Valid {
		t, _ := time.Parse("2006-01-02", dueDate.String)
		quest.DueDate = &t
	}
	if completedAt.Valid {
		t, _ := time.Parse(time.RFC3339, completedAt.String)
		quest.CompletedAt = &t
	}

	return &quest, nil
}

// ListByStatus retrieves quests filtered by status
func (qr *QuestRepository) ListByStatus(status domain.Status) ([]*domain.Quest, error) {
	return qr.listQuests("WHERE status = ? ORDER BY created_at DESC, id ASC", string(status))
}

// ListAll retrieves all quests
func (qr *QuestRepository) ListAll() ([]*domain.Quest, error) {
	return qr.listQuests("ORDER BY created_at DESC, id ASC")
}

// ListDone retrieves completed quests
func (qr *QuestRepository) ListDone() ([]*domain.Quest, error) {
	return qr.listQuests("WHERE status = 'DONE' ORDER BY created_at DESC, id ASC")
}

func (qr *QuestRepository) listQuests(whereClause string, args ...interface{}) ([]*domain.Quest, error) {
	query := `SELECT id, title, status, due_date, created_at, completed_at FROM quests ` + whereClause
	rows, err := qr.repo.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list quests: %w", err)
	}
	defer rows.Close()

	var quests []*domain.Quest
	for rows.Next() {
		var quest domain.Quest
		var dueDate, createdAt, completedAt sql.NullString

		err := rows.Scan(
			&quest.ID,
			&quest.Title,
			&quest.Status,
			&dueDate,
			&createdAt,
			&completedAt,
		)
		if err != nil {
			return nil, err
		}

		if dueDate.Valid {
			t, _ := time.Parse("2006-01-02", dueDate.String)
			quest.DueDate = &t
		}
		if createdAt.Valid {
			quest.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		}
		if completedAt.Valid {
			t, _ := time.Parse(time.RFC3339, completedAt.String)
			quest.CompletedAt = &t
		}

		quests = append(quests, &quest)
	}

	return quests, rows.Err()
}
