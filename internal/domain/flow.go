package domain

type FlowStatus string

const (
	FlowStatusSingularity FlowStatus = "singularity"
	FlowStatusBurning     FlowStatus = "burning"
	FlowStatusSmooth      FlowStatus = "smooth"
	FlowStatusHazy        FlowStatus = "hazy"
)

func (s FlowStatus) IsValid() bool {
	switch s {
	case FlowStatusSingularity, FlowStatusBurning, FlowStatusSmooth, FlowStatusHazy:
		return true
	default:
		return false
	}
}
