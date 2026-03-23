package cli

import (
	"fmt"

	"github.com/H-BlackGom/questline/internal/engine"
	"github.com/H-BlackGom/questline/internal/repository"
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

	// Get database path
	dbPath := GetDBPath()

	// Initialize repository
	repo, err := repository.New(dbPath)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer repo.Close()

	// Get player before completing quest
	playerRepo := repository.NewPlayerRepository(repo)
	player, err := playerRepo.Get()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 플레이어 정보 조회 실패: %v\n", err)
		return ErrDatabase
	}

	oldLevel := player.Level

	// Complete quest and award XP
	err = playerRepo.CompleteQuest(questID, 50)
	if err != nil {
		if err.Error() == "quest not found" {
			fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트를 찾을 수 없습니다: %s\n", questID)
			return ErrNotFound
		}
		if err.Error() == "quest already completed" {
			fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 이미 완료된 퀘스트입니다.")
			return ErrAlreadyDone
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 완료 실패: %v\n", err)
		return ErrDatabase
	}

	// Get updated player
	player, err = playerRepo.Get()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 플레이어 정보 조회 실패: %v\n", err)
		return ErrDatabase
	}

	// Output
	fmt.Fprintln(cmd.OutOrStdout(), "✓ 퀘스트 완료! +50 XP")

	// Check for level up
	if player.Level > oldLevel {
		fmt.Fprintf(cmd.OutOrStdout(), "\n🎉 레벨업! Lv.%d → Lv.%d\n", oldLevel, player.Level)
		fmt.Fprintf(cmd.OutOrStdout(), "   칭호: %s\n", engine.GetTitle(player.Level))
	}

	requiredXP := engine.GetRequiredXPForNextLevel(player.Level)
	fmt.Fprintf(cmd.OutOrStdout(), "   다음 레벨까지: %d/%d XP\n", player.CurrentXP, requiredXP)

	return nil
}

var (
	// ErrNotFound is returned when quest is not found
	ErrNotFound = fmt.Errorf("quest not found")
	// ErrAlreadyDone is returned when quest is already completed
	ErrAlreadyDone = fmt.Errorf("quest already completed")
)
