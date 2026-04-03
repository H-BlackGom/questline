package service

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
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
	if questType == domain.QuestTypeSub {
		if parentID == nil {
			return nil, fmt.Errorf("sub quest requires parent")
		}
		return s.CreateSubQuest(trimmed, *parentID, dueDate)
	}
	if parentID != nil {
		return nil, fmt.Errorf("only sub quest can have parent")
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

func (s *questService) CreateSubQuest(title string, parentID string, dueDate *time.Time) (*domain.Quest, error) {
	parent, err := s.questRepo.GetByID(parentID)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, ErrQuestNotFound
	}
	if parent.ParentID != nil {
		return nil, fmt.Errorf("third-depth sub quest is not supported")
	}
	if parent.Type != domain.QuestTypeEpic && parent.Type != domain.QuestTypeGuild {
		return nil, fmt.Errorf("sub quest parent must be epic or guild")
	}

	pid := parentID
	quest := &domain.Quest{
		ID:        uuid.New().String()[:8],
		Title:     strings.TrimSpace(title),
		Status:    domain.StatusPending,
		Type:      domain.QuestTypeSub,
		ParentID:  &pid,
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
	var parentID sql.NullString
	err = tx.QueryRow("SELECT status, parent_id FROM quests WHERE id = ? AND deleted_at IS NULL", questID).Scan(&status, &parentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrQuestNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load quest for completion: %w", err)
	}
	if status == string(domain.StatusCompleted) {
		return nil, ErrQuestAlreadyCompleted
	}

	now := time.Now().UTC()
	nowValue := now.Format(time.RFC3339)
	if _, err := tx.Exec("UPDATE quests SET status = ?, completed_at = ? WHERE id = ? AND deleted_at IS NULL", string(domain.StatusCompleted), nowValue, questID); err != nil {
		return nil, fmt.Errorf("failed to update quest status: %w", err)
	}

	questXP := 0
	flowMultiplier := 1.0
	parentTransitionedToPending := false
	var level int
	var currentXP int
	var totalXPEarned int
	var questsCompleted int
	leveling := engine.LevelingResult{NewLevel: 0, NewXP: 0, LeveledUp: false}

	err = tx.QueryRow("SELECT level, current_xp, total_xp_earned, quests_completed FROM player WHERE id = 1").Scan(&level, &currentXP, &totalXPEarned, &questsCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to load player progression: %w", err)
	}
	leveling = engine.LevelingResult{NewLevel: level, NewXP: currentXP, LeveledUp: false}

	if parentID.Valid {
		transitioned, err := s.transitionParentAfterSubCompletionTx(tx, parentID.String)
		if err != nil {
			return nil, err
		}
		parentTransitionedToPending = transitioned
	} else {
		allChildrenCompleted, hasChildren, err := s.hasOnlyCompletedChildrenTx(tx, questID)
		if err != nil {
			return nil, err
		}
		if hasChildren && !allChildrenCompleted {
			return nil, fmt.Errorf("cannot complete parent quest until all sub quests are completed")
		}

		if hasChildren {
			multiplier, err := s.flowMultiplierTx(tx, now)
			if err != nil {
				return nil, err
			}
			flowMultiplier = multiplier
			questXP = int(math.Round(50 * multiplier))
		} else {
			questXP = 50
		}
		if questXP < 1 {
			questXP = 1
		}

		leveling = engine.AddXP(level, currentXP, questXP)
		if _, err := tx.Exec(
			`UPDATE player
			 SET level = ?, current_xp = ?, total_xp_earned = ?, quests_completed = ?, updated_at = datetime('now')
			 WHERE id = 1`,
			leveling.NewLevel,
			leveling.NewXP,
			totalXPEarned+questXP,
			questsCompleted+1,
		); err != nil {
			return nil, fmt.Errorf("failed to persist player progression: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit completion transaction: %w", err)
	}

	quest, err := s.questRepo.GetByID(questID)
	if err != nil {
		return nil, err
	}

	return &CompletionResult{
		Quest:                       quest,
		XPEarned:                    questXP,
		FlowMultiplier:              flowMultiplier,
		XPBefore:                    currentXP,
		XPAfter:                     leveling.NewXP,
		LevelBefore:                 level,
		LevelAfter:                  leveling.NewLevel,
		LevelUpOccurred:             questXP > 0 && leveling.LeveledUp,
		ParentTransitionedToPending: parentTransitionedToPending,
	}, nil
}

func (s *questService) transitionParentAfterSubCompletionTx(tx *sql.Tx, parentID string) (bool, error) {
	allCompleted, hasChildren, err := s.hasOnlyCompletedChildrenTx(tx, parentID)
	if err != nil {
		return false, err
	}
	if !hasChildren {
		return false, nil
	}

	nextStatus := domain.StatusInProgress
	if allCompleted {
		nextStatus = domain.StatusPendingCompletion
	}

	if _, err := tx.Exec("UPDATE quests SET status = ?, completed_at = NULL WHERE id = ? AND deleted_at IS NULL", string(nextStatus), parentID); err != nil {
		return false, fmt.Errorf("failed to update parent quest status: %w", err)
	}
	return nextStatus == domain.StatusPendingCompletion, nil
}

func (s *questService) hasOnlyCompletedChildrenTx(tx *sql.Tx, parentID string) (allCompleted bool, hasChildren bool, err error) {
	var totalChildren int
	var completedChildren int
	err = tx.QueryRow(
		`SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0)
		 FROM quests
		 WHERE parent_id = ? AND deleted_at IS NULL`,
		string(domain.StatusCompleted),
		parentID,
	).Scan(&totalChildren, &completedChildren)
	if err != nil {
		return false, false, fmt.Errorf("failed to inspect child quest completion: %w", err)
	}
	if totalChildren == 0 {
		return false, false, nil
	}
	return totalChildren == completedChildren, true, nil
}

func (s *questService) flowMultiplierTx(tx *sql.Tx, date time.Time) (float64, error) {
	dateKey := date.Format("2006-01-02")

	var total int
	var completed int
	err := tx.QueryRow(
		`SELECT
			COUNT(*),
			COALESCE(SUM(CASE WHEN status = ? THEN 1 ELSE 0 END), 0)
		 FROM quests
		 WHERE type = ? AND (scheduled_date = ? OR date(created_at) = ?) AND deleted_at IS NULL`,
		string(domain.StatusCompleted),
		string(domain.QuestTypeDaily),
		dateKey,
		dateKey,
	).Scan(&total, &completed)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate flow multiplier: %w", err)
	}

	switch engine.EvaluateFlowStatus(total, completed) {
	case domain.FlowStatusSingularity:
		return 2.0, nil
	case domain.FlowStatusBurning:
		return 1.5, nil
	case domain.FlowStatusHazy:
		return 0.5, nil
	default:
		return 1.0, nil
	}
}

func (s *questService) GetQuest(questID string) (*domain.Quest, error) {
	return s.questRepo.GetByID(questID)
}

func (s *questService) ListQuests(filter QuestFilter) ([]*domain.Quest, error) {
	if len(filter.Types) == 1 && len(filter.Statuses) == 0 && filter.ParentID == nil {
		return s.questRepo.GetByType(filter.Types[0])
	}
	if filter.ParentID != nil {
		quests, err := s.questRepo.ListByParent(*filter.ParentID)
		if err != nil {
			return nil, err
		}
		return filterQuests(quests, filter), nil
	}
	if len(filter.Statuses) == 1 {
		quests, err := s.questRepo.ListByStatus(filter.Statuses[0])
		if err != nil {
			return nil, err
		}
		return filterQuests(quests, filter), nil
	}
	quests, err := s.questRepo.ListAll()
	if err != nil {
		return nil, err
	}
	return filterQuests(quests, filter), nil
}

func (s *questService) GetQuestTree() ([]*domain.Quest, error) {
	return s.questRepo.ListAll()
}

func filterQuests(quests []*domain.Quest, filter QuestFilter) []*domain.Quest {
	if len(filter.Types) == 0 && len(filter.Statuses) == 0 {
		return quests
	}

	typeSet := map[domain.QuestType]struct{}{}
	for _, questType := range filter.Types {
		typeSet[questType] = struct{}{}
	}
	statusSet := map[domain.QuestStatus]struct{}{}
	for _, status := range filter.Statuses {
		statusSet[status] = struct{}{}
	}

	filtered := make([]*domain.Quest, 0, len(quests))
	for _, quest := range quests {
		if len(typeSet) > 0 {
			if _, ok := typeSet[quest.Type]; !ok {
				continue
			}
		}
		if len(statusSet) > 0 {
			if _, ok := statusSet[quest.Status]; !ok {
				continue
			}
		}
		filtered = append(filtered, quest)
	}

	return filtered
}
