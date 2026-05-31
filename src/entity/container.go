package entity

import "time"

type Container struct {
	Id        ContainerId
	OrgId     OrgId
	ProjectId ProjectId
	ServiceId ServiceId
	Status    *ContainerStatus
	Number    int
	Name      string
	Hostname  string
	Created   time.Time
}
