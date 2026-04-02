package cli

import (
	"bytes"
	"testing"

	"github.com/H-BlackGom/questline/internal/service"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeTeaProgram struct {
	run func() (tea.Model, error)
}

func (f fakeTeaProgram) Run() (tea.Model, error) {
	return f.run()
}

func TestCheckCommandStartsProgram(t *testing.T) {
	originalBootstrap := bootstrap
	originalNewCheckProgram := newCheckProgram
	defer func() {
		bootstrap = originalBootstrap
		newCheckProgram = originalNewCheckProgram
	}()

	bootstrapCalls := 0
	bootstrap = func(_ string) (*service.Services, error) {
		bootstrapCalls++
		return &service.Services{
			Quest:  &fakeQuestService{},
			Player: &fakePlayerService{},
			Sync:   &fakeSyncService{},
		}, nil
	}

	runCalled := false
	newCheckProgram = func(model tea.Model, _ ...tea.ProgramOption) teaProgram {
		return fakeTeaProgram{run: func() (tea.Model, error) {
			runCalled = true
			return model, nil
		}}
	}

	buf := new(bytes.Buffer)
	resetCLIFlags()
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"check"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected check command to start program, got error: %v", err)
	}
	if bootstrapCalls != 1 {
		t.Fatalf("expected bootstrap to be called once, got %d", bootstrapCalls)
	}
	if !runCalled {
		t.Fatal("expected check command to run the Bubble Tea program")
	}
}
