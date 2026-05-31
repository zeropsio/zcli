package entity

// Domain identifier types. They keep the compile-time safety zerops-go's uuid
// types gave us (a ServiceId can't be passed where a ProjectId is expected)
// without leaking zerops-go into the rest of the codebase. The repository layer
// converts to and from the zerops-go id types at the API boundary.
//
// Each is a string underneath, so callers can render it with %s or convert with
// string(id), and the String method keeps text/template `--out` output identical
// to what the old zerops-go id types produced.

type (
	ProjectId     string
	OrgId         string
	ServiceId     string
	ContainerId   string
	AppVersionId  string
	ProcessId     string
	LocationId    string
	ServiceTypeId string
)

func (id ProjectId) String() string     { return string(id) }
func (id OrgId) String() string         { return string(id) }
func (id ServiceId) String() string     { return string(id) }
func (id ContainerId) String() string   { return string(id) }
func (id AppVersionId) String() string  { return string(id) }
func (id ProcessId) String() string     { return string(id) }
func (id LocationId) String() string    { return string(id) }
func (id ServiceTypeId) String() string { return string(id) }
