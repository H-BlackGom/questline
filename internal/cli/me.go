package cli

import (
	"fmt"
	"strings"

	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/repository"
	"github.com/spf13/cobra"
)

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show player profile",
	RunE:  runMe,
}

func init() {
	rootCmd.AddCommand(meCmd)
}

func runMe(cmd *cobra.Command, args []string) error {
	// Get database path
	dbPath := GetDBPath()

	// Initialize repository
	repo, err := repository.New(dbPath)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer repo.Close()

	// Get player
	playerRepo := repository.NewPlayerRepository(repo)
	player, err := playerRepo.Get()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 플레이어 정보 조회 실패: %v\n", err)
		return ErrDatabase
	}

	requiredXP := engine.GetRequiredXPForNextLevel(player.Level)
	progress := getProgressBar(player.CurrentXP, requiredXP, 20)
	progressPercent := (player.CurrentXP * 100) / requiredXP

	// Print profile box
	fmt.Fprintln(cmd.OutOrStdout(), "╔══════════════════════════════════╗")
	fmt.Fprintln(cmd.OutOrStdout(), "║        퀘스트라인 캐릭터         ║")
	fmt.Fprintln(cmd.OutOrStdout(), "╠══════════════════════════════════╣")
	fmt.Fprintf(cmd.OutOrStdout(), "║  레벨: Lv.%d%s║\n", player.Level, strings.Repeat(" ", 24-len(fmt.Sprintf("%d", player.Level))))
	fmt.Fprintf(cmd.OutOrStdout(), "║  칭호: %s%s║\n", engine.GetTitle(player.Level), strings.Repeat(" ", 24-len(engine.GetTitle(player.Level))))
	fmt.Fprintf(cmd.OutOrStdout(), "║  누적 XP: %d%s║\n", player.TotalXPEarned, strings.Repeat(" ", 21-len(fmt.Sprintf("%d", player.TotalXPEarned))))
	fmt.Fprintln(cmd.OutOrStdout(), "║                                  ║")
	fmt.Fprintf(cmd.OutOrStdout(), "║  다음 레벨까지: %d/%d XP%s║\n", player.CurrentXP, requiredXP, strings.Repeat(" ", 15-len(fmt.Sprintf("%d/%d", player.CurrentXP, requiredXP))))
	fmt.Fprintf(cmd.OutOrStdout(), "║  [%s] %d%%%s║\n", progress, progressPercent, strings.Repeat(" ", 16-len(fmt.Sprintf("%d", progressPercent))))
	fmt.Fprintln(cmd.OutOrStdout(), "║                                  ║")
	fmt.Fprintf(cmd.OutOrStdout(), "║  완료한 퀘스트: %d개%s║\n", player.QuestsCompleted, strings.Repeat(" ", 20-len(fmt.Sprintf("%d", player.QuestsCompleted))))
	fmt.Fprintln(cmd.OutOrStdout(), "╚══════════════════════════════════╝")

	return nil
}

func getProgressBar(current, required, width int) string {
	if required == 0 {
		return strings.Repeat("█", width)
	}
	filled := (current * width) / required
	if filled > width {
		filled = width
	}
	empty := width - filled
	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}
