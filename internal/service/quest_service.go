package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/repository"
	"github.com/google/uuid"
)

type questService struct {
	questRepo  *repository.QuestRepository
	playerRepo *repository.PlayerRepository
	db         *sql.DB
}

func NewQuestService(questRepo *repository.QuestRepository, playerRepo *repository.PlayerRepository, repo *repository.Repository) QuestService {
	return &questService{
		questRepo:  questRepo,
		playerRepo: playerRepo,
		db:         repo.DB(),
	}
}

func (s *questService) CreateQuest(title string, questType domain.QuestType, parentID *string, dueDate *time.Time) (*domain.Quest, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return nil, fmt.Errorf("title is required")
	}
	if !questType.IsValid() {
		return nil, fmt.Errorf("invalid quest type: %s", questType)
	}
	if questType == domain.QuestTypeSub && parentID == nil {
		return nil, fmt.Errorf("sub quest requires parent")
	}
	if questType != domain.QuestTypeSub && parentID != nil {
		return nil, fmt.Errorf("only sub quest can have parent")
	}
	if parentID != nil {
		parent, err := s.questRepo.GetByID(*parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrQuestNotFound
		}
		if parent.ParentID != nil {
			return nil, fmt.Errorf("third-depth sub quest is not supported")
		}
	}

	quest := &domain.Quest{
		ID:        uuid.New().String()[:8],
		Title:     trimmed,
		Status:    domain.StatusPending,
		Type:      questType,
		ParentID:  parentID,
		DueDate:   dueDate,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.questRepo.Create(quest); err != nil {
		return nil, err
	}
	return quest, nil
}

func (s *questService) CompleteQuest(questID string) (*CompletionResult, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to start completion transaction: %w", err)
	}
	defer tx.Rollback()

	var status string
	err = tx.QueryRow("SELECT status FROM quests WHERE id = ?", questID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrQuestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load quest status: %w", err)
	}
	if status == string(domain.StatusCompleted) {
		return nil, ErrQuestAlreadyCompleted
	}

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec("UPDATE quests SET status = ?, completed_at = ? WHERE id = ?", string(domain.StatusCompleted), now, questID); err != nil {
		return nil, fmt.Errorf("failed to update quest status: %w", err)
	}

	var level int
	var currentXP int
	var totalXPEarned int
	var questsCompleted int
	err = tx.QueryRow("SELECT level, current_xp, total_xp_earned, quests_completed FROM player WHERE id = 1").Scan(&level, &currentXP, &totalXPEarned, &questsCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to load player progression: %w", err)
	}

	leveling := engine.AddXP(level, currentXP, 50)
	if _, err := tx.Exec(
		`UPDATE player
		 SET level = ?, current_xp = ?, total_xp_earned = ?, quests_completed = ?, updated_at = datetime('now')
		 WHERE id = 1`,
		leveling.NewLevel,
		leveling.NewXP,
		totalXPEarned+50,
		questsCompleted+1,
	); err != nil {
		return nil, fmt.Errorf("failed to persist player progression: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit completion transaction: %w", err)
	}

	quest, err := s.questRepo.GetByID(questID)
	if err != nil {
		return nil, err
	}

	return &CompletionResult{
		Quest:           quest,
		XPBefore:        currentXP,
		XPAfter:         leveling.NewXP,
		LevelBefore:     level,
		LevelAfter:      leveling.NewLevel,
		LevelUpOccurred: leveling.LeveledUp,
	}, nil
}

func (s *questService) GetQuest(questID string) (*domain.Quest, error) {
	return s.questRepo.GetByID(questID)
}

func (s *questService) ListQuests(filter QuestFilter) ([]*domain.Quest, error) {
	if filter.ParentID != nil {
		return s.questRepo.ListByParent(*filter.ParentID)
	}
	if len(filter.Statuses) == 1 {
		return s.questRepo.ListByStatus(filter.Statuses[0])
	}
	return s.questRepo.ListAll()
}

func (s *questService) GetQuestTree() ([]*domain.Quest, error) {
	return s.questRepo.ListAll()
}
