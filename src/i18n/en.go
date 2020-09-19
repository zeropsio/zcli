package i18n

const (
	/// cmd
	CmdDeployDesc    = "deploy your application into zerops.io"
	CmdLogin         = "log you into zerops.io"
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

	// process
	ProcessInvalidState = "process is in wrong state"

	// cert
	CertInvalidCredentials = "invalid credentials, try `login` command"

	// zipClient
	ZipClientWorkingDirectory = "working directory"
	ZipClientMaxOneAsterix    = "only one *(asterisk) is allowed"
	ZipClientPackingDirectory = "packing directory"
	ZipClientPackingFile      = "packing file"

	// login
	LoginZeropsLoginMissing    = "param zeropsLogin must be set"
	LoginZeropsPasswordMissing = "param zeropsPassword must be set"
	LoginSuccess               = "you are logged"

	// deploy
	DeployProjectNameMissing      = "project name must be filled"
	DeployServiceStackNameMissing = "service name must be filled"
	DeployProjectNotFound         = "project not found"
	DeployProjectsWithSameName    = "there are multiple projects with same name"
	DeployServiceStatus           = "service status"
	DeployTemporaryShutdown       = "temporaryShutdown"
	DeployCreatingPackageStart    = "creating package start"
	DeployCreatingPackageDone     = "creating package done"
	DeployPackageSavedInto        = "package file saved into"
	DeployUploadingStart          = "uploading start"
	DeployUploadingDone           = "uploading done"
	DeployDeployingStart          = "deploying start"
	DeployUploadArchiveFailed          = "upload archive failed"
	DeploySuccess                 = "project deployed"

	// vpn start
	VpnStartProjectNameIsEmpty      = "project name must be filled"
	VpnStartProjectNotFound         = "project not found"
	VpnStartProjectsWithSameName    = "there are multiple projects with same name"
	VpnStartDaemonIsUnavailable     = "daemon is currently unavailable, did you install it?"
	VpnStartInstallDaemonPrompt     = "is it ok if we are going to install daemon for you?"
	VpnStartTerminatedByUser        = "when you will be ready, try `zcli daemon install`"
	VpnStartUserIsUnableToWriteYorN = "type 'y' or 'n' please"
	VpnStartSuccess                 = "vpn connection was established"

	// vpn status
	VpnStatusActive   = "vpn is active"
	VpnStatusInactive = "vpn is inactive, try `startVpn` command"

	// vpn stop
	VpnStopSuccess = "vpn connection was closed"

	// daemon
	DaemonInstallerDesc = "zerops daemon"

	// daemon install
	DaemonInstallSuccess = "zerops daemon has been installed"

	// daemon remove
	DaemonRemoveStopVpnUnavailable = "zerops daemon isn't running, vpn couldn't be removed"
	DaemonRemoveSuccess            = "zerops daemon has been removed"
)
