package service

import (
	"errors"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
)

var (
	ErrQuestNotFound         = errors.New("quest not found")
	ErrQuestAlreadyCompleted = errors.New("quest already completed")
)

type Services struct {
	repo   repositoryCloser
	Quest  QuestService
	Sync   SyncService
	Player PlayerService
}

type repositoryCloser interface {
	Close() error
}

func (s *Services) Close() error {
	if s == nil || s.repo == nil {
		return nil
	}
	return s.repo.Close()
}

type QuestService interface {
	CreateQuest(title string, questType domain.QuestType, parentID *string, dueDate *time.Time) (*domain.Quest, error)
	CompleteQuest(questID string) (*CompletionResult, error)
	GetQuest(questID string) (*domain.Quest, error)
	ListQuests(filter QuestFilter) ([]*domain.Quest, error)
	GetQuestTree() ([]*domain.Quest, error)
}

type QuestFilter struct {
	Types    []domain.QuestType
	Statuses []domain.QuestStatus
	ParentID *string
}

type CompletionResult struct {
	Quest                       *domain.Quest
	XPEarned                    int
	FlowMultiplier              float64
	XPBefore                    int
	XPAfter                     int
	LevelBefore                 int
	LevelAfter                  int
	LevelUpOccurred             bool
	ParentTransitionedToPending bool
}

type SyncService interface {
	EvaluateLazySync() (*SyncResult, error)
	CalculateFlowGrade(date time.Time) (*FlowGrade, error)
}

type SyncResult struct {
	EvaluatedAt time.Time
}

type FlowGrade struct {
	Date           time.Time
	CompletionRate float64
	Grade          domain.FlowStatus
	Multiplier     float64
}

type PlayerService interface {
	GetPlayer() (*domain.Player, error)
	AwardXP(baseXP int, questID string) (*AwardResult, error)
}

type AwardResult struct {
	XPBefore        int
	XPAfter         int
	LevelBefore     int
	LevelAfter      int
	LevelUpOccurred bool
}
