package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestMeCommand(t *testing.T) {
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
			name:       "First run profile",
			args:       []string{"me"},
			wantErr:    false,
			wantOutput: "Lv.1",
		},
		{
			name:       "Profile shows Intern",
			args:       []string{"me"},
			wantErr:    false,
			wantOutput: "Intern",
		},
		{
			name:       "Profile shows XP",
			args:       []string{"me"},
			wantErr:    false,
			wantOutput: "누적 XP:",
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
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantOutput != "" && !strings.Contains(buf.String(), tt.wantOutput) {
				t.Errorf("Output %q does not contain %q", buf.String(), tt.wantOutput)
			}
		})
	}
}

func TestMeWithFlow(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	buf := new(bytes.Buffer)
	resetCLIFlags()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"me", "--flow"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("me --flow failed: %v", err)
	}

	output := buf.String()
	for _, label := range []string{"Flow 상태:", "Flow 배율:", "연속 일수:", "다음 평가 시점:"} {
		if !strings.Contains(output, label) {
			t.Fatalf("expected flow label %q in output: %s", label, output)
		}
	}
}
