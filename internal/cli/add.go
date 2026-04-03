package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   `add "<title>" [-t daily|weekly|epic|guild|sub] [-p parent-id] [-d YYYY-MM-DD]`,
	Short: "Add a new quest",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

var (
	dueDate      string
	addQuestType string
	addParentID  string
)

func init() {
	addCmd.Flags().StringVarP(&dueDate, "due", "d", "", "Due date (YYYY-MM-DD)")
	addCmd.Flags().StringVarP(&addQuestType, "type", "t", string(domain.QuestTypeGuild), "Quest type (daily, weekly, epic, guild, sub)")
	addCmd.Flags().StringVarP(&addParentID, "parent", "p", "", "Parent quest ID (required for sub)")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	title := strings.TrimSpace(args[0])

	// Validate title
	if title == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 제목이 필요합니다.")
		return ErrInvalidInput
	}
	if len(title) > 200 {
		fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 제목은 200자 이하여야 합니다.")
		return ErrInvalidInput
	}

	questType, err := parseQuestType(addQuestType)
	if err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 타입은 daily, weekly, epic, guild, sub 중 하나여야 합니다.")
		return ErrInvalidInput
	}

	parentIDText := strings.TrimSpace(addParentID)
	var parentID *string
	if questType == domain.QuestTypeSub {
		if parentIDText == "" {
			fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: sub 타입은 부모 퀘스트 ID가 필요합니다.")
			return ErrInvalidInput
		}
		parentID = &parentIDText
	} else if parentIDText != "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: sub 타입이 아닌 퀘스트에는 부모를 지정할 수 없습니다.")
		return ErrInvalidInput
	}

	// Parse due date
	var due *time.Time
	if dueDate != "" {
		t, err := time.Parse("2006-01-02", dueDate)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 날짜 형식이 올바르지 않습니다. (YYYY-MM-DD)")
			return ErrInvalidInput
		}
		// Check if date is today or later
		today := time.Now().Truncate(24 * time.Hour)
		if t.Before(today) {
			fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 마감일은 오늘 또는 미래여야 합니다.")
			return ErrInvalidInput
		}
		due = &t
	}

	services, err := loadServices()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer services.Close()

	quest, err := services.Quest.CreateQuest(
		title,
		questType,
		parentID,
		due,
	)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 생성 실패: %v\n", err)
		return ErrDatabase
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ 퀘스트 #%s 생성됨: \"%s\"\n", quest.ID, quest.Title)
	return nil
}

func parseQuestType(raw string) (domain.QuestType, error) {
	questType := domain.QuestType(strings.ToLower(strings.TrimSpace(raw)))
	if !questType.IsValid() {
		return "", fmt.Errorf("invalid quest type")
	}
	return questType, nil
}

var (
	// ErrInvalidInput is returned for invalid user input
	ErrInvalidInput = fmt.Errorf("invalid input")
	// ErrDatabase is returned for database errors
	ErrDatabase = fmt.Errorf("database error")
)
