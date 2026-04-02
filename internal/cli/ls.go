package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/repository"
	"github.com/spf13/cobra"
)

func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r < 127 {
			width++
		} else {
			width += 2
		}
	}
	return width
}

func truncateDisplay(s string, maxWidth int) string {
	if displayWidth(s) <= maxWidth {
		return s
	}

	result := ""
	width := 0
	for _, r := range s {
		runeWidth := 1
		if r >= 127 {
			runeWidth = 2
		}
		if width+runeWidth > maxWidth-3 {
			break
		}
		result += string(r)
		width += runeWidth
	}
	return result + "..."
}

var (
	listAll  bool
	listDone bool
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List quests",
	RunE:  runList,
}

func init() {
	lsCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show all quests")
	lsCmd.Flags().BoolVarP(&listDone, "done", "d", false, "Show completed quests only")
	rootCmd.AddCommand(lsCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Validate flags
	if listAll && listDone {
		fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: --all과 --done는 함께 사용할 수 없습니다.")
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

	// Get quests
	questRepo := repository.NewQuestRepository(repo)
	var quests []*domain.Quest

	switch {
	case listAll:
		quests, err = questRepo.ListAll()
	case listDone:
		quests, err = questRepo.ListDone()
	default:
		quests, err = questRepo.ListByStatus(domain.StatusTODO)
	}

	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 목록 조회 실패: %v\n", err)
		return ErrDatabase
	}

	// Empty list message
	if len(quests) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "퀘스트가 없습니다. 'ql add'로 새 퀘스트를 만들어보세요!")
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\t제목\t상태\t마감일")
	fmt.Fprintln(w, "----\t------\t----\t------")

	for _, q := range quests {
		dueStr := "-"
		if q.DueDate != nil {
			dueStr = q.DueDate.Format("01-02")
		}
		// Truncate title to max 35 display width to maintain table alignment
		truncatedTitle := truncateDisplay(q.Title, 35)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", q.ID, truncatedTitle, q.Status, dueStr)
	}

	w.Flush()
	return nil
}
