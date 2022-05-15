package pin

type Job struct {
	Name            string
	Image           string
	Script          []string
	WorkDir         string
	CopyFiles       bool
	Status          int
	SoloExecution   bool
	Port            []Port
	RemoveContainer bool
	Next            *Job
}

type Port struct {
	Out string
	In  string
}
