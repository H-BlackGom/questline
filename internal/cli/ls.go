package cli

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/H-BlackGom/questline/internal/domain"
	"github.com/H-BlackGom/questline/internal/service"
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
	listType string
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List quests",
	RunE:  runList,
}

func init() {
	lsCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show all quests")
	lsCmd.Flags().BoolVarP(&listDone, "done", "d", false, "Show completed quests only")
	lsCmd.Flags().StringVar(&listType, "type", "", "Filter by quest type (daily, weekly, epic, guild, sub)")
	rootCmd.AddCommand(lsCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Validate flags
	if listAll && listDone {
		fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: --all과 --done는 함께 사용할 수 없습니다.")
		return ErrInvalidInput
	}

	services, err := loadServices()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 데이터베이스 초기화 실패: %v\n", err)
		return ErrDatabase
	}
	defer services.Close()

	var filter service.QuestFilter
	if strings.TrimSpace(listType) != "" {
		questType, err := parseQuestType(listType)
		if err != nil {
			fmt.Fprintln(cmd.ErrOrStderr(), "✗ 오류: 타입은 daily, weekly, epic, guild, sub 중 하나여야 합니다.")
			return ErrInvalidInput
		}
		filter.Types = []domain.QuestType{questType}
	}

	switch {
	case listAll:
		filter.Statuses = nil
	case listDone:
		filter.Statuses = []domain.QuestStatus{domain.StatusCompleted}
	default:
		filter.Statuses = []domain.QuestStatus{domain.StatusPending}
	}

	quests, err := services.Quest.ListQuests(filter)

	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "✗ 오류: 퀘스트 목록 조회 실패: %v\n", err)
		return ErrDatabase
	}

	if !listAll && !listDone && strings.TrimSpace(listType) == "" {
		sort.SliceStable(quests, func(i, j int) bool {
			return compareForRootOrdering(quests[i], quests[j])
		})
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

func compareForRootOrdering(left, right *domain.Quest) bool {
	leftRoot := left.ParentID == nil
	rightRoot := right.ParentID == nil
	if leftRoot != rightRoot {
		return leftRoot
	}

	leftRank := typeOrderRank(left.Type)
	rightRank := typeOrderRank(right.Type)
	if leftRank != rightRank {
		return leftRank < rightRank
	}
	if !left.CreatedAt.Equal(right.CreatedAt) {
		return left.CreatedAt.Before(right.CreatedAt)
	}
	return left.ID < right.ID
}

func typeOrderRank(questType domain.QuestType) int {
	switch questType {
	case domain.QuestTypeDaily:
		return 1
	case domain.QuestTypeWeekly:
		return 2
	case domain.QuestTypeEpic:
		return 3
	case domain.QuestTypeGuild:
		return 4
	default:
		return 5
	}
}
