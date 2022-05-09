package i18n

const (
	// cmd
	CmdDeployDesc    = "deploy your application into Zerops"
	CmdPushDesc      = "deploy your application into Zerops and build it"
	CmdLogin         = "log you into Zerops"
	CmdVpn           = "vpn commands group"
	CmdVpnStart      = "start vpn"
	CmdVpnStop       = "stop vpn"
	CmdVpnStatus     = "show vpn status"
	CmdLog           = "logs commands"
	CmdLogShow       = "show logs"
	CmdDaemon        = "daemon commands group"
	CmdDaemonRun     = "run daemon"
	CmdDaemonInstall = "install daemon"
	CmdDaemonRemove  = "remove daemon"
	CmdVersion       = "version"
	CmdProjectStart  = "run process to start the project and wait until finished"
	CmdProjectStop   = "run process to stop the project and wait until finished"
	CmdProjectDelete = "run process to delete the project and wait until finished"
	CmdProjectImport = "create project in zerops.io and add service(s)"
	CmdServiceImport = "create one or more services for given project"
	CmdServiceStart  = "run process to start the service and wait until finished"
	CmdServiceStop   = "run process to stop the service and wait until finished"
	CmdServiceDelete = "run process to delete the service and wait until finished"

	// flags description
	BuildVersionName     = "custom version name"
	SourceName           = "zerops.yml source service"
	BuildWorkingDir      = "working dir, all files path are relative to this directory"
	BuildZipFilePath     = "save final zip file"
	ZeropsYamlLocation   = "zerops yaml location relative to working directory"
	ImportYamlLocation   = "import yaml location relative to working directory"
	ClientId             = "client ID"
	ConfirmDeleteProject = "confirm delete project"

	// process
	ProcessInvalidState = "last command has finished with error, identifier for communication with our support: %s"

	// zipClient
	ZipClientWorkingDirectory = "working directory: %s"
	ZipClientMaxOneTilde      = "only one ~(tilde) is allowed"
	ZipClientPackingDirectory = "packing directory: %s"
	ZipClientPackingFile      = "packing file: %s"

	// login
	LoginParamsMissing = "either login with password or token must be passed"
	LoginSuccess       = "you are logged in"
	LoginVpnClosed     = "vpn connection was closed"

	// region
	RegionNotFound = "region not found"

	// import project
	ImportYamlEmpty     = "config file import.yml is empty"
	ImportYamlTooLarge  = "max. size of import.yml is 10 KB"
	ImportYamlFound     = "import.yml found"
	ImportYamlNotFound  = "import.yml not found"
	ImportYamlCorrupted = "import yaml corrupted"

	// import service
	ImportServiceFailed = "import service failed"

	// find project by name
	ProjectNotFound      = "project not found"
	ProjectsWithSameName = "there are multiple projects with the same name"
	ProjectNameIsEmpty   = "project name must be filled"

	// start project
	StartProjectProcessInit = "starting the project"
	StartProcessSuccess     = "project started successfully"

	// stop project
	StopProjectProcessInit = "stopping the project"
	StopProcessSuccess     = "project stopped successfully"

	// delete project
	ConfirmDelete            = "Please confirm you want to delete the project (y/n): "
	CanceledByUser           = "delete project command canceled by user"
	DeleteProjectProcessInit = "going to delete the project"
	DeleteProcessSuccess     = "project deleted successfully"

	// deploy
	BuildDeployProjectNameMissing      = "project name must be filled"
	BuildDeployServiceStackNameMissing = "service name must be filled"
	BuildDeployProjectNotFound         = "project not found"
	BuildDeployProjectsWithSameName    = "there are multiple projects with same name"
	BuildDeployServiceStatus           = "service status: %s"
	BuildDeployCreatingPackageStart    = "creating package"
	BuildDeployCreatingPackageDone     = "package created"
	BuildDeployPackageSavedInto        = "package file saved into: %s"
	BuildDeployUploadingPackageStart   = "uploading package"
	BuildDeployUploadingPackageDone    = "package uploaded"
	BuildDeployUploadPackageFailed     = "package upload failed"
	BuildDeployDeployingStart          = "deploying service"
	BuildDeployZeropsYamlEmpty         = "config file zerops.yml is empty"
	BuildDeployZeropsYamlTooLarge      = "max. size of zerops.yml is 10 KB"
	BuildDeployZeropsYamlFound         = "zerops.yml found"
	BuildDeployZeropsYamlNotFound      = "zerops.yml not found"
	BuildDeploySuccess                 = "service deployed"

	// vpn start
	VpnStartProjectNameIsEmpty         = "project name must be filled"
	VpnStartProjectNotFound            = "project not found"
	VpnStartInterfaceAssignFailed      = "interface name assign failed"
	VpnStartWireguardInterfaceNotfound = "wireguard interface not found"
	VpnStartProjectsWithSameName       = "there are multiple projects with same name"
	VpnStartDaemonIsUnavailable        = "daemon is currently unavailable, did you install it?"
	VpnStartInstallDaemonPrompt        = "is it ok if we are going to install daemon for you?"
	VpnStartTerminatedByUser           = "when you will be ready, try `/path/to/zcli daemon install`"
	VpnStartUserIsUnableToWriteYorN    = "type 'y' or 'n' please"
	VpnStartWireguardUtunError         = "we weren't able to start vpn, there is possibility that you have another vpn, if so, try to shut it down"
	VpnStartVpnNotReachable            = "zerops vpn servers aren't reachable"
	VpnStartTunnelIsNotAlive           = "we weren't able to establish zerops vpn"
	VpnStartExpectedProjectName        = "expected project name as a positional argument"

	// vpn status
	VpnStatusDaemonIsUnavailable     = "daemon is currently unavailable, did you install it?"
	VpnStatusTunnelStatusActive      = "wireguard tunnel is working properly"
	VpnStatusTunnelStatusSetInactive = "wireguard tunnel is established but it isn't working properly, try `/path/to/zcli vpn start` command"
	VpnStatusTunnelStatusUnset       = "wireguard tunnel isn't established, try `/path/to/zcli vpn start` command"
	VpnStatusDnsStatusActive         = "dns is working properly"
	VpnStatusDnsStatusSetInactive    = "dns is set but it isn't working properly, try `/path/to/zcli vpn start` command"
	VpnStatusDnsStatusUnset          = "dns isn't set, try `/path/to/zcli vpn start` command"
	VpnStatusAdditionalInfo          = "additional info:"
	VpnStatusDnsCheckError           = "we weren't able to check that dns working correctly"
	VpnStatusDnsNoCheckFunction      = "there is no function for dns check"

	// vpn stop
	VpnStopDaemonIsUnavailable   = "daemon is currently unavailable, did you install it?"
	VpnStopSuccess               = "vpn connection was closed"
	VpnStopAdditionalInfo        = "additional info:"
	VpnStopAdditionalInfoMessage = "dns could be set by yourself, if so it must be removed manually"

	// daemon
	DaemonInstallerDesc = "zerops daemon"
	DaemonElevated      = "operation continues in the new window"

	// daemon install
	DaemonInstallSuccess                 = "zerops daemon has been installed"
	DaemonInstallWireguardNotFound       = "wireguard was not found"
	DaemonInstallWireguardNotFoundDarwin = "wireguard was not found, try `brew install wireguard-tools`"

	// daemon remove
	DaemonRemoveStopVpnUnavailable = "zerops daemon isn't running, vpn couldn't be removed"
	DaemonRemoveSuccess            = "zerops daemon has been removed"

	// generic
	GrpcApiTimeout    = "zerops api didn't response within assigned time, try it again later"
	GrpcVpnApiTimeout = "zerops vpn server didn't response within assigned time, try it again later"
)
