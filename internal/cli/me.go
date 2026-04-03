package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/spf13/cobra"
)

var meFlow bool

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Show player profile",
	RunE:  runMe,
}

func init() {
	meCmd.Flags().BoolVar(&meFlow, "flow", false, "Show detailed flow status")
	rootCmd.AddCommand(meCmd)
}

func runMe(cmd *cobra.Command, args []string) error {
	services, err := loadServices()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer services.Close()

	player, err := services.Player.GetPlayer()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 플레이어 정보 조회 실패: %v\n", err)
		return ErrDatabase
	}

	requiredXP := engine.GetRequiredXPForNextLevel(player.Level)
	progress := getProgressBar(player.CurrentXP, requiredXP, 20)
	progressPercent := (player.CurrentXP * 100) / requiredXP

	boxWidth := 32

	fmt.Fprintln(cmd.OutOrStdout(), "╔══════════════════════════════════╗")
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineCenter("퀘스트라인 캐릭터", boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), "╠══════════════════════════════════╣")
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("레벨: Lv.%d", player.Level), boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("칭호: %s", engine.GetTitle(player.Level)), boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("누적 XP: %d", player.TotalXPEarned), boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft("", boxWidth))
	if meFlow {
		flowStatus := player.FlowStatus
		if !flowStatus.IsValid() {
			flowStatus = domain.FlowStatusSmooth
		}
		multiplier := flowMultiplierByStatus(flowStatus)
		nextEval := nextFlowEvaluation(time.Now())
		fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("Flow 상태: %s", flowStatusLabel(flowStatus)), boxWidth))
		fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("Flow 배율: %.1fx", multiplier), boxWidth))
		fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("연속 일수: %d일", player.StreakDays), boxWidth))
		fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("다음 평가 시점: %s", nextEval.Format("2006-01-02 15:04")), boxWidth))
		fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft("", boxWidth))
	}
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("다음 레벨까지: %d/%d XP", player.CurrentXP, requiredXP), boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("[%s] %d%%", progress, progressPercent), boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft("", boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), formatBoxLineLeft(fmt.Sprintf("완료한 퀘스트: %d개", player.QuestsCompleted), boxWidth))
	fmt.Fprintln(cmd.OutOrStdout(), "╚══════════════════════════════════╝")

	return nil
}

func formatBoxLineLeft(content string, boxWidth int) string {
	contentWidth := displayWidth(content)
	padding := boxWidth - contentWidth - 2
	if padding < 0 {
		padding = 0
	}
	return "║  " + content + strings.Repeat(" ", padding) + "║"
}

func formatBoxLineCenter(content string, boxWidth int) string {
	contentWidth := displayWidth(content)
	padding := boxWidth - contentWidth
	if padding < 0 {
		padding = 0
	}
	leftPad := padding / 2
	rightPad := padding - leftPad
	return "║" + strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad) + "║"
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

func flowMultiplierByStatus(status domain.FlowStatus) float64 {
	switch status {
	case domain.FlowStatusSingularity:
		return 2.0
	case domain.FlowStatusBurning:
		return 1.5
	case domain.FlowStatusHazy:
		return 0.5
	default:
		return 1.0
	}
}

func flowStatusLabel(status domain.FlowStatus) string {
	switch status {
	case domain.FlowStatusSingularity:
		return "✨ SINGULARITY"
	case domain.FlowStatusBurning:
		return "🔥 BURNING"
	case domain.FlowStatusHazy:
		return "🌫️ HAZY"
	default:
		return "🌊 SMOOTH"
	}
}

func nextFlowEvaluation(now time.Time) time.Time {
	locNow := now.Local()
	next := time.Date(locNow.Year(), locNow.Month(), locNow.Day(), 4, 0, 0, 0, locNow.Location())
	if !locNow.Before(next) {
		next = next.Add(24 * time.Hour)
	}
	return next
}
