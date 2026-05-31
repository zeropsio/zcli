package entity

// ImportServiceStack is one service produced by a project/service import, along
// with the processes the import queued for it.
type ImportServiceStack struct {
	Id        ServiceId
	Name      string
	Processes []Process
}
