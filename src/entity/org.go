package entity

type Org struct {
	Id     OrgId
	Role   OrgRole
	Name   string
	Status OrgStatus
}

var OrgFields = entityTemplateFields[Org]()
