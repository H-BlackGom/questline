package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/repository"
)

func TestListCommand(t *testing.T) {
	// Create temp directory for test DB
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Setup: Create quests
	os.RemoveAll(filepath.Join(tmpDir, ".questline"))

	repo, err := repository.New(filepath.Join(tmpDir, ".questline", "data.db"))
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	questRepo := repository.NewQuestRepository(repo)

	// Create TODO quest
	quest1 := &domain.Quest{
		ID:     "todo1234",
		Title:  "Todo Quest",
		Status: domain.StatusTODO,
	}
	questRepo.Create(quest1)

	// Create DONE quest
	quest2 := &domain.Quest{
		ID:     "done5678",
		Title:  "Done Quest",
		Status: domain.StatusDONE,
	}
	questRepo.Create(quest2)

	repo.Close()

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "List TODO quests",
			args:       []string{"ls"},
			wantErr:    false,
			wantOutput: "Todo Quest",
		},
		{
			name:       "List DONE quests",
			args:       []string{"ls", "--done"},
			wantErr:    false,
			wantOutput: "Done Quest",
		},
		{
			name:       "List all quests",
			args:       []string{"ls", "--all"},
			wantErr:    false,
			wantOutput: "ID",
		},
		{
			name:    "Conflicting flags",
			args:    []string{"ls", "--all", "--done"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			listAll = false
			listDone = false

			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantOutput != "" && !strings.Contains(buf.String(), tt.wantOutput) {
				t.Errorf("Output %q does not contain %q", buf.String(), tt.wantOutput)
			}
		})
	}
}
