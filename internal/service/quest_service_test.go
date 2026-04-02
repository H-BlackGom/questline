package service

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/H-BlackGom/questline/internal/domain"
)

func TestCreateSubQuestValidation(t *testing.T) {
	services := newQuestServiceTestBootstrap(t)
	defer services.Close()

	parent, err := services.Quest.CreateQuest("epic parent", domain.QuestTypeEpic, nil, nil)
	if err != nil {
		t.Fatalf("create parent quest failed: %v", err)
	}

	sub, err := services.Quest.CreateQuest("epic sub", domain.QuestTypeSub, &parent.ID, nil)
	if err != nil {
		t.Fatalf("create sub quest failed: %v", err)
	}
	if sub.Type != domain.QuestTypeSub {
		t.Fatalf("expected sub type, got %s", sub.Type)
	}
	if sub.ParentID == nil || *sub.ParentID != parent.ID {
		t.Fatalf("expected parent id %s, got %v", parent.ID, sub.ParentID)
	}
}

func TestCreateSubQuestRejectsDailyParent(t *testing.T) {
	services := newQuestServiceTestBootstrap(t)
	defer services.Close()

	parent, err := services.Quest.CreateQuest("daily parent", domain.QuestTypeDaily, nil, nil)
	if err != nil {
		t.Fatalf("create daily parent failed: %v", err)
	}

	_, err = services.Quest.CreateQuest("invalid sub", domain.QuestTypeSub, &parent.ID, nil)
	if err == nil {
		t.Fatalf("expected validation error for daily parent")
	}
	if !strings.Contains(err.Error(), "epic or guild") {
		t.Fatalf("expected epic/guild validation error, got %v", err)
	}
}

func TestCreateSubQuestRejectsThirdDepth(t *testing.T) {
	services := newQuestServiceTestBootstrap(t)
	defer services.Close()

	parent, err := services.Quest.CreateQuest("guild parent", domain.QuestTypeGuild, nil, nil)
	if err != nil {
		t.Fatalf("create guild parent failed: %v", err)
	}

	child, err := services.Quest.CreateQuest("sub child", domain.QuestTypeSub, &parent.ID, nil)
	if err != nil {
		t.Fatalf("create first depth sub failed: %v", err)
	}

	_, err = services.Quest.CreateQuest("sub grandchild", domain.QuestTypeSub, &child.ID, nil)
	if err == nil {
		t.Fatalf("expected third-depth validation error")
	}
	if !strings.Contains(err.Error(), "third-depth") {
		t.Fatalf("expected third-depth validation error, got %v", err)
	}
}

