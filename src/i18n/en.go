package i18n

import "fmt"

const (
	// help
	DisplayHelp       = "Displays help for "
	GroupHelp         = "any command."
	DaemonInstallHelp = "the daemon install command."
	DaemonRemoveHelp  = "the daemon remove command."
	DaemonRunHelp     = "the daemon run command."
	DeployHelp        = "the deploy command."
	LogShowHelp       = "the log show command."
	LoginHelp         = "the login command."
	ProjectStartHelp  = "the project start command."
	ProjectStopHelp   = "the project stop command."
	ProjectDeleteHelp = "the project delete command."
	ProjectImportHelp = "the project import command."
	PushHelp          = "the push command."
	RegionListHelp    = "the region list command."
	ServiceStartHelp  = "the service start command."
	ServiceStopHelp   = "the service stop command."
	ServiceImportHelp = "the service import command."
	ServiceDeleteHelp = "the service delete command."
	ServiceLogHelp    = "the service log command."
	VersionHelp       = "the version command."
	VpnStartHelp      = "the vpn start command."
	VpnStopHelp       = "the vpn stop command."
	VpnStatusHelp     = "the vpn status command."
	BucketCreateHelp  = "the bucket create command."
	BucketDeleteHelp  = "the bucket delete command."

	// cmd short
	CmdDeployDesc    = "Deploys your application to Zerops."
	CmdPushDesc      = "Builds your application in Zerops and deploys it."
	CmdLogin         = "Logs you into Zerops. Use a generated Zerops token or your login e-mail and password."
	CmdVpn           = "VPN commands group."
	CmdVpnStart      = "Starts a VPN session."
	CmdVpnStop       = "Stops the existing VPN session."
	CmdVpnStatus     = "Shows the status of the VPN session."
	CmdLog           = "zCLI log commands group."
	CmdLogShow       = "Shows zCLI logs."
	CmdDaemon        = "Zerops VPN daemon commands group."
	CmdDaemonRun     = "Runs Zerops VPN daemon."
	CmdDaemonInstall = "Installs Zerops VPN daemon."
	CmdDaemonRemove  = "Removes Zerops VPN daemon."
	CmdVersion       = "Shows the current zCLI version."
	CmdRegion        = "Zerops region commands group."
	CmdRegionList    = "Lists all Zerops regions."
	CmdProject       = "Project commands group."
	CmdService       = "Zerops service commands group."
	CmdProjectStart  = "Starts the project and the services that were running before the project was stopped."
	CmdProjectStop   = "Stops the project and all of its services."
	CmdProjectDelete = "Deletes the project and all of its services."
	CmdProjectImport = "Creates a new project with one or more services."
	CmdServiceImport = "Creates one or more Zerops services in an existing project."
	CmdServiceStart  = "Starts the Zerops service."
	CmdServiceStop   = "Stops the Zerops service."
	CmdServiceDelete = "Deletes the Zerops service."
	CmdServiceLog    = "Get service runtime or build log to stdout."
	CmdBucket        = "S3 storage management"
	CmdBucketZerops  = "Management via Zerops API"
	CmdBucketS3      = "Management directly via S3 API"
	CmdBucketCreate  = "Creates a bucket in an existing object storage."
	CmdBucketDelete  = "Deletes a bucket from an existing object storage."

	// cmd long
	ProjectImportLong = "Creates a new project with one or more services according to the definition in the import YAML file."
	DeployDescLong    = "pathToFileOrDir defines a path to one or more directories and/or files relative to the working\ndirectory. The working directory is by default the current directory and can be changed\nusing the --workingDir flag. zCLI deploys selected directories and/or files to Zerops."
	PushDescLong      = "The command triggers the build pipeline defined in zerops.yml. Zerops.yml must be in the working\ndirectory. The working directory is by default the current directory and can be changed\nusing the --workingDir flag. zCLI uploads all files and subdirectories of the working\ndirectory to Zerops and starts the build pipeline. Files found in the .gitignore\nfile will be ignored.\n\nIf you just want to deploy your application to Zerops, use the zcli deploy command instead."
	// CmdServiceLogFull = "Returns service runtime or build log to stdout with a streaming option. By default, the command returns the last 100 log messages from all service runtime containers and exits. Use --follow flag to continuously pool for new log messages.\n"
	CmdServiceLogLong    = "Returns service runtime or build log to stdout. By default, the command returns the last 100\nlog messages from all service runtime containers and exits.\n"
	ServiceLogAdditional = "\nUse the <serviceName> alone in the command to return log messages from all runtime containers.\nSet <serviceName>@1 to return log messages from the first runtime container only.\nSet <serviceName>@build to return log messages from the last build if available."
	VpnStartLong         = "Starts a VPN session in the selected Zerops project. You can't be connected to multiple projects\nat the same time. If the previous VPN session is active, it will be stopped automatically\nand a new VPN session will start.\n"

	// flags description
	RegionFlag            = "Choose one of Zerops regions. Use the \"zcli region list\" command to list all Zerops regions."
	ZeropsLoginFlag       = "Your login e-mail. Automatically filled if the ZEROPSLOGIN environment variable exists."
	ZeropsPwdFlag         = "Your password. Automatically filled if the ZEROPSPASSWORD environment variable exists."
	ZeropsTokenFlag       = "Zerops token. Automatically filled if the ZEROPSTOKEN environment variable exists."
	RegionUrlFlag         = "Zerops region file url."
	BuildVersionName      = "Adds a custom version name. Automatically filled if the VERSIONNAME environment variable exists."
	SourceName            = "Override zerops.yml service name."
	BuildWorkingDir       = "Sets a custom working directory. Default working directory is the current directory."
	BuildArchiveFilePath  = "If set, zCLI creates a tar.gz archive with the application code in the required path relative\nto the working directory. By default, no archive is created."
	ZeropsYamlLocation    = "Sets a custom path to the zerops.yml file relative to the working directory. By default zCLI\nlooks for zerops.yml in the working directory."
	UploadGitFolder       = "If set, zCLI the .git folder is also uploaded. By default, the .git folder is ignored."
	ClientId              = "If you have access to more than one client, you must specify the client ID for which the\nproject is to be created."
	ConfirmDelete         = "If set, zCLI will not ask for confirmation."
	LogLimitFlag          = "How many of the most recent log messages will be returned. Allowed interval is <1;1000>.\nDefault value = 100."
	LogMinSeverityFlag    = "Returns log messages with requested or higher severity. Set either severity number in the interval\n<0;7> or one of following severity codes:\nEMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFORMATIONAL, DEBUG."
	LogMsgTypeFlag        = "Select either APPLICATION or WEBSERVER log messages to be returned. Default value = APPLICATION."
	LogFollowFlag         = "If set, zCLI will continuously poll for new log messages. By default, the command will exit\nonce there are no more logs to display. To exit from this mode, use Control-C."
	LogFormatFlag         = "The format of returned log messages. Following formats are supported: \nFULL: This is the default format. Messages will be returned in the complete Syslog format. \nSHORT: Returns only timestamp and log message.\nJSON: Messages will be returned as one JSON object.\nJSONSTREAM: Messages will be returned as stream of JSON objects."
	LogFormatTemplateFlag = "Set a custom log format. Can be used only with --format=FULL.\nExample: --formatTemplate=\"{{.timestamp}} {{.severity}} {{.facility}} {{.message}}\".\nSupports standard GoLang template format and functions."
	MtuFlag               = "Sets a custom MTU for VPN interface. Default value is 1420."
	PreferredPortFlag     = "????"

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
	RegionUrl          = "zerops region file url"

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
	ProjectNotFound              = "project not found"
	ProjectWrongId               = "Please, provide correct project ID."
	ProjectsWithSameName         = "found multiple projects with the same name"
	AvailableProjectIds          = "available project IDs are: "
	ProjectNameOrIdEmpty         = "project name or ID must be filled"
	ProjectDeleteConfirm         = "Please confirm that you would like to delete the project (y/n): "
	ServiceNotFound              = "service not found"
	ServiceNameIsEmpty           = "service name must be filled"
	ServiceDeleteConfirm         = "Please confirm that you would like to delete the service (y/n): "
	ProcessInit                  = " command initialized"
	Success                      = " successfully"
	ProjectStart                 = "project start"
	ProjectStop                  = "project stop"
	ProjectDelete                = "project delete"
	ProjectStarted               = "project started"
	ProjectStopped               = "project stopped"
	ProjectDeleted               = "project deleted"
	ProjectCreated               = "project created"
	ServiceStart                 = "service start"
	ServiceStop                  = "service stop"
	ServiceDelete                = "service delete"
	ServiceStarted               = "service started"
	ServiceStopped               = "service stopped"
	ServiceDeleted               = "service deleted"
	ProjectImported              = "project imported"
	ServiceImported              = "service(s) imported"
	LogLimitInvalid              = "Invalid --limit value. Allowed interval is <1;1000>"
	LogMinSeverityInvalid        = "Invalid --minimumSeverity value."
	LogMinSeverityStringLimitErr = "Allowed values are EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFORMATIONAL, DEBUG."
	LogMinSeverityNumLimitErr    = "Allowed interval is <0;7>."
	LogFormatInvalid             = "Invalid --format value. Allowed values are FULL, SHORT, JSON, JSONSTREAM."
	LogFormatTemplateMismatch    = "--formatTemplate can be used only in combination with --format=FULL."
	LogFormatStreamMismatch      = "--format=JSON cannot be used in combination with --follow. Use --format=JSONSTREAM instead."
	LogServiceNameInvalid        = "Invalid serviceName value. Multiple @ characters are not supported. See -h for help."
	LogFormatTemplateInvalid     = "Invalid --formatTemplate content. The custom template failed with following error:"
	LogFormatTemplateNoSpace     = "Template items must be split by a (single) space."
	LogSuffixInvalid             = "Invalid serviceName value. Use <serviceName>@<int> to  return log messages from the N-th runtime container only.\nUse <serviceName>@BUILD to return log messages from the last build if available."
	LogRuntimeOnly               = "This command can be used on runtime services only."
	LogNoContainerFound          = "No runtime container was found."
	LogTooFewContainers          = "There %s only %d runtime container%s at the moment. Select a lower container index."
	LogNoBuildFound              = "No build was found for this service."
	LogBuildStatusUploading      = "Service status UPLOADING, need to wait for app version data."
	LogAccessFailed              = "Request for access to logs failed."
	LogMsgTypeInvalid            = "Invalid --messageType value. Allowed values are APPLICATION, WEBSERVER."
	LogReadingFailed             = "Log reading failed."

	// deploy
	DeployHintPush                   = "To build your application in Zerops, use the zcli push command instead."
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

	// vpn
	VpnDaemonUnavailable = "Zerops VPN daemon needs to be installed on your machine to handle VPN connections."

	// vpn start
	VpnStartInstallDaemonPrompt               = "Do you want to install the VPN daemon?\nYou will be prompted for your root/administrator password to confirm the installation."
	VpnStartTerminatedByUser                  = "when you are ready, try `/path/to/zcli daemon install`"
	VpnStartUserIsUnableToWriteYorN           = "type 'y' or 'n' please"
	VpnStartWireguardUtunError                = "we failed to start vpn, there is a possibility that you have another vpn. If so, try to shut it down"
	VpnStartVpnNotReachable                   = "zerops vpn servers aren't reachable"
	VpnStartTunnelIsNotAlive                  = "we failed to establish zerops vpn"
	VpnStartUnableToConfigureNetworkInterface = "unable to configure network interface"
	VpnStartUnableToUpdateRoutingTable        = "unable to update routing table"
	VpnStartNetworkInterfaceNotFound          = "network interface not found"
	VpnStartInvalidServerPublicKey            = "invalid server public key"
	VpnStartInvalidVpnAddress                 = "invalid VPN address"
	VpnStartTunnelConfigurationFailed         = "tunnel configuration failed"

	// vpn status
	VpnStatusTunnelStatusActive      = "wireguard tunnel is working properly"
	VpnStatusTunnelStatusSetInactive = "wireguard tunnel isn't established, try `/path/to/zcli vpn start` command"
	VpnStatusDnsStatusActive         = "dns is working properly"
	VpnStatusDnsStatusSetInactive    = "dns isn't set, try `/path/to/zcli vpn start` command"
	VpnStatusAdditionalInfo          = "additional info:"
	VpnStatusDnsCheckError           = "we failed to check that dns is working correctly"
	VpnStatusDnsInterfaceNotFound    = "vpn interface not found"
	VpnStatusWireguardNotAvailable   = "wireguard interface not available"
	VpnStatusCheckInvalidAddress     = "invalid address"

	// vpn stop
	VpnStopSuccess                       = "vpn connection was closed"
	VpnStopAdditionalInfo                = "additional info:"
	VpnStopAdditionalInfoMessage         = "dns could be set by yourself, if so it must be removed manually"
	VpnStopUnableToRemoveTunnelInterface = "unable to remove tunnel interface"

	// daemon
	DaemonInstallerDesc             = "zerops daemon"
	DaemonUnableToSaveConfiguration = "unable to save configuration"
	DaemonElevated                  = "operation continues in a new window"
	PathNotFound                    = "path not found"

	// daemon install
	DaemonInstallSuccess                 = "zerops daemon has been installed"
	DaemonInstallWireguardNotFound       = "wireguard was not found"
	DaemonInstallWireguardNotFoundDarwin = "wireguard was not found, try `brew install wireguard-tools`"

	// daemon remove
	DaemonRemoveStopVpnUnavailable = "zerops daemon isn't running, vpn couldn't be removed"
	DaemonRemoveSuccess            = "zerops daemon has been removed"

	// S3
	BucketGenericXAmzAcl              = "Defines one of predefined grants, known as canned ACLs.\nValid values are: private, public-read, public-read-write, authenticated-read."
	BucketGenericXAmzAclInvalid       = "Invalid --x-amz-acl value. Allowed values are: private, public-read, public-read-write, authenticated-read."
	BucketGenericOnlyForObjectStorage = "This command can be used on object storage services only."
	BucketGenericBucketNamePrefixed   = "Bucket names are prefixed by object storage service ID to make the bucket names unique.\nLearn more about bucket naming conventions at https://docs.zerops.io/documentation/services/storage/s3.html#used-technology"

	BucketCreated                 = "Bucket created"
	BucketCreateCreatingDirect    = "Creating bucket %s directly on S3 API.\n"
	BucketCreateCreatingZeropsApi = "Creating bucket %s using Zerops API.\n"

	BucketDeleted                 = "Bucket deleted"
	BucketDeleteDeletingDirect    = "Deleting bucket %s directly on S3 API.\n"
	BucketDeleteDeletingZeropsApi = "Deleting bucket %s using Zerops API.\n"

	BucketS3Region              = "When using direct S3 API choose one of Zerops regions.\nUse the \"zcli region list\" command to list all Zerops regions.\nAutomatically filled if the REGION environment variable exists or the user is logged in."
	BucketS3AccessKeyId         = "When using direct S3 API the accessKeyId to the Zerops object storage is required.\nAutomatically filled if the {serviceName}_accessKeyId environment variable exists."
	BucketS3SecretAccessKey     = "When using direct S3 API the secretAccessKey to the Zerops object storage is required.\nAutomatically filled if the {serviceName}_secretAccessKey environment variable exists."
	BucketS3FlagBothMandatory   = "If you are specifying accessKeyId or secretAccessKey, both flags are mandatory."
	BucketS3EnvBothMandatory    = "If you are using env for accessKeyId or secretAccessKey, both env variables must be set."
	BucketS3RequestFailed       = "S3 API request failed: %s"
	BucketS3BucketAlreadyExists = "The bucket name already exists under a different object storage user. Set a different bucket name."

	// generic
	UnauthenticatedUser = `unauthenticated user, login before proceeding with this command
zcli login {token | username password}
more info: https://docs.zerops.io/documentation/cli/authorization.html`

	GrpcApiTimeout    = "zerops api didn't respond within assigned time, try it again later"
	GrpcVpnApiTimeout = "zerops vpn server didn't respond within assigned time, try it again later"

	HintChangeRegion    = "hint: try to change your region (you can list available regions using `zcli region list`)"
	InternalServerError = "internal server error"
)

func AddHintChangeRegion(err error) error {
	return fmt.Errorf("%w\n%s", err, HintChangeRegion)
}
