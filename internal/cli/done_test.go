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

func TestDoneCommand(t *testing.T) {
	// Create temp directory for test DB
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Setup: Create a quest first
	os.RemoveAll(filepath.Join(tmpDir, ".questline"))

	repo, err := repository.New(filepath.Join(tmpDir, ".questline", "data.db"))
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	questRepo := repository.NewQuestRepository(repo)
	quest := &domain.Quest{
		ID:     "test1234",
		Title:  "Test Quest",
		Status: domain.StatusTODO,
	}
	if err := questRepo.Create(quest); err != nil {
		t.Fatalf("Failed to create quest: %v", err)
	}
	repo.Close()

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "Complete existing quest",
			args:       []string{"done", "test1234"},
			wantErr:    false,
			wantOutput: "✓ 퀘스트 완료! +50 XP",
		},
		{
			name:    "Complete already done quest",
			args:    []string{"done", "test1234"},
			wantErr: true,
		},
		{
			name:    "Complete non-existent quest",
			args:    []string{"done", "deadbeef"},
			wantErr: true,
		},
		{
			name:    "Missing quest ID",
			args:    []string{"done"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
