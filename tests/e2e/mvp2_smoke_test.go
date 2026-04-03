package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildBinary builds the ql binary for testing
func buildBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "ql")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/ql")
	cmd.Dir = "../.."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, string(output))
	}

	return binaryPath
}

// runCommand runs a ql command with a temp HOME directory
func runCommand(t *testing.T, binaryPath, homeDir string, args ...string) (string, string, error) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("HOME=%s", homeDir))
	
	output, err := cmd.CombinedOutput()
	return string(output), cmd.String(), err
}

// TestMVP2CoreWorkflow tests the complete MVP2 workflow:
// add daily -> add epic -> add sub with parent -> ls -> done -> me --flow
func TestMVP2CoreWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build binary
	binaryPath := buildBinary(t)
	homeDir := t.TempDir()

	// Step 1: Add a daily quest
	t.Run("AddDailyQuest", func(t *testing.T) {
		output, cmdStr, err := runCommand(t, binaryPath, homeDir, "add", "아침 스트레칭", "-t", "daily")
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		if !strings.Contains(output, "아침 스트레칭") {
			t.Errorf("Expected output to contain '아침 스트레칭', got: %s", output)
		}
	})

	time.Sleep(100 * time.Millisecond)

	// Step 2: Add an epic quest
	t.Run("AddEpicQuest", func(t *testing.T) {
		output, cmdStr, err := runCommand(t, binaryPath, homeDir, "add", "웹툰 연재", "-t", "epic")
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		if !strings.Contains(output, "웹툰 연재") {
			t.Errorf("Expected output to contain '웹툰 연재', got: %s", output)
		}
	})

	time.Sleep(100 * time.Millisecond)

	// Step 3: List quests to get the epic ID
	var epicID string
	t.Run("ListQuests", func(t *testing.T) {
		output, cmdStr, err := runCommand(t, binaryPath, homeDir, "ls")
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		if !strings.Contains(output, "웹툰 연재") {
			t.Errorf("Expected output to contain '웹툰 연재', got: %s", output)
		}
		
		// Extract epic ID from output
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "웹툰 연재") {
				fields := strings.Fields(line)
				if len(fields) > 0 {
					epicID = fields[0]
					break
				}
			}
		}
		if epicID == "" {
			t.Fatalf("Could not find epic ID in output: %s", output)
		}
	})

	time.Sleep(100 * time.Millisecond)

	// Step 4: Add a sub quest with parent
	t.Run("AddSubQuest", func(t *testing.T) {
		if epicID == "" {
			t.Skip("No epic ID available")
		}
		output, cmdStr, err := runCommand(t, binaryPath, homeDir, "add", "시나리오 초안", "-t", "sub", "-p", epicID)
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		if !strings.Contains(output, "시나리오 초안") {
			t.Errorf("Expected output to contain '시나리오 초안', got: %s", output)
		}
	})

	time.Sleep(100 * time.Millisecond)

	// Step 5: List all quests
	t.Run("ListAllQuests", func(t *testing.T) {
		output, cmdStr, err := runCommand(t, binaryPath, homeDir, "ls", "--all")
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		if !strings.Contains(output, "아침 스트레칭") {
			t.Errorf("Expected output to contain '아침 스트레칭', got: %s", output)
		}
		if !strings.Contains(output, "웹툰 연재") {
			t.Errorf("Expected output to contain '웹툰 연재', got: %s", output)
		}
		if !strings.Contains(output, "시나리오 초안") {
			t.Errorf("Expected output to contain '시나리오 초안', got: %s", output)
		}
	})

	time.Sleep(100 * time.Millisecond)

	// Step 6: Complete a quest
	var dailyID string
	t.Run("CompleteQuest", func(t *testing.T) {
		// First list to get ID
		output, cmdStr, err := runCommand(t, binaryPath, homeDir, "ls")
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "아침 스트레칭") {
				fields := strings.Fields(line)
				if len(fields) > 0 {
					dailyID = fields[0]
					break
				}
			}
		}
		if dailyID == "" {
			t.Fatalf("Could not find daily ID")
		}
		
		// Complete the quest
		output, cmdStr, err = runCommand(t, binaryPath, homeDir, "done", dailyID)
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		if !strings.Contains(output, "완료") && !strings.Contains(output, "complete") {
			t.Errorf("Expected output to contain completion message, got: %s", output)
		}
	})

	time.Sleep(100 * time.Millisecond)

	// Step 7: Check me --flow
	t.Run("CheckFlowStatus", func(t *testing.T) {
		output, cmdStr, err := runCommand(t, binaryPath, homeDir, "me", "--flow")
		if err != nil {
			t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
		}
		// Flow output should contain level or XP info
		if !strings.Contains(output, "Level") && !strings.Contains(output, "XP") && !strings.Contains(output, "레벨") {
			t.Errorf("Expected output to contain level/XP info, got: %s", output)
		}
	})
}

// TestQuestTypes tests all supported quest types
func TestQuestTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := buildBinary(t)
	homeDir := t.TempDir()

	questTypes := []struct {
		name  string
		typ   string
		title string
	}{
		{"Daily", "daily", "아침 운동"},
		{"Weekly", "weekly", "주간 회고"},
		{"Epic", "epic", "대형 프로젝트"},
		{"Guild", "guild", "사이드 프로젝트"},
	}

	for _, qt := range questTypes {
		t.Run(qt.name, func(t *testing.T) {
			output, cmdStr, err := runCommand(t, binaryPath, homeDir, "add", qt.title, "-t", qt.typ)
			if err != nil {
				t.Fatalf("Command failed: %v\nCmd: %s\nOutput: %s", err, cmdStr, output)
			}
			if !strings.Contains(output, qt.title) {
				t.Errorf("Expected output to contain '%s', got: %s", qt.title, output)
			}
		})
		time.Sleep(50 * time.Millisecond)
	}
}

// TestInvalidSubWithoutParent tests that sub quest without parent fails
func TestInvalidSubWithoutParent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	binaryPath := buildBinary(t)
	homeDir := t.TempDir()

	output, _, err := runCommand(t, binaryPath, homeDir, "add", "Invalid Sub", "-t", "sub")
	if err == nil {
		t.Errorf("Command should have failed but succeeded. Output: %s", output)
	}
	// Error message should mention parent
	if !strings.Contains(strings.ToLower(output), "parent") {
		t.Errorf("Expected error to mention 'parent', got: %s", output)
	}
}
