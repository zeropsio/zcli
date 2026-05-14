package kongCmd

// ProjectScope contributes --project-id / -P. Pure flag bundle; resolution
// lives in (*Auth).RequireProject.
type ProjectScope struct {
	ProjectID string `name:"project-id" short:"P" help:"Project ID."`
}

// ServiceScope contributes --service-id / -S and a service-id-or-name
// positional. Embeds ProjectScope so a single embed yields both flag sets.
type ServiceScope struct {
	ServiceID  string `name:"service-id" short:"S" help:"Service ID."`
	ServiceArg string `arg:"" name:"service-id-or-name" optional:"" help:"Service ID or name."`
	ProjectScope
}

// Per-call options passed to Auth.Require* — compile-time only, not flags.

type projectReqCfg struct {
	createNew  bool
	skipSelect bool
}

type ProjectOpt func(*projectReqCfg)

func WithCreateNewProject() ProjectOpt  { return func(c *projectReqCfg) { c.createNew = true } }
func WithSkipSelectProject() ProjectOpt { return func(c *projectReqCfg) { c.skipSelect = true } }

type serviceReqCfg struct {
	createNew bool
}

type ServiceOpt func(*serviceReqCfg)

func WithCreateNewService() ServiceOpt { return func(c *serviceReqCfg) { c.createNew = true } }
