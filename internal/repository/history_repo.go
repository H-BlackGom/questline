package repository

import (
	"fmt"
	"time"

	"github.com/H-BlackGom/questline/internal/domain"
)

type HistoryRepository struct {
	repo *Repository
}

type DailyStats struct {
	Date              string
	TotalRoutines     int
	CompletedRoutines int
	TotalXPEarned     int
}

type DailyEvaluation struct {
	Date              string
	TotalRoutines     int
	CompletedRoutines int
	CompletionRate    float64
	FlowGrade         domain.FlowStatus
	EvaluatedAt       time.Time
}

func NewHistoryRepository(repo *Repository) *HistoryRepository {
	return &HistoryRepository{repo: repo}
}

func (hr *HistoryRepository) SaveQuestHistory(questID, date string, status domain.QuestStatus, completed bool, xpEarned int) error {
	_, err := hr.repo.db.Exec(
		`INSERT INTO quest_history (quest_id, date, status, completed, xp_earned)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(quest_id, date)
		 DO UPDATE SET status = excluded.status, completed = excluded.completed, xp_earned = excluded.xp_earned`,
		questID,
		date,
		string(status),
		completed,
		xpEarned,
	)
	if err != nil {
		return fmt.Errorf("failed to save quest history: %w", err)
	}
	return nil
}

func (hr *HistoryRepository) GetDailyStats(date string) (*DailyStats, error) {
	stats := &DailyStats{Date: date}
	err := hr.repo.db.QueryRow(
		`SELECT COUNT(*),
		        SUM(CASE WHEN completed THEN 1 ELSE 0 END),
		        COALESCE(SUM(xp_earned), 0)
		 FROM quest_history
		 WHERE date = ?`,
		date,
	).Scan(&stats.TotalRoutines, &stats.CompletedRoutines, &stats.TotalXPEarned)
	if err != nil {
		return nil, fmt.Errorf("failed to read daily stats: %w", err)
	}
	return stats, nil
}

func (hr *HistoryRepository) SaveDailyEvaluation(date string, totalRoutines, completedRoutines int, completionRate float64, flowGrade domain.FlowStatus) error {
	_, err := hr.repo.db.Exec(
		`INSERT INTO daily_evaluation (date, total_routines, completed_routines, completion_rate, flow_grade)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(date)
		 DO UPDATE SET total_routines = excluded.total_routines,
		               completed_routines = excluded.completed_routines,
		               completion_rate = excluded.completion_rate,
		               flow_grade = excluded.flow_grade,
		               evaluated_at = datetime('now')`,
		date,
		totalRoutines,
		completedRoutines,
		completionRate,
		string(flowGrade),
	)
	if err != nil {
		return fmt.Errorf("failed to save daily evaluation: %w", err)
	}
	return nil
}

func (hr *HistoryRepository) GetEvaluations(limit int) ([]*DailyEvaluation, error) {
	query := `SELECT date, total_routines, completed_routines, completion_rate, flow_grade, evaluated_at
		      FROM daily_evaluation
		      ORDER BY date DESC, id DESC`
	args := []interface{}{}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	rows, err := hr.repo.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list evaluations: %w", err)
	}
	defer rows.Close()

	evaluations := make([]*DailyEvaluation, 0)
	for rows.Next() {
		entry := &DailyEvaluation{}
		var evaluatedAtRaw string
		if err := rows.Scan(
			&entry.Date,
			&entry.TotalRoutines,
			&entry.CompletedRoutines,
			&entry.CompletionRate,
			&entry.FlowGrade,
			&evaluatedAtRaw,
		); err != nil {
			return nil, fmt.Errorf("failed to scan evaluation row: %w", err)
		}
		if parsed, ok := parseDBTime(evaluatedAtRaw); ok {
			entry.EvaluatedAt = parsed
		}
		evaluations = append(evaluations, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read evaluation rows: %w", err)
	}
	return evaluations, nil
}
