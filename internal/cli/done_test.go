package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDoneCommand(t *testing.T) {
	// Create temp directory for test DB
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	questID := seedQuestViaAdd(t, "Test Quest")

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "Complete existing quest",
			args:       []string{"done", questID},
			wantErr:    false,
			wantOutput: "✓ 퀘스트 완료! +50 XP",
		},
		{
			name:    "Complete already done quest",
			args:    []string{"done", questID},
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
