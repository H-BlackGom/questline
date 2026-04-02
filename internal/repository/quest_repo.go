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

const questSelectColumns = `id, title, status, due_date, created_at, completed_at, type, parent_id, scheduled_date, deleted_at`

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

	questType := quest.Type
	if questType == "" {
		questType = domain.QuestTypeDaily
	}

	_, err := qr.repo.db.Exec(
		`INSERT INTO quests (id, title, status, due_date, created_at, completed_at, type, parent_id, scheduled_date, deleted_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		quest.ID,
		quest.Title,
		string(quest.Status),
		dueDate,
		quest.CreatedAt.Format(time.RFC3339),
		nil,
		string(questType),
		quest.ParentID,
		nil,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create quest: %w", err)
	}
	return nil
}

// GetByID retrieves a quest by ID
func (qr *QuestRepository) GetByID(id string) (*domain.Quest, error) {
	row := qr.repo.db.QueryRow(
		`SELECT `+questSelectColumns+`
		 FROM quests WHERE id = ? AND deleted_at IS NULL`,
		id,
	)

	quest, err := scanQuest(row.Scan)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get quest: %w", err)
	}

	return quest, nil
}

// ListByStatus retrieves quests filtered by status
func (qr *QuestRepository) ListByStatus(status domain.QuestStatus) ([]*domain.Quest, error) {
	return qr.listQuests("WHERE status = ? AND deleted_at IS NULL ORDER BY created_at DESC, id ASC", string(status))
}

// ListAll retrieves all quests
func (qr *QuestRepository) ListAll() ([]*domain.Quest, error) {
	return qr.listQuests("WHERE deleted_at IS NULL ORDER BY created_at DESC, id ASC")
}

// ListDone retrieves completed quests
func (qr *QuestRepository) ListDone() ([]*domain.Quest, error) {
	return qr.listQuests("WHERE status = ? AND deleted_at IS NULL ORDER BY created_at DESC, id ASC", string(domain.StatusCompleted))
}

func (qr *QuestRepository) GetByType(questType domain.QuestType) ([]*domain.Quest, error) {
	return qr.listQuests("WHERE type = ? AND deleted_at IS NULL ORDER BY created_at ASC, id ASC", string(questType))
}

func (qr *QuestRepository) GetByParent(parentID string) ([]*domain.Quest, error) {
	return qr.listQuests("WHERE parent_id = ? AND deleted_at IS NULL ORDER BY created_at ASC, id ASC", parentID)
}

func (qr *QuestRepository) GetRootQuests() ([]*domain.Quest, error) {
	return qr.listQuests(`
		WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY CASE type
			WHEN 'daily' THEN 1
			WHEN 'weekly' THEN 2
			WHEN 'epic' THEN 3
			WHEN 'guild' THEN 4
			ELSE 5
		END, created_at ASC, id ASC`)
}

func (qr *QuestRepository) GetQuestTree() ([]*domain.Quest, error) {
	rows, err := qr.repo.db.Query(`
		WITH RECURSIVE tree AS (
			SELECT ` + questSelectColumns + `,
			       id AS root_id,
			       type AS root_type,
			       0 AS depth
			FROM quests
			WHERE parent_id IS NULL AND deleted_at IS NULL

			UNION ALL

			SELECT q.id, q.title, q.status, q.due_date, q.created_at, q.completed_at,
			       q.type, q.parent_id, q.scheduled_date, q.deleted_at,
			       tree.root_id,
			       tree.root_type,
			       tree.depth + 1
			FROM quests q
			JOIN tree ON q.parent_id = tree.id
			WHERE q.deleted_at IS NULL
		)
		SELECT id, title, status, due_date, created_at, completed_at, type, parent_id, scheduled_date, deleted_at
		FROM tree
		ORDER BY CASE root_type
			WHEN 'daily' THEN 1
			WHEN 'weekly' THEN 2
			WHEN 'epic' THEN 3
			WHEN 'guild' THEN 4
			ELSE 5
		END, root_id ASC, depth ASC, created_at ASC, id ASC;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch quest tree: %w", err)
	}
	defer rows.Close()

	var quests []*domain.Quest
	for rows.Next() {
		quest, err := scanQuest(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quest tree row: %w", err)
		}
		quests = append(quests, quest)
	}

	return quests, rows.Err()
}

func (qr *QuestRepository) ListByParent(parentID string) ([]*domain.Quest, error) {
	return qr.GetByParent(parentID)
}

func (qr *QuestRepository) UpdateStatus(questID string, status domain.QuestStatus, completedAt *time.Time) error {
	var completedValue interface{}
	if completedAt != nil {
		completedValue = completedAt.Format(time.RFC3339)
	}

	result, err := qr.repo.db.Exec(
		"UPDATE quests SET status = ?, completed_at = ? WHERE id = ? AND deleted_at IS NULL",
		string(status),
		completedValue,
		questID,
	)
	if err != nil {
		return fmt.Errorf("failed to update quest status: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to inspect updated quest rows: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (qr *QuestRepository) Delete(questID string) error {
	result, err := qr.repo.db.Exec(
		"UPDATE quests SET deleted_at = ? WHERE id = ? AND deleted_at IS NULL",
		time.Now().UTC().Format(time.RFC3339),
		questID,
	)
	if err != nil {
		return fmt.Errorf("failed to soft-delete quest: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to inspect deleted quest rows: %w", err)
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (qr *QuestRepository) listQuests(whereClause string, args ...interface{}) ([]*domain.Quest, error) {
	query := `SELECT ` + questSelectColumns + ` FROM quests ` + whereClause
	rows, err := qr.repo.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list quests: %w", err)
	}
	defer rows.Close()

	var quests []*domain.Quest
	for rows.Next() {
		quest, err := scanQuest(rows.Scan)
		if err != nil {
			return nil, err
		}

		quests = append(quests, quest)
	}

	return quests, rows.Err()
}

func scanQuest(scanFn func(dest ...interface{}) error) (*domain.Quest, error) {
	var quest domain.Quest
	var dueDate, createdAt, completedAt sql.NullString
	var parentID sql.NullString
	var scheduledDate sql.NullString
	var deletedAt sql.NullString

	err := scanFn(
		&quest.ID,
		&quest.Title,
		&quest.Status,
		&dueDate,
		&createdAt,
		&completedAt,
		&quest.Type,
		&parentID,
		&scheduledDate,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	if dueDate.Valid {
		t, _ := time.Parse("2006-01-02", dueDate.String)
		quest.DueDate = &t
	}
	if createdAt.Valid {
		if t, ok := parseDBTime(createdAt.String); ok {
			quest.CreatedAt = t
		}
	}
	if completedAt.Valid {
		if t, ok := parseDBTime(completedAt.String); ok {
			quest.CompletedAt = &t
		}
	}
	if parentID.Valid {
		pid := parentID.String
		quest.ParentID = &pid
	}
	if quest.Type == "" {
		quest.Type = domain.QuestTypeDaily
	}

	_ = scheduledDate
	_ = deletedAt

	return &quest, nil
}
