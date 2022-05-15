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
	CopyIgnore      []string
	RemoveContainer bool
	Previous        *Job
	ErrorChannel    chan error
}

type Port struct {
	Out string
	In  string
}
