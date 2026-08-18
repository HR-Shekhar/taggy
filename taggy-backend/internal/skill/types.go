package skill

import "time"

// MilestoneAction is the service-layer input for updating milestone progress.
type MilestoneAction string

const (
	MilestoneActionComplete MilestoneAction = "COMPLETE"
	MilestoneActionPostpone MilestoneAction = "POSTPONE"
)

type UpdateMilestoneInput struct {
	Action         MilestoneAction
	PostponedUntil *time.Time
}
