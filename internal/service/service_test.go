package service

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
)

func TestBootstrap(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")

	services, err := Bootstrap(dbPath)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	defer services.Close()

	player, err := services.Player.GetPlayer()
	if err != nil {
		t.Fatalf("failed to get player after bootstrap: %v", err)
	}
	if player.LastSyncedAt == nil {
		t.Fatalf("expected lazy sync checkpoint to be initialized")
	}
}

func TestQuestServiceCreateAndComplete(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")

	services, err := Bootstrap(dbPath)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	defer services.Close()

	quest, err := services.Quest.CreateQuest("service quest", domain.QuestTypeDaily, nil, nil)
	if err != nil {
		t.Fatalf("create quest failed: %v", err)
	}

	result, err := services.Quest.CompleteQuest(quest.ID)
	if err != nil {
		t.Fatalf("complete quest failed: %v", err)
	}
	if result.XPAfter != 50 {
		t.Fatalf("expected xp after completion to be 50, got %d", result.XPAfter)
	}

	updatedQuest, err := services.Quest.GetQuest(quest.ID)
	if err != nil {
		t.Fatalf("failed to get updated quest: %v", err)
	}
	if updatedQuest.Status != domain.StatusCompleted {
		t.Fatalf("expected completed status, got %s", updatedQuest.Status)
	}

	_, err = services.Quest.CompleteQuest(quest.ID)
	if !errors.Is(err, ErrQuestAlreadyCompleted) {
		t.Fatalf("expected already-completed error, got %v", err)
	}

	_, err = services.Quest.CompleteQuest("missing")
	if !errors.Is(err, ErrQuestNotFound) {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestSyncServiceCalculateFlowGrade(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")

	services, err := Bootstrap(dbPath)
	if err != nil {
		t.Fatalf("bootstrap failed: %v", err)
	}
	defer services.Close()

	grade, err := services.Sync.CalculateFlowGrade(time.Now())
	if err != nil {
		t.Fatalf("calculate flow grade failed: %v", err)
	}
	if grade.Grade != domain.FlowStatusSmooth {
		t.Fatalf("expected default smooth grade when no dailies, got %s", grade.Grade)
	}
}
