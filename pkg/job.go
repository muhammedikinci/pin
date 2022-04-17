package pin

type Job struct {
	Name      string
	Image     string
	Script    []string
	WorkDir   string
	CopyFiles bool
	Status    int
	Next      *Job
}
