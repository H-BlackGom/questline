package cli

import (
	"fmt"

	"github.com/H-BlackGom/questline/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

type teaProgram interface {
	Run() (tea.Model, error)
}

var newCheckProgram = func(model tea.Model, opts ...tea.ProgramOption) teaProgram {
	return tea.NewProgram(model, opts...)
}

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Open the quest dashboard",
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	services, err := loadServices()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer services.Close()

	model := tui.NewModel(nil, nil).WithServices(services)
	model.Loading = true

	program := newCheckProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: TUI 실행 실패: %v\n", err)
		return err
	}

	return nil
}
