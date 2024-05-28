package scope

import (
	"github.com/zeropsio/zcli/src/cmdBuilder"
	"github.com/zeropsio/zcli/src/generic"
	"github.com/zeropsio/zerops-go/types/enum"
)

var projectScope *project

func Project() cmdBuilder.ScopeLevel {
	return projectScope
}

var serviceScope *service

type ServiceScopeOption generic.Option[service]

func WithServiceCategoryRestriction(restrictions ...enum.ServiceStackTypeCategoryEnum) ServiceScopeOption {
	return func(s *service) {
		s.serviceCategoryRestrictions = append(s.serviceCategoryRestrictions, restrictions...)
	}
}

func Service(opts ...ServiceScopeOption) cmdBuilder.ScopeLevel {
	s := generic.ApplyOptionsWithDefault(*serviceScope, opts...)
	return &s
}

func init() {
	projectScope = &project{}
	serviceScope = &service{
		parent: projectScope,
	}
}
