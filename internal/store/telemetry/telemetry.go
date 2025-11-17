package telemtry

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TelemetryRepo struct {
	db *sqlx.DB
}

type trackDTO struct {
	ID     uuid.UUID `db:"id"`
	Name   string    `db:"name"`
	Layout string    `db:"layout"`
}

type carDTO struct {
	ID      uuid.UUID `db:"id"`
	Name    string    `db:"name"`
	ClassID uuid.UUID `db:"class_id"`
}

type classDTO struct {
	ID    uuid.UUID   `db:"id"`
	Title string      `db:"title"`
	Cars  []uuid.UUID `db:"cars"`
}

type eventDTO struct {
	ID           uuid.UUID   `db:"id"`
	Title        string      `db:"title"`
	Classes      []uuid.UUID `db:"classes"`
	State        string      `db:"state"`
	Picture      string      `db:"picture`
	TrackID      uuid.UUID   `db:"track_id"`
	Laps         uint        `db:"laps"`
	StartsAt     time.Time   `db:"starts_at"`
	EndsAt       time.Time   `db:"ends_at"`
	IsHighReward bool        `db:"is_high_reward"`
}
