package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/repository"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   `add "<title>" [-d YYYY-MM-DD]`,
	Short: "Add a new quest",
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

var dueDate string

func init() {
	addCmd.Flags().StringVarP(&dueDate, "due", "d", "", "Due date (YYYY-MM-DD)")
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

	// Get database path
	dbPath := GetDBPath()

	// Initialize repository
	repo, err := repository.New(dbPath)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer repo.Close()

	// Create quest
	quest := &domain.Quest{
		ID:        uuid.New().String()[:8],
		Title:     title,
		Status:    domain.StatusTODO,
		DueDate:   due,
		CreatedAt: time.Now(),
	}

	// Save quest
	questRepo := repository.NewQuestRepository(repo)
	if err := questRepo.Create(quest); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 생성 실패: %v\n", err)
		return ErrDatabase
	}

	fmt.Fprintf(cmd.OutOrStdout(), "✓ 퀘스트 #%s 생성됨: \"%s\"\n", quest.ID, quest.Title)
	return nil
}

var (
	// ErrInvalidInput is returned for invalid user input
	ErrInvalidInput = fmt.Errorf("invalid input")
	// ErrDatabase is returned for database errors
	ErrDatabase = fmt.Errorf("database error")
)
