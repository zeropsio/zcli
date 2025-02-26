package i18n

import "fmt"

var en = map[string]string{
	// root
	GuestWelcome: `Welcome to zCli by Zerops!

To unlock the full potential of zCLI, you need to log in using your Zerops account.
Logging in enables you to access various features and interact with Zerops services seamlessly.

To log in, simply use the following command: zcli login <your_token>
Replace <your_token> with the authentication token generated from your Zerops account.
Once logged in, you'll be able to manage projects, deploy applications, configure VPN,
and much more directly from the command line interface.

If you encounter any issues during the login process or have any questions,
feel free to find out how to contact our support team by running 'zcli support'.`,
	LoggedWelcome: `Welcome in Zerops!
You are logged as %s
and your %s.`,

	// env
	GlobalEnvVariables:        "Global Env Variables:",
	CurrentlyUsedEnvVariables: "Curently used variables:",

	// login
	CmdHelpLogin:          "the login command.",
	CmdDescLogin:          "Login into Zerops with generated Zerops token",
	LoginSuccess:          "You are logged as %s <%s>",
	RegionNotFound:        "Selected region %s not found",
	RegionTableColumnName: "Name",

	// logout
	CmdHelpLogout:          "the logout command.",
	CmdDescLogout:          "Disconnect from VPN and log out from your Zerops account",
	LogoutVpnDisconnecting: "Disconnecting from VPN. Please provide your password if prompted.",
	LogoutSuccess:          "Successfully logged out. You are now disconnected from Zerops services.",

	// scope
	CmdHelpScope: "the scope command.",
	CmdDescScope: "Scope commands group",

	// scope project
	CmdHelpScopeProject: "the scope project command.",
	CmdDescScopeProject: "Sets the scope for project. All commands that require project ID will use the selected one.",

	// scope reset
	CmdHelpScopeReset: "the scope reset command.",
	CmdDescScopeReset: "Resets the scope for project and service.",

	// project
	CmdHelpProject: "the project command.",
	CmdDescProject: "Project commands group",

	// project lit
	CmdHelpProjectList: "the project list command.",
	CmdDescProjectList: "Lists all projects.",

	// project delete
	CmdHelpProjectDelete: "the project delete command.",
	CmdDescProjectDelete: "Deletes a project and all of its services.",
	ProjectDeleteConfirm: "Project %s will be deleted? \n Are you sure?",
	ServiceDeleteConfirm: "Service %s will be deleted? \n Are you sure?",
	ProjectDeleting:      "Project is being deleted",
	ProjectDeleteFailed:  "Project deletion failed",
	ProjectDeleted:       "Project was deleted",

	// project import
	CmdHelpProjectImport:     "The project import command. Use \"-\" as importYamlPath for taking yaml content from stdin.",
	CmdDescProjectImport:     "Creates a new project with one or more services.",
	CmdDescProjectImportLong: "Creates a new project with one or more services according to the definition in the import YAML file.",
	ProjectImported:          "project imported",

	// project service import
	CmdHelpProjectServiceImport: "The project service import command. Use \"-\" as importYamlPath for taking yaml content from stdin.",
	CmdDescProjectServiceImport: "Creates one or more Zerops services in an existing project.",
	ServiceImported:             "service(s) imported",

	// service
	CmdHelpService: "the service command.",
	CmdDescService: "Zerops service commands group",

	// service start
	CmdHelpServiceStart: "the service start command.",
	CmdDescServiceStart: "Starts the Zerops service.",
	ServiceStarting:     "Service is being started",
	ServiceStartFailed:  "Service start failed",
	ServiceStarted:      "Service was started",

	// service stop
	CmdHelpServiceStop: "the enable Zerops subdomain command.",
	CmdDescServiceStop: "Starts the Zerops service.",
	ServiceStopping:    "Service is being stopped",
	ServiceStopFailed:  "Service stop failed",
	ServiceStopped:     "Service was stopped",

	// service delete
	CmdHelpServiceDelete: "the service delete command.",
	CmdDescServiceDelete: "Deletes the Zerops service.",
	ServiceDeleting:      "Service is being deleted",
	ServiceDeleteFailed:  "Service deletion failed",
	ServiceDeleted:       "Service was deleted",

	// service log
	CmdHelpServiceLog: "the service log command.",
	CmdDescServiceLog: "Get service runtime or build log to stdout.",
	CmdDescServiceLogLong: "Returns service runtime or build log to stdout. By default, the command returns the last 100\n" +
		"log messages from all service runtime containers and exits.\n\n" +
		"Use the <serviceName> alone in the command to return log messages from all runtime containers.\n" +
		"Set <serviceName>@1 to return log messages from the first runtime container only.\n" +
		"Set <serviceName>@build to return log messages from the last build if available.",
	LogLimitInvalid:              "Invalid --limit value. Allowed interval is <1;1000>",
	LogMinSeverityInvalid:        "Invalid --minimumSeverity value.",
	LogMinSeverityStringLimitErr: "Allowed values are EMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFORMATIONAL, DEBUG.",
	LogMinSeverityNumLimitErr:    "Allowed interval is <0;7>.",
	LogFormatInvalid:             "Invalid --format value. Allowed values are FULL, SHORT, JSON, JSONSTREAM.",
	LogFormatTemplateMismatch:    "--formatTemplate can be used only in combination with --format=FULL.",
	LogFormatStreamMismatch:      "--format=JSON cannot be used in combination with --follow. Use --format=JSONSTREAM instead.",
	LogFormatTemplateInvalid:     "Invalid --formatTemplate content. The custom template failed with following error:",
	LogFormatTemplateNoSpace:     "Template items must be split by a (single) space.",
	LogNoBuildFound:              "No build was found for this service.",
	LogBuildStatusUploading:      "Service status UPLOADING, need to wait for app version data.",
	LogAccessFailed:              "Request for access to logs failed.",
	LogMsgTypeInvalid:            "Invalid --messageType value. Allowed values are APPLICATION, WEBSERVER.",
	LogReadingFailed:             "Log reading failed.",

	// service deploy
	CmdHelpServiceDeploy: "the service deploy command.",
	CmdDescDeploy:        "Deploys your application to Zerops.",
	CmdDescDeployLong: "Deploys your application to Zerops. \n\n" +
		"pathToFileOrDir defines a path to one or more directories and/or files relative to the working\n" +
		"directory. The working directory is by default the current directory and can be changed\n" +
		"using the --workingDir flag. zCLI deploys selected directories and/or files to Zerops. \n\n" +
		"To build your application in Zerops, use the zcli push command instead.",
	DeployRunning:  "Deploy is running",
	DeployFailed:   "Deploy failed",
	DeployFinished: "Deploy finished",

	// push
	CmdHelpPush: "the service push command.",
	CmdDescPush: "Builds your application in Zerops and deploys it",
	CmdDescPushLong: "Builds your application in Zerops and deploys it. \n\n" +
		"The command triggers the build pipeline defined in zerops.yml. Zerops.yml must be in the working\n" +
		"directory. The working directory is by default the current directory and can be changed\n" +
		"using the --workingDir flag. zCLI uploads all files and subdirectories of the working\n" +
		"directory to Zerops and starts the build pipeline. Files found in the .gitignore\n" +
		"file will be ignored.\n\n" +
		"If you just want to deploy your application to Zerops, use the zcli deploy command instead.",
	PushRunning:  "Push is running",
	PushFinished: "Push finished",
	PushFailed:   "Push failed",

	// push && deploy
	PushDeployCreatingPackageStart:  "creating package",
	PushDeployCreatingPackageDone:   "package created",
	PushDeployPackageSavedInto:      "package file saved into: %s",
	PushDeployUploadingPackageStart: "uploading package",
	PushDeployUploadingPackageDone:  "package uploaded",
	PushDeployUploadPackageFailed:   "package upload failed",
	PushDeployDeployingStart:        "deploying service",
	PushDeployZeropsYamlEmpty:       "config file zerops.yml is empty",
	PushDeployZeropsYamlTooLarge:    "max. size of zerops.yml is 10 KB",
	PushDeployZeropsYamlFound:       "File zerops.yml found. Path: %s.",
	PushDeployZeropsYamlNotFound: "File zerops.yml not found. Checked paths: [%s]. \n" +
		" Please, create a zerops.yml file in the root directory of your project. \n" +
		" Alternatively you can use the --zeropsYaml flag to specify the path to the zerops.yml file or \n" +
		" use the --workingDir flag to set the working directory to the directory where the zerops.yml file is located.",

	// service list
	CmdHelpServiceList: "the service list command.",
	CmdDescServiceList: "Lists all services in the project.",

	// service enable subdomain
	CmdHelpServiceEnableSubdomain: "the service stop command.",
	CmdDescServiceEnableSubdomain: "Enables access through Zerops subdomain.",
	ServiceEnablingSubdomain:      "enabling subdomain access",
	ServiceEnableSubdomainFailed:  "subdomain access enabling failed",
	ServiceEnabledSubdomain:       "subdomain access enabled",

	// status show debug logs
	CmdHelpStatusShowDebugLogs: "the status show debug logs command.",
	CmdDescStatusShowDebugLogs: "Shows zCLI debug logs",
	DebugLogsNotFound:          "Debug logs not found",

	// update
	CmdHelpUpdate: "the update command",
	CmdDescUpdate: "Updates zCLI to the latest version",

	// version
	CmdHelpVersion: "the version command.",
	CmdDescVersion: "Shows the current zCLI version",

	// support
	CmdHelpSupport: "the support command.",
	CmdDescSupport: "How to contact Zerops support for assistance",
	Contact:        "You can contact Zerops support via:",
	Documentation: `Additionally, you can explore our documentation
at https://docs.zerops.io/references/cli for further details.`,

	// env
	CmdHelpEnv: "the env command.",
	CmdDescEnv: "Displays global environment variables, their paths and additional options",

	// vpn
	CmdHelpVpn: "the vpn command.",
	CmdDescVpn: "VPN commands group",

	// vpn up
	CmdHelpVpnUp: "the vpn up command.",
	CmdDescVpnUp: "Connects to the Zerops VPN.",
	VpnUp:        "VPN connected",

	VpnConfigSaved:           "VPN config saved",
	VpnPrivateKeyCorrupted:   "VPN private key corrupted, a new one will be created",
	VpnPrivateKeyCreated:     "VPN private key created",
	VpnDisconnectionPrompt:   "VPN is active, do you want to disconnect?",
	VpnDisconnectionPromptNo: "VPN is active, you can disconnect using the 'zcli vpn down' command",
	VpnCheckingConnection:    "Checking VPN connection",
	VpnPingFailed: fmt.Sprintf("Wireguard adapter was created, but we are not able to establish a connection,"+
		"this could indicate a problem on our side. Please contact our support team via %s, %s or join our discord %s.", CustomerSupportLink, CustomerSupportEmail, DiscordCommunityLink),

	// vpn down
	CmdHelpVpnDown: "the vpn down command.",
	CmdDescVpnDown: "Disconnects from the Zerops VPN.",
	VpnDown:        "VPN disconnected",

	// vpn shared
	VpnWgQuickIsNotInstalled:        "wg-quick is not installed, please visit https://www.wireguard.com/install/",
	VpnResolveCtlIsNotInstalled:     "resolvectl is not installed, please install systemd-resolved or equivalent for your distribution",
	VpnWgQuickIsNotInstalledWindows: "wireguard is not installed, please visit https://www.wireguard.com/install/",

	// flags description
	RegionFlag:            "Choose one of Zerops regions. Use the \"zcli region list\" command to list all Zerops regions.",
	RegionUrlFlag:         "Zerops region file url.",
	BuildVersionName:      "Adds a custom version name. Automatically filled if the VERSIONNAME environment variable exists.",
	BuildWorkingDir:       "Sets a custom working directory. Default working directory is the current directory.",
	BuildArchiveFilePath:  "If set, zCLI creates a tar.gz archive with the application code in the required path relative\nto the working directory. By default, no archive is created.",
	ZeropsYamlLocation:    "Sets a custom path to the zerops.yml file relative to the working directory. By default zCLI\nlooks for zerops.yml in the working directory.",
	UploadGitFolder:       "If set, zCLI the .git folder is also uploaded. By default, the .git folder is ignored.",
	OrgIdFlag:             "If you have access to more than one organization, you must specify the org ID for which the\nproject is to be created.",
	LogLimitFlag:          "How many of the most recent log messages will be returned. Allowed interval is <1;1000>.\nDefault value = 100.",
	LogMinSeverityFlag:    "Returns log messages with requested or higher severity. Set either severity number in the interval\n<0;7> or one of following severity codes:\nEMERGENCY, ALERT, CRITICAL, ERROR, WARNING, NOTICE, INFORMATIONAL, DEBUG.",
	LogMsgTypeFlag:        "Select either APPLICATION or WEBSERVER log messages to be returned. Default value = APPLICATION.",
	LogShowBuildFlag:      "If set, zCLI will return build log messages instead of runtime log messages.",
	LogFollowFlag:         "If set, zCLI will continuously poll for new log messages. By default, the command will exit\nonce there are no more logs to display. To exit from this mode, use Control-C.",
	LogFormatFlag:         "The format of returned log messages. Following formats are supported: \nFULL: This is the default format. Messages will be returned in the complete Syslog format. \nSHORT: Returns only timestamp and log message.\nJSON: Messages will be returned as one JSON object.\nJSONSTREAM: Messages will be returned as stream of JSON objects.",
	LogFormatTemplateFlag: "Set a custom log format. Can be used only with --format=FULL.\nExample: --formatTemplate=\"{{.timestamp}} {{.severity}} {{.facility}} {{.message}}\".\nSupports standard GoLang template format and functions.",
	ConfirmFlag:           "If set, zCLI will not ask for confirmation of destructive operations.",
	ServiceIdFlag:         "If you have access to more than one service, you must specify the service ID for which the\ncommand is to be executed.",
	ProjectIdFlag:         "If you have access to more than one project, you must specify the project ID for which the\ncommand is to be executed.",
	VpnAutoDisconnectFlag: "If set, zCLI will automatically disconnect from the VPN if it is already connected.",
	VpnMtuFlag:            "If set, Wireguard interface will use this value for MTU. If VPN is not working, try a lower value.",
	ZeropsYamlSetup:       "Choose setup to be used from zerops.yml.",

	// archiveClient
	ArchClientWorkingDirectory:  "working directory: %s",
	ArchClientMaxOneTilde:       "only one ~(tilde) is allowed",
	ArchClientPackingDirectory:  "packing directory: %s",
	ArchClientPackingFile:       "packing file: %s",
	ArchClientFileAlreadyExists: "file [%s] already exists",
	ArchClientFileIgnored:       "file ignored: %s",

	// import
	ImportYamlOk:        "Yaml file was checked",
	ImportYamlEmpty:     "Config file import yaml is empty",
	ImportYamlTooLarge:  "Max. size of import yaml is 100 KB",
	ImportYamlFound:     "Import yaml found",
	ImportYamlNotFound:  "Import yaml not found",
	ImportYamlCorrupted: "Import yaml corrupted",
	ServiceCount:        "Number of services to be added: %d",
	QueuedProcesses:     "Queued processes: %d",
	CoreServices:        "Core services activation started",

	// status info
	StatusInfoCliDataFilePath:        "Zerops CLI data file path",
	StatusInfoLogFilePath:            "Zerops CLI log file path",
	StatusInfoWgConfigFilePath:       "Zerops CLI wg config file path",
	StatusInfoLoggedUser:             "Logged user",
	StatusInfoVpnStatus:              "VPN status",
	VpnCheckingConnectionIsActive:    "VPN connection is active",
	VpnCheckingConnectionIsNotActive: "VPN connection is not active",

	// //////////
	// global //
	// //////////
	ProcessInvalidState: "last command has finished with error, identifier for communication with our support: %s",

	CliTerminalModeEnvVar: "If enabled provides a rich UI to communicate with a user. Possible values: auto, enabled, disabled. Default value is auto.",
	CliLogFilePathEnvVar:  "Path to a log file.",
	CliDataFilePathEnvVar: "Path to data file.",

	UnknownTerminalMode:       "Unknown terminal mode: %s. Falling back to auto-discovery. Possible values: auto, enabled, disabled.",
	UnableToDecodeJsonFile:    "Unable to decode json file: %s",
	UnableToWriteCliData:      "Unable to write zcli data, paths tested: %s",
	UnableToWriteLogFile:      "Unable to write zcli debug log file, paths tested: %s",
	UnableToWriteWgConfigFile: "Unable to write zcli wireguard config file, paths tested: %s",

	// args
	ArgsOnlyOneOptionalAllowed: "optional arg %s can be only the last one",
	ArgsOnlyOneArrayAllowed:    "array arg %s can be only the last one",
	ArgsNotEnoughRequiredArgs:  "expected at least %d arg(s), got %d",
	ArgsTooManyArgs:            "expected no more than %d arg(s), got %d",

	// ux helpers
	ProjectSelectorListEmpty:       "You don't have any projects yet. Create a new project using `zcli project import` command.",
	ProjectSelectorPrompt:          "Please, select a project",
	ProjectSelectorOutOfRangeError: "We couldn't find a project with the index you entered. Please, try again or contact our support team.",
	ServiceSelectorListEmpty:       "Project doesn't have any services yet. Create a new service using `zcli project service-import` command",
	ServiceSelectorPrompt:          "Please, select a service",
	ServiceSelectorOutOfRangeError: "We couldn't find a service with the index you entered. Please, try again or contact our support team.",
	OrgSelectorListEmpty:           "You don't belong to any organization yet. Please, contact our support team.",
	OrgSelectorPrompt:              "Please, select an org",
	OrgSelectorOutOfRangeError:     "We couldn't find an org with the index you entered. Please, try again or contact our support team.",
	SelectorAllowedOnlyInTerminal:  "Interactive selection can be used only in terminal mode. Use command flags to specify missing parameters.",
	PromptAllowedOnlyInTerminal:    "Interactive prompt can be used only in terminal mode. Use --confirm=true flag to confirm it",

	UnauthenticatedUser: `unauthenticated user, login before proceeding with this command
zcli login {token}
more info: https://docs.zerops.io/references/cli/`,

	// scope
	SelectedProject:         "Selected project",
	SelectedService:         "Selected service",
	ScopedProject:           "Scoped project",
	ScopedProjectNotFound:   "Scoped project wasn't found, select a different project using `zcli scope project` command.",
	PreviouslyScopedProject: "Previously scoped project",
	ScopeReset:              "Scope was reset",

	DestructiveOperationConfirmationFailed: "You have to confirm a destructive operation.",

	// errors
	ErrorInvalidProjectId:       "Invalid project ID [%s], %s", // values: projectId, message
	ErrorInvalidScopedProjectId: "Invalid ID of the scoped project [%s], select a different project using `zcli scope project` command.",
	ErrorInvalidServiceId:       "Invalid service ID [%s], %s", // values: serviceId, message
	ErrorServiceNotFound:        "Service [%s] not found",
}