func TestCompleteSubQuestTransitionsParent(t *testing.T) {
	services := newQuestServiceTestBootstrap(t)
	defer services.Close()

	parent, err := services.Quest.CreateQuest("guild parent", domain.QuestTypeGuild, nil, nil)
	if err != nil {
		t.Fatalf("create parent quest failed: %v", err)
	}

	subA, err := services.Quest.CreateQuest("sub A", domain.QuestTypeSub, &parent.ID, nil)
	if err != nil {
		t.Fatalf("create sub A failed: %v", err)
	}
	subB, err := services.Quest.CreateQuest("sub B", domain.QuestTypeSub, &parent.ID, nil)
	if err != nil {
		t.Fatalf("create sub B failed: %v", err)
	}

	if _, err := services.Quest.CompleteQuest(subA.ID); err != nil {
		t.Fatalf("complete sub A failed: %v", err)
	}

	updatedParent, err := services.Quest.GetQuest(parent.ID)
	if err != nil {
		t.Fatalf("get parent after first sub completion failed: %v", err)
	}
	if updatedParent.Status != domain.StatusInProgress {
		t.Fatalf("expected parent status in_progress after first sub, got %s", updatedParent.Status)
	}

	player, err := services.Player.GetPlayer()
	if err != nil {
		t.Fatalf("get player after first sub completion failed: %v", err)
	}
	if player.TotalXPEarned != 0 {
		t.Fatalf("expected no xp awarded for sub completion, got %d", player.TotalXPEarned)
	}

	if _, err := services.Quest.CompleteQuest(subB.ID); err != nil {
		t.Fatalf("complete sub B failed: %v", err)
	}

	updatedParent, err = services.Quest.GetQuest(parent.ID)
	if err != nil {
		t.Fatalf("get parent after all sub completion failed: %v", err)
	}
	if updatedParent.Status != domain.StatusPendingCompletion {
		t.Fatalf("expected parent status pending_completion after all sub, got %s", updatedParent.Status)
	}

	player, err = services.Player.GetPlayer()
	if err != nil {
		t.Fatalf("get player after all sub completion failed: %v", err)
	}
	if player.TotalXPEarned != 0 {
		t.Fatalf("expected no xp awarded before parent confirmation, got %d", player.TotalXPEarned)
	}

	if _, err := services.Sync.EvaluateLazySync(); err != nil {
		t.Fatalf("lazy sync failed: %v", err)
	}

	updatedParent, err = services.Quest.GetQuest(parent.ID)
	if err != nil {
		t.Fatalf("get parent after lazy sync failed: %v", err)
	}
	if updatedParent.Status != domain.StatusPendingCompletion {
		t.Fatalf("expected pending_completion to be preserved across lazy sync, got %s", updatedParent.Status)
	}
}

func TestCompleteParentPendingAwardsXP(t *testing.T) {
	services := newQuestServiceTestBootstrap(t)
	defer services.Close()

	daily, err := services.Quest.CreateQuest("daily for burning flow", domain.QuestTypeDaily, nil, nil)
	if err != nil {
		t.Fatalf("create daily quest failed: %v", err)
	}
	if _, err := services.Quest.CompleteQuest(daily.ID); err != nil {
		t.Fatalf("complete daily quest failed: %v", err)
	}

	parent, err := services.Quest.CreateQuest("epic parent", domain.QuestTypeEpic, nil, nil)
	if err != nil {
		t.Fatalf("create parent quest failed: %v", err)
	}
	sub, err := services.Quest.CreateQuest("sub", domain.QuestTypeSub, &parent.ID, nil)
	if err != nil {
		t.Fatalf("create sub quest failed: %v", err)
	}
	if _, err := services.Quest.CompleteQuest(sub.ID); err != nil {
		t.Fatalf("complete sub quest failed: %v", err)
	}

	parentBeforeComplete, err := services.Quest.GetQuest(parent.ID)
	if err != nil {
		t.Fatalf("get parent before completion failed: %v", err)
	}
	if parentBeforeComplete.Status != domain.StatusPendingCompletion {
		t.Fatalf("expected parent to be pending_completion before manual confirm, got %s", parentBeforeComplete.Status)
	}

	completion, err := services.Quest.CompleteQuest(parent.ID)
	if err != nil {
		t.Fatalf("complete parent quest failed: %v", err)
	}
	if completion.XPBefore != 50 {
		t.Fatalf("expected xp before parent completion to be 50, got %d", completion.XPBefore)
	}
	if completion.XPAfter != 125 {
		t.Fatalf("expected burning flow multiplier XP result (125), got %d", completion.XPAfter)
	}

	player, err := services.Player.GetPlayer()
	if err != nil {
		t.Fatalf("get player after parent completion failed: %v", err)
	}
	if player.TotalXPEarned != 125 {
		t.Fatalf("expected total xp earned to be 125, got %d", player.TotalXPEarned)
	}
	if player.QuestsCompleted != 2 {
		t.Fatalf("expected quests completed to be 2 (daily + parent), got %d", player.QuestsCompleted)
	}
}

func newQuestServiceTestBootstrap(t *testing.T) *Services {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "data.db")
	services, err := Bootstrap(dbPath)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	return services
}
