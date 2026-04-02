package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestListCommand(t *testing.T) {
	// Create temp directory for test DB
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	todoTitle := "Todo Quest"
	doneTitle := "Done Quest"
	doneID := seedQuestViaAdd(t, doneTitle)
	if err := runCommandForTest([]string{"add", todoTitle}); err != nil {
		t.Fatalf("failed to seed todo quest: %v", err)
	}
	if err := runCommandForTest([]string{"done", doneID}); err != nil {
		t.Fatalf("failed to seed done quest: %v", err)
	}

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
