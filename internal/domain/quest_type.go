package domain

type QuestType string

const (
	QuestTypeDaily  QuestType = "daily"
	QuestTypeWeekly QuestType = "weekly"
	QuestTypeEpic   QuestType = "epic"
	QuestTypeGuild  QuestType = "guild"
	QuestTypeSub    QuestType = "sub"
)

func (t QuestType) IsValid() bool {
	switch t {
	case QuestTypeDaily, QuestTypeWeekly, QuestTypeEpic, QuestTypeGuild, QuestTypeSub:
		return true
	default:
		return false
	}
}
