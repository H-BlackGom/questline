package cli

import (
	"bytes"
	"regexp"
	"testing"
)

func seedQuestViaAdd(t *testing.T, title string) string {
	t.Helper()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"add", title})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("failed to seed quest via add command: %v", err)
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
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}
