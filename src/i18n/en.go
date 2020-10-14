package i18n

const (
	// cmd
	CmdDeployDesc    = "deploy your application into zerops.io"
	CmdPushDesc      = "deploy your application into zerops.io and build it"
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

	// flags description
	BuildVersionName = "custom version name"
	BuildWorkingDir  = "working dir, all files path are relative to this directory"
	BuildZipFilePath = "save final zip file"

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
	BuildDeployProjectNameMissing      = "project name must be filled"
	BuildDeployServiceStackNameMissing = "service name must be filled"
	BuildDeployProjectNotFound         = "project not found"
	BuildDeployProjectsWithSameName    = "there are multiple projects with same name"
	BuildDeployServiceStatus           = "service status"
	BuildDeployTemporaryShutdown       = "temporaryShutdown"
	BuildDeployCreatingPackageStart    = "creating package"
	BuildDeployCreatingPackageDone     = "package created"
	BuildDeployPackageSavedInto        = "package file saved into"
	BuildDeployUploadingPackageStart   = "uploading package"
	BuildDeployUploadingPackageDone    = "package uploaded"
	BuildDeployUploadPackageFailed     = "package upload failed"
	BuildDeployDeployingStart          = "deploying service"
	BuildDeployBuildConfigNotFound     = "config file zerops_build.yml is not found"
	BuildDeployBuildConfigEmpty        = "config file zerops_build.yml is empty"
	BuildDeployBuildConfigTooLarge     = "max. size of zerops_build.yml is 10 MB"
	BuildDeploySuccess                 = "service deployed"

	// vpn start
	VpnStartProjectNameIsEmpty         = "project name must be filled"
	VpnStartProjectNotFound            = "project not found"
	VpnStartInterfaceAssignFailed      = "interface name assign failed"
	VpnStartWireguardInterfaceNotfound = "wireguard interface not found"
	VpnStartProjectsWithSameName       = "there are multiple projects with same name"
	VpnStartDaemonIsUnavailable        = "daemon is currently unavailable, did you install it?"
	VpnStartInstallDaemonPrompt        = "is it ok if we are going to install daemon for you?"
	VpnStartTerminatedByUser           = "when you will be ready, try `zcli daemon install`"
	VpnStartUserIsUnableToWriteYorN    = "type 'y' or 'n' please"
	VpnStartSuccess                    = "vpn connection was established"

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
