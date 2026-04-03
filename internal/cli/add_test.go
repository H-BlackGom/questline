package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddCommand(t *testing.T) {
	// Create temp directory for test DB
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "Add quest without due date",
			args:       []string{"add", "Buy milk"},
			wantErr:    false,
			wantOutput: "✓ 퀘스트 #",
		},
		{
			name:       "Add quest with due date",
			args:       []string{"add", "Write docs", "-d", "2026-12-31"},
			wantErr:    false,
			wantOutput: "✓ 퀘스트 #",
		},
		{
			name:    "Empty title",
			args:    []string{"add", ""},
			wantErr: true,
		},
		{
			name:    "Invalid date format",
			args:    []string{"add", "Buy milk", "-d", "2026/03/24"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear temp DB for each test
			os.RemoveAll(filepath.Join(tmpDir, ".questline"))

			buf := new(bytes.Buffer)
			resetCLIFlags()
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

func TestAddWithTypeAndParent(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	epicID := seedQuestViaArgs(t, []string{"add", "Epic Parent", "-t", "epic"})

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "Add sub quest with parent",
			args:       []string{"add", "Sub Quest", "-t", "sub", "-p", epicID},
			wantOutput: "✓ 퀘스트 #",
		},
		{
			name:    "Sub quest without parent",
			args:    []string{"add", "Sub Quest", "-t", "sub"},
			wantErr: true,
		},
		{
			name:    "Daily quest forbids parent",
			args:    []string{"add", "Daily Quest", "-t", "daily", "-p", epicID},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			resetCLIFlags()
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)

			err := rootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v, output=%q", err, tt.wantErr, buf.String())
			}
			if tt.wantOutput != "" && !strings.Contains(buf.String(), tt.wantOutput) {
				t.Fatalf("Output %q does not contain %q", buf.String(), tt.wantOutput)
			}
		})
	}
}
