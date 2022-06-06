package i18n

import "fmt"

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
	CmdRegion        = "region commands group"
	CmdRegionList    = "list available regions"
	CmdProject       = "project commands group"
	CmdService       = "service commands group"
	CmdProjectStart  = "run process to start the project and wait until finished"
	CmdProjectStop   = "run process to stop the project and wait until finished"
	CmdProjectDelete = "run process to delete the project and wait until finished"
	CmdProjectImport = "create project in Zerops and add service(s)"
	CmdServiceImport = "create one or more services for given project"
	CmdServiceStart  = "run process to start the service and wait until finished"
	CmdServiceStop   = "run process to stop the service and wait until finished"
	CmdServiceDelete = "run process to delete the service and wait until finished"

	// flags description
	BuildVersionName     = "custom version name"
	SourceName           = "zerops.yml source service"
	BuildWorkingDir      = "working dir, all files path are relative to this directory"
	BuildArchiveFilePath = "path (including file name) where the final tar.gz archive file should be saved (if not set, archive won't be saved)"
	ZeropsYamlLocation   = "zerops yaml location relative to working directory"
	DeployGitFolder      = "whether `.git` folder should also be deployed during `zcli push` command"
	ClientId             = "client ID"
	ConfirmDeleteProject = "confirm to delete the project"
	ConfirmDeleteService = "confirm to delete the service"

	// process
	ProcessInvalidState        = "last command has finished with error, identifier for communication with our support: %s"
	ProcessInvalidStateProcess = "process finished with error, identifier for communication with our support:"
	QueuedProcesses            = "queued processes: "
	ProcessStart               = "process started"
	ProcessEnd                 = "process finished"

	// archiveClient
	ArchClientWorkingDirectory  = "working directory: %s"
	ArchClientMaxOneTilde       = "only one ~(tilde) is allowed"
	ArchClientPackingDirectory  = "packing directory: %s"
	ArchClientPackingFile       = "packing file: %s"
	ArchClientFileAlreadyExists = "file [%s] already exists"

	// login
	LoginParamsMissing = "either login with password or token must be passed"
	LoginSuccess       = "you are logged in"
	LoginVpnClosed     = "vpn connection was closed"

	// region
	RegionNotFound = "region not found"

	// client ID
	MultipleClientIds  = "you have assigned multiple client IDs, please use the --clientId flag"
	AvailableClientIds = "your client IDs are: "
	MissingClientId    = "no client ID found four your account"

	// import
	YamlCheck             = "yaml file check started"
	ImportYamlOk          = "yaml file ok"
	ImportYamlEmpty       = "config file import yaml is empty"
	ImportYamlTooLarge    = "max. size of import yaml is 100 KB"
	ImportYamlFound       = "import yaml found"
	ImportYamlNotFound    = "import yaml not found"
	ImportYamlCorrupted   = "import yaml corrupted"
	ServiceStackCount     = "number of services to be added: "
	CoreServices          = "core services activation started"
	ReadyToImportServices = "ready to import services"

	// delete cmd
	DeleteCanceledByUser = "delete command canceled by user"

	// project + service
	ProjectNotFound      = "project not found"
	ProjectIdInvalid     = "project ID invalid"
	ProjectWrongId       = "Please, provide correct project ID."
	ProjectsWithSameName = "found multiple projects with the same name"
	AvailableProjectIds  = "available project IDs are: "
	ProjectNameOrIdEmpty = "project name or ID must be filled"
	ProjectDeleteConfirm = "Please confirm that you would like to delete the project (y/n): "
	ServiceNotFound      = "service not found"
	ServiceNameIsEmpty   = "service name must be filled"
	ServiceDeleteConfirm = "Please confirm that you would like to delete the service (y/n): "
	ProcessInit          = " command initialized"
	Success              = " successfully"
	ProjectStart         = "project start"
	ProjectStop          = "project stop"
	ProjectDelete        = "project delete"
	ProjectStarted       = "project started"
	ProjectStopped       = "project stopped"
	ProjectDeleted       = "project deleted"
	ProjectCreated       = "project created"
	ServiceStart         = "service start"
	ServiceStop          = "service stop"
	ServiceDelete        = "service delete"
	ServiceStarted       = "service started"
	ServiceStopped       = "service stopped"
	ServiceDeleted       = "service deleted"
	ProjectImported      = "project imported"
	ServiceImported      = "service(s) imported"

	// deploy
	BuildDeployServiceStatus         = "service status: %s"
	BuildDeployCreatingPackageStart  = "creating package"
	BuildDeployCreatingPackageDone   = "package created"
	BuildDeployPackageSavedInto      = "package file saved into: %s"
	BuildDeployUploadingPackageStart = "uploading package"
	BuildDeployUploadingPackageDone  = "package uploaded"
	BuildDeployUploadPackageFailed   = "package upload failed"
	BuildDeployDeployingStart        = "deploying service"
	BuildDeployZeropsYamlEmpty       = "config file zerops.yml is empty"
	BuildDeployZeropsYamlTooLarge    = "max. size of zerops.yml is 10 KB"
	BuildDeployZeropsYamlFound       = "zerops.yml found"
	BuildDeployZeropsYamlNotFound    = "zerops.yml not found"
	BuildDeploySuccess               = "service deployed"

	// vpn start
	VpnStartInterfaceAssignFailed      = "interface name assign failed"
	VpnStartWireguardInterfaceNotfound = "wireguard interface not found"
	VpnStartDaemonIsUnavailable        = "daemon is currently unavailable, did you install it?"
	VpnStartInstallDaemonPrompt        = "is it ok to install zerops daemon for you?"
	VpnStartTerminatedByUser           = "when you are ready, try `/path/to/zcli daemon install`"
	VpnStartUserIsUnableToWriteYorN    = "type 'y' or 'n' please"
	VpnStartWireguardUtunError         = "we failed to start vpn, there is possibility that you have another vpn, if so, try to shut it down"
	VpnStartVpnNotReachable            = "zerops vpn servers aren't reachable"
	VpnStartTunnelIsNotAlive           = "we failed to establish zerops vpn"
	VpnStartExpectedProjectName        = "expected project name or ID as a positional argument"

	// vpn status
	VpnStatusDaemonIsUnavailable     = "daemon is currently unavailable, did you install it?"
	VpnStatusTunnelStatusActive      = "wireguard tunnel is working properly"
	VpnStatusTunnelStatusSetInactive = "wireguard tunnel is established but it isn't working properly, try `/path/to/zcli vpn start` command"
	VpnStatusTunnelStatusUnset       = "wireguard tunnel isn't established, try `/path/to/zcli vpn start` command"
	VpnStatusDnsStatusActive         = "dns is working properly"
	VpnStatusDnsStatusSetInactive    = "dns is set but it isn't working properly, try `/path/to/zcli vpn start` command"
	VpnStatusDnsStatusUnset          = "dns isn't set, try `/path/to/zcli vpn start` command"
	VpnStatusAdditionalInfo          = "additional info:"
	VpnStatusDnsCheckError           = "we failed to check that dns is working correctly"
	VpnStatusDnsNoCheckFunction      = "there is no function for dns check"

	// vpn stop
	VpnStopDaemonIsUnavailable   = "daemon is currently unavailable, did you install it?"
	VpnStopSuccess               = "vpn connection was closed"
	VpnStopAdditionalInfo        = "additional info:"
	VpnStopAdditionalInfoMessage = "dns could be set by yourself, if so it must be removed manually"

	// daemon
	DaemonInstallerDesc = "zerops daemon"
	DaemonElevated      = "operation continues in a new window"
	PathNotFound        = "path not found"

	// daemon install
	DaemonInstallSuccess                 = "zerops daemon has been installed"
	DaemonInstallWireguardNotFound       = "wireguard was not found"
	DaemonInstallWireguardNotFoundDarwin = "wireguard was not found, try `brew install wireguard-tools`"

	// daemon remove
	DaemonRemoveStopVpnUnavailable = "zerops daemon isn't running, vpn couldn't be removed"
	DaemonRemoveSuccess            = "zerops daemon has been removed"

	// generic
	GrpcApiTimeout    = "zerops api didn't respond within assigned time, try it again later"
	GrpcVpnApiTimeout = "zerops vpn server didn't respond within assigned time, try it again later"

	HintChangeRegion = "hint: try to change your region (you can list available regions using `zcli region list`)"
)

func AddHintChangeRegion(err error) error {
	return fmt.Errorf("%w\n%s", err, HintChangeRegion)
}
