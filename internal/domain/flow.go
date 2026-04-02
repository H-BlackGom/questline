package domain

type FlowStatus string

const (
	FlowStatusBurning FlowStatus = "burning"
	FlowStatusSmooth  FlowStatus = "smooth"
	FlowStatusHazy    FlowStatus = "hazy"
)

func (s FlowStatus) IsValid() bool {
	switch s {
	case FlowStatusBurning, FlowStatusSmooth, FlowStatusHazy:
		return true
	default:
		return false
	}
}
