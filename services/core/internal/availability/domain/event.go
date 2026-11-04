package domain

import "time"

type Event struct {
	id         string
	typeName   string
	spotID     string
	coverageID string
	releaserID string
	cellR10    int64
	confidence float64
	occurredAt time.Time
}

func NewSpotAnnouncedEvent(id string, spot Spot, occurredAt time.Time) Event {
	return Event{
		id:         id,
		typeName:   "availability.spot.announced.v1",
		spotID:     spot.ID(),
		coverageID: spot.CoverageID(),
		releaserID: spot.ReleaserID(),
		cellR10:    spot.CellR10(),
		confidence: spot.Confidence(),
		occurredAt: occurredAt,
	}
}

func (e Event) ID() string            { return e.id }
func (e Event) Type() string          { return e.typeName }
func (e Event) SpotID() string        { return e.spotID }
func (e Event) CoverageID() string    { return e.coverageID }
func (e Event) ReleaserID() string    { return e.releaserID }
func (e Event) CellR10() int64        { return e.cellR10 }
func (e Event) Confidence() float64   { return e.confidence }
func (e Event) OccurredAt() time.Time { return e.occurredAt }
