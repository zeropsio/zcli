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
	VpnStartTunnelStatusActive         = "wireguard tunnel is active"
	VpnStartTunnelStatusInactive       = "wireguard tunnel is inactive, try `startVpn` command"
	VpnStartDnsStatusActive            = "dns is active"
	VpnStartDnsStatusInactive          = "dns is inactive, we weren't able to set dns"
	VpnStartAdditionalInfo             = "additional info:"

	// vpn status
	VpnStatusTunnelStatusActive   = "wireguard tunnel is active"
	VpnStatusTunnelStatusInactive = "wireguard tunnel is inactive, try `startVpn` command"
	VpnStatusDnsStatusActive      = "dns is active"
	VpnStatusDnsStatusInactive    = "dns is inactive, we weren't able to set dns"
	VpnStatusAdditionalInfo       = "additional info:"

	// vpn stop
	VpnStopSuccess               = "vpn connection was closed"
	VpnStopAdditionalInfo        = "additional info:"
	VpnStopAdditionalInfoMessage = "dns could be set by yourself, if so it must be removed manually"

	// daemon
	DaemonInstallerDesc = "zerops daemon"

	// daemon install
	DaemonInstallSuccess = "zerops daemon has been installed"

	// daemon remove
	DaemonRemoveStopVpnUnavailable = "zerops daemon isn't running, vpn couldn't be removed"
	DaemonRemoveSuccess            = "zerops daemon has been removed"
)
