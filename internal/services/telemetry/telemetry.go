package telemetry

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Track struct {
	ID     uuid.UUID
	Name   string
	Layout string
}

type TrackParams struct {
	Name   string
	Layout string
}

type Car struct {
	ID    uuid.UUID
	Name  string
	Class *Class
}

type CarParams struct {
	Name    string
	ClassID uuid.UUID
}

type Class struct {
	ID    uuid.UUID
	Title string
	Cars  []*Car
}

type ClassParams struct {
	Title string
	Cars  []uuid.UUID
}

type Event struct {
	ID           uuid.UUID
	Title        string
	Classes      []*Class
	State        string // ongoing, finished, etc.
	Picture      string
	Track        *Track
	Laps         uint
	StartsAt     time.Time
	EndsAt       time.Time
	IsHighReward bool
}

type EventParams struct {
	Title        string
	Classes      []uuid.UUID
	State        string
	Picture      string
	TrackID      uuid.UUID
	Laps         uint
	StartsAt     time.Time
	EndsAt       time.Time
	IsHighReward bool
}

type TelemetryRepo interface {
	// TODO: CRUD for every type
	GetTrack(context.Context, uuid.UUID) (*Track, error)
	CreateTrack(context.Context, *TrackParams) (*Track, error)
	UpdateTrack(context.Context, uuid.UUID, *TrackParams) (*Track, error)
	DeleteTrack(context.Context, uuid.UUID) error

	GetCar(context.Context, uuid.UUID) (*Car, error)
	CreateCar(context.Context, *CarParams) (*Car, error)
	UpdateCar(context.Context, uuid.UUID, *CarParams) (*Car, error)
	DeleteCar(context.Context, uuid.UUID) error

	GetClass(context.Context, uuid.UUID) (*Class, error)
	CreateClass(context.Context, *ClassParams) (*Class, error)
	UpdateClass(context.Context, uuid.UUID, *ClassParams) (*Class, error)
	DeleteClass(context.Context, uuid.UUID) error

	GetEvent(context.Context, uuid.UUID) (*Event, error)
	CreateEvent(context.Context, *EventParams) (*Event, error)
	UpdateEvent(context.Context, uuid.UUID, *EventParams) (*Event, error)
	DeleteEvent(context.Context, uuid.UUID) error
}
