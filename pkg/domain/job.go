package domain

type Job struct {
	ExitCode int
	Finished int64
	Started  int64
	Status   string
}
