package domain

import "time"

type Build struct {
	Created  int64
	Event    string
	Finished int64
	Link     string
	Number   int
	Started  int64
	Status   string
}

func (b *Build) CreatedToTimeFormat(format string) string {
	return time.Unix(b.Created, 0).Format(format)
}

func (b *Build) FinishedToTimeFormat(format string) string {
	return time.Unix(b.Finished, 0).Format(format)
}

func (b *Build) IsEvent(expectedEvent string) bool {
	return expectedEvent == b.Event
}

func (b *Build) IsStatus(expectedStatus string) bool {
	return expectedStatus == b.Status
}

func (b *Build) StartedToTimeFormat(format string) string {
	return time.Unix(b.Started, 0).Format(format)
}
