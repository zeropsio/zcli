package entity

// Domain enums. Each mirrors a zerops-go enum but lives in the domain layer so
// consumers compare and display statuses without importing zerops-go. The
// repository layer maps to and from the zerops-go enums at the API boundary.
// Only the values and predicates the CLI actually uses are reproduced here.

type ProjectMode string

const (
	ProjectModeLight   ProjectMode = "LIGHT"
	ProjectModeSerious ProjectMode = "SERIOUS"
	ProjectModeLegacy  ProjectMode = "LEGACY"
)

func (m ProjectMode) String() string { return string(m) }

// AllPublicProjectModes are the modes a user may pass to `project create`.
func AllPublicProjectModes() []ProjectMode {
	return []ProjectMode{ProjectModeLight, ProjectModeSerious}
}

type ProjectStatus string

func (s ProjectStatus) String() string { return string(s) }

type OrgStatus string

const OrgStatusActive OrgStatus = "ACTIVE"

func (s OrgStatus) String() string { return string(s) }
func (s OrgStatus) IsActive() bool { return s == OrgStatusActive }

type OrgRole string

func (r OrgRole) String() string { return string(r) }

type ServiceStatus string

func (s ServiceStatus) String() string { return string(s) }

type ServiceTypeCategory string

func (c ServiceTypeCategory) String() string { return string(c) }

type ContainerStatus string

func (s ContainerStatus) String() string { return string(s) }

type AppVersionStatus string

const AppVersionStatusUploading AppVersionStatus = "UPLOADING"

func (s AppVersionStatus) String() string    { return string(s) }
func (s AppVersionStatus) IsUploading() bool { return s == AppVersionStatusUploading }

type AppVersionSource string

func (s AppVersionSource) String() string { return string(s) }

type ProcessStatus string

func (s ProcessStatus) String() string { return string(s) }
