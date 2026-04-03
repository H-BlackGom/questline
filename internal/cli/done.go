package cli

import (
	"errors"
	"fmt"

	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/service"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <quest-id>",
	Short: "Complete a quest and earn XP",
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func init() {
	rootCmd.AddCommand(doneCmd)
}

func runDone(cmd *cobra.Command, args []string) error {
	questID := args[0]

	if questID == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 ID가 필요합니다.")
		return ErrInvalidInput
	}

	services, err := loadServices()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer services.Close()

	completion, err := services.Quest.CompleteQuest(questID)
	if err != nil {
		if errors.Is(err, service.ErrQuestNotFound) {
			fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트를 찾을 수 없습니다: %s\n", questID)
			return ErrQuestNotFound
		}
		if errors.Is(err, service.ErrQuestAlreadyCompleted) {
			fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 이미 완료된 퀘스트입니다.")
			return ErrAlreadyDone
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 완료 실패: %v\n", err)
		return ErrDatabase
	}

	// Output
	fmt.Fprintf(cmd.OutOrStdout(), "✓ 퀘스트 완료! +%d XP\n", completion.XPEarned)
	if completion.XPEarned > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "   Flow 배율: %.1fx\n", completion.FlowMultiplier)
	}
	if completion.ParentTransitionedToPending {
		fmt.Fprintln(cmd.OutOrStdout(), "   부모 퀘스트가 PENDING 상태로 전이되었습니다.")
	}

	// Check for level up
	if completion.LevelUpOccurred {
		fmt.Fprintf(cmd.OutOrStdout(), "\n🎉 레벨업! Lv.%d → Lv.%d\n", completion.LevelBefore, completion.LevelAfter)
		fmt.Fprintf(cmd.OutOrStdout(), "   칭호: %s\n", engine.GetTitle(completion.LevelAfter))
	}

	requiredXP := engine.GetRequiredXPForNextLevel(completion.LevelAfter)
	fmt.Fprintf(cmd.OutOrStdout(), "   다음 레벨까지: %d/%d XP\n", completion.XPAfter, requiredXP)

	return nil
}

var (
	ErrQuestNotFound = fmt.Errorf("quest not found")
	ErrNotFound      = ErrQuestNotFound
	ErrAlreadyDone   = fmt.Errorf("quest already completed")
)
