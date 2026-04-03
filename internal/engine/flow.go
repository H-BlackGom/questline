package engine

import "github.com/H-BlackGom/questline/internal/domain"

func CalculateFlowGrade(completionRate float64) domain.FlowStatus {
	switch {
	case completionRate == 1.0:
		return domain.FlowStatusSingularity
	case completionRate >= 0.8:
		return domain.FlowStatusBurning
	case completionRate >= 0.5:
		return domain.FlowStatusSmooth
	default:
		return domain.FlowStatusHazy
	}
}

func EvaluateFlowStatus(totalDaily, completedDaily int) domain.FlowStatus {
	if totalDaily <= 0 {
		return domain.FlowStatusSmooth
	}
	if completedDaily < 0 {
		completedDaily = 0
	}
	if completedDaily > totalDaily {
		completedDaily = totalDaily
	}
	rate := float64(completedDaily) / float64(totalDaily)
	return CalculateFlowGrade(rate)
}
