package cli

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/H-BlackGom/questline/internal/domain"
)

func resetCLIFlags() {
	dueDate = ""
	addQuestType = string(domain.QuestTypeDaily)
	addParentID = ""
	listAll = false
	listDone = false
	listType = ""
	meFlow = false
}

func seedQuestViaAdd(t *testing.T, title string) string {
	t.Helper()
	return seedQuestViaArgs(t, []string{"add", title})
}

func seedQuestViaArgs(t *testing.T, args []string) string {
	t.Helper()

	buf := new(bytes.Buffer)
	resetCLIFlags()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("failed to seed quest via args %v: %v", args, err)
	}

	re := regexp.MustCompile(`#([a-zA-Z0-9]{8})`)
	matches := re.FindStringSubmatch(buf.String())
	if len(matches) < 2 {
		t.Fatalf("failed to parse quest ID from output: %q", buf.String())
	}
	return matches[1]
}

func runCommandForTest(args []string) error {
	buf := new(bytes.Buffer)
	resetCLIFlags()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
