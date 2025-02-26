package i18n

import "fmt"

func T(textConst string, args ...interface{}) string {
	translation, exists := en[textConst]
	if !exists {
		return "[missing translation] " + textConst
	}
	if len(args) > 0 {
		return fmt.Sprintf(translation, args...)
	}
	return translation
}

const CustomerSupportLink = "https://support.zerops.io"
const CustomerSupportEmail = "support@zerops.io"
const DiscordCommunityLink = "https://discord.com/invite/WDvCZ54"

const (
	// root
	GuestWelcome  = "GuestWelcome"
	LoggedWelcome = "LoggedWelcome"

	// env
	GlobalEnvVariables        = "GlobalEnvVariables"
	CurrentlyUsedEnvVariables = "CurrentlyUsedEnvVariables"

	// login
	CmdHelpLogin          = "CmdHelpLogin"
	CmdDescLogin          = "CmdDescLogin"
	LoginSuccess          = "LoginSuccess"
	RegionNotFound        = "RegionNotFound"
	RegionTableColumnName = "RegionTableColumnName"

	// logout
	CmdHelpLogout          = "CmdHelpLogout"
	CmdDescLogout          = "CmdDescLogout"
	LogoutVpnDisconnecting = "LogoutVpnDisconnecting"
	LogoutSuccess          = "LogoutSuccess"

	// scope
	CmdHelpScope = "CmdHelpScope"
	CmdDescScope = "CmdDescScope"

	// scope project
	CmdHelpScopeProject = "CmdHelpScopeProject"
	CmdDescScopeProject = "CmdDescScopeProject"

	// scope reset
	CmdHelpScopeReset = "CmdHelpScopeReset"
	CmdDescScopeReset = "CmdDescScopeReset"

	// project
	CmdHelpProject = "CmdHelpProject"
	CmdDescProject = "CmdDescProject"

	// project lit
	CmdHelpProjectList = "CmdHelpProjectList"
	CmdDescProjectList = "CmdDescProjectList"

	// project delete
	CmdHelpProjectDelete = "CmdHelpProjectDelete"
	CmdDescProjectDelete = "CmdDescProjectDelete"
	ProjectDeleteConfirm = "ProjectDeleteConfirm"
	ServiceDeleteConfirm = "ServiceDeleteConfirm"
	ProjectDeleting      = "ProjectDeleting"
	ProjectDeleteFailed  = "ProjectDeleteFailed"
	ProjectDeleted       = "ProjectDeleted"

	// project import
	CmdHelpProjectImport     = "CmdHelpProjectImport"
	CmdDescProjectImport     = "CmdDescProjectImport"
	CmdDescProjectImportLong = "CmdDescProjectImportLong"
	ProjectImported          = "ProjectImported"

	// project service import
	CmdHelpProjectServiceImport = "CmdHelpProjectServiceImport"
	CmdDescProjectServiceImport = "CmdDescProjectServiceImport"
	ServiceImported             = "ServiceImported"

	// service
	CmdHelpService = "CmdHelpService"
	CmdDescService = "CmdDescService"

	// service start
	CmdHelpServiceStart = "CmdHelpServiceStart"
	CmdDescServiceStart = "CmdDescServiceStart"
	ServiceStarting     = "ServiceStarting"
	ServiceStartFailed  = "ServiceStartFailed"
	ServiceStarted      = "ServiceStarted"

	// service stop
	CmdHelpServiceStop = "CmdHelpServiceStop"
	CmdDescServiceStop = "CmdDescServiceStop"
	ServiceStopping    = "ServiceStopping"
	ServiceStopFailed  = "ServiceStopFailed"
	ServiceStopped     = "ServiceStopped"

	// service delete
	CmdHelpServiceDelete = "CmdHelpServiceDelete"
	CmdDescServiceDelete = "CmdDescServiceDelete"
	ServiceDeleting      = "ServiceDeleting"
	ServiceDeleteFailed  = "ServiceDeleteFailed"
	ServiceDeleted       = "ServiceDeleted"

	// service log
	CmdHelpServiceLog            = "CmdHelpServiceLog"
	CmdDescServiceLog            = "CmdDescServiceLog"
	CmdDescServiceLogLong        = "CmdDescServiceLogLong"
	LogLimitInvalid              = "LogLimitInvalid"
	LogMinSeverityInvalid        = "LogMinSeverityInvalid"
	LogMinSeverityStringLimitErr = "LogMinSeverityStringLimitErr"
	LogMinSeverityNumLimitErr    = "LogMinSeverityNumLimitErr"
	LogFormatInvalid             = "LogFormatInvalid"
	LogFormatTemplateMismatch    = "LogFormatTemplateMismatch"
	LogFormatStreamMismatch      = "LogFormatStreamMismatch"
	LogFormatTemplateInvalid     = "LogFormatTemplateInvalid"
	LogFormatTemplateNoSpace     = "LogFormatTemplateNoSpace"
	LogNoBuildFound              = "LogNoBuildFound"
	LogBuildStatusUploading      = "LogBuildStatusUploading"
	LogAccessFailed              = "LogAccessFailed"
	LogMsgTypeInvalid            = "LogMsgTypeInvalid"
	LogReadingFailed             = "LogReadingFailed"

	// service deploy
	CmdHelpServiceDeploy = "CmdHelpServiceDeploy"
	CmdDescDeploy        = "CmdDescDeploy"
	CmdDescDeployLong    = "CmdDescDeployLong"
	DeployRunning        = "DeployRunning"
	DeployFailed         = "DeployFailed"
	DeployFinished       = "DeployFinished"

	// push
	CmdHelpPush     = "CmdHelpPush"
	CmdDescPush     = "CmdDescPush"
	CmdDescPushLong = "CmdDescPushLong"
	PushRunning     = "PushRunning"
	PushFailed      = "PushFailed"
	PushFinished    = "PushFinished"

	// push && deploy
	PushDeployCreatingPackageStart  = "PushDeployCreatingPackageStart"
	PushDeployCreatingPackageDone   = "PushDeployCreatingPackageDone"
	PushDeployPackageSavedInto      = "PushDeployPackageSavedInto"
	PushDeployUploadingPackageStart = "PushDeployUploadingPackageStart"
	PushDeployUploadingPackageDone  = "PushDeployUploadingPackageDone"
	PushDeployUploadPackageFailed   = "PushDeployUploadPackageFailed"
	PushDeployDeployingStart        = "PushDeployDeployingStart"
	PushDeployZeropsYamlEmpty       = "PushDeployZeropsYamlEmpty"
	PushDeployZeropsYamlTooLarge    = "PushDeployZeropsYamlTooLarge"
	PushDeployZeropsYamlFound       = "PushDeployZeropsYamlFound"
	PushDeployZeropsYamlNotFound    = "PushDeployZeropsYamlNotFound"

	// service list
	CmdHelpServiceList = "CmdHelpServiceList"
	CmdDescServiceList = "CmdDescServiceList"

	// service enable subdomain
	CmdHelpServiceEnableSubdomain = "CmdHelpServiceEnableSubdomain"
	CmdDescServiceEnableSubdomain = "CmdDescServiceEnableSubdomain"
	ServiceEnablingSubdomain      = "ServiceEnablingSubdomain"
	ServiceEnableSubdomainFailed  = "ServiceEnableSubdomainFailed"
	ServiceEnabledSubdomain       = "ServiceEnabledSubdomain"

	// status show debug logs
	CmdHelpStatusShowDebugLogs = "CmdHelpStatusShowDebugLogs"
	CmdDescStatusShowDebugLogs = "CmdDescStatusShowDebugLogs"
	DebugLogsNotFound          = "DebugLogsNotFound"

	// update
	CmdHelpUpdate = "CmdHelpUpdate"
	CmdDescUpdate = "CmdDescUpdate"

	// version
	CmdHelpVersion = "CmdHelpVersion"
	CmdDescVersion = "CmdDescVersion"

	// support
	CmdHelpSupport = "CmdHelpSupport"
	CmdDescSupport = "CmdDescSupport"

	// support
	CmdHelpEnv    = "CmdHelpEnv"
	CmdDescEnv    = "CmdDescEnv"
	Contact       = "Contact"
	Documentation = "Documentation"

	// vpn
	CmdHelpVpn = "CmdHelpVpn"
	CmdDescVpn = "CmdDescVpn"

	// vpn up
	CmdHelpVpnUp             = "CmdHelpVpnUp"
	CmdDescVpnUp             = "CmdDescVpnUp"
	VpnUp                    = "VpnUp"
	VpnConfigSaved           = "VpnConfigSaved"
	VpnPrivateKeyCorrupted   = "VpnPrivateKeyCorrupted"
	VpnPrivateKeyCreated     = "VpnPrivateKeyCreated"
	VpnDisconnectionPrompt   = "VpnDisconnectionPrompt"
	VpnDisconnectionPromptNo = "VpnDisconnectionPromptNo"
	VpnCheckingConnection    = "VpnCheckingConnection"
	VpnPingFailed            = "VpnPingFailed"

	// vpn down
	CmdHelpVpnDown = "CmdHelpVpnDown"
	CmdDescVpnDown = "CmdDescVpnDown"
	VpnDown        = "VpnDown"

	// vpn shared
	VpnWgQuickIsNotInstalled        = "VpnWgQuickIsNotInstalled"
	VpnResolveCtlIsNotInstalled     = "VpnResolveCtlIsNotInstalled"
	VpnWgQuickIsNotInstalledWindows = "VpnWgQuickIsNotInstalledWindows"

	// flags description
	RegionFlag            = "RegionFlag"
	RegionUrlFlag         = "RegionUrlFlag"
	BuildVersionName      = "BuildVersionName"
	BuildWorkingDir       = "BuildWorkingDir"
	BuildArchiveFilePath  = "BuildArchiveFilePath"
	ZeropsYamlLocation    = "ZeropsYamlLocation"
	UploadGitFolder       = "UploadGitFolder"
	OrgIdFlag             = "OrgIdFlag"
	LogLimitFlag          = "LogLimitFlag"
	LogMinSeverityFlag    = "LogMinSeverityFlag"
	LogMsgTypeFlag        = "LogMsgTypeFlag"
	LogFollowFlag         = "LogFollowFlag"
	LogShowBuildFlag      = "LogShowBuildFlag"
	LogFormatFlag         = "LogFormatFlag"
	LogFormatTemplateFlag = "LogFormatTemplateFlag"
	ConfirmFlag           = "ConfirmFlag"
	ServiceIdFlag         = "ServiceIdFlag"
	ProjectIdFlag         = "ProjectIdFlag"
	VpnAutoDisconnectFlag = "VpnAutoDisconnectFlag"
	VpnMtuFlag            = "VpnMtuFlag"
	ZeropsYamlSetup       = "ZeropsYamlSetup"

	// archiveClient
	ArchClientWorkingDirectory  = "ArchClientWorkingDirectory"
	ArchClientMaxOneTilde       = "ArchClientMaxOneTilde"
	ArchClientPackingDirectory  = "ArchClientPackingDirectory"
	ArchClientPackingFile       = "ArchClientPackingFile"
	ArchClientFileAlreadyExists = "ArchClientFileAlreadyExists"
	ArchClientFileIgnored       = "ArchClientFileIgnored"

	// import
	ImportYamlOk        = "ImportYamlOk"
	ImportYamlEmpty     = "ImportYamlEmpty"
	ImportYamlTooLarge  = "ImportYamlTooLarge"
	ImportYamlFound     = "ImportYamlFound"
	ImportYamlNotFound  = "ImportYamlNotFound"
	ImportYamlCorrupted = "ImportYamlCorrupted"
	ServiceCount        = "ServiceCount"
	QueuedProcesses     = "QueuedProcesses"
	CoreServices        = "CoreServices"

	// status info
	StatusInfoCliDataFilePath        = "StatusInfoCliDataFilePath"
	StatusInfoLogFilePath            = "StatusInfoLogFilePath"
	StatusInfoWgConfigFilePath       = "StatusInfoWgConfigFilePath"
	StatusInfoLoggedUser             = "StatusInfoLoggedUser"
	StatusInfoVpnStatus              = "StatusInfoVpnStatus"
	VpnCheckingConnectionIsActive    = "VpnCheckingConnectionIsActive"
	VpnCheckingConnectionIsNotActive = "VpnCheckingConnectionIsNotActive"

	// //////////
	// global //
	// //////////
	ProcessInvalidState = "ProcessInvalidState"

	CliTerminalModeEnvVar = "TerminalModeEnv"
	CliLogFilePathEnvVar  = "CliLogFilePathEnvVar"
	CliDataFilePathEnvVar = "CliDataFilePathEnvVar"

	UnknownTerminalMode       = "UnknownTerminalMode"
	UnableToDecodeJsonFile    = "UnableToDecodeJsonFile"
	UnableToWriteCliData      = "UnableToWriteCliData"
	UnableToWriteLogFile      = "UnableToWriteLogFile"
	UnableToWriteWgConfigFile = "UnableToWriteWgConfigFile"

	// args
	ArgsOnlyOneOptionalAllowed = "ArgsOnlyOneOptionalAllowed"
	ArgsOnlyOneArrayAllowed    = "ArgsOnlyOneArrayAllowed"
	ArgsNotEnoughRequiredArgs  = "ArgsNotEnoughRequiredArgs"
	ArgsTooManyArgs            = "ArgsTooManyArgs"

	// ux helpers
	ProjectSelectorListEmpty       = "ProjectSelectorListEmpty"
	ProjectSelectorPrompt          = "ProjectSelectorPrompt"
	ProjectSelectorOutOfRangeError = "ProjectSelectorOutOfRangeError"
	ServiceSelectorListEmpty       = "ServiceSelectorListEmpty"
	ServiceSelectorPrompt          = "ServiceSelectorPrompt"
	ServiceSelectorOutOfRangeError = "ServiceSelectorOutOfRangeError"
	OrgSelectorListEmpty           = "OrgSelectorListEmpty"
	OrgSelectorPrompt              = "OrgSelectorPrompt"
	OrgSelectorOutOfRangeError     = "OrgSelectorOutOfRangeError"
	SelectorAllowedOnlyInTerminal  = "SelectorAllowedOnlyInTerminal"
	PromptAllowedOnlyInTerminal    = "PromptAllowedOnlyInTerminal"

	UnauthenticatedUser = "UnauthenticatedUser"

	// scope
	SelectedProject         = "SelectedProject"
	SelectedService         = "SelectedService"
	ScopedProject           = "ScopedProject"
	ScopedProjectNotFound   = "ScopedProjectNotFound"
	PreviouslyScopedProject = "PreviouslyScopedProject"
	ScopeReset              = "ScopeReset"

	DestructiveOperationConfirmationFailed = "DestructiveOperationConfirmationFailed"

	// errors
	ErrorInvalidProjectId       = "ErrorInvalidProjectId"
	ErrorInvalidScopedProjectId = "ErrorInvalidScopedProjectId"
	ErrorInvalidServiceId       = "ErrorInvalidServiceId"
	ErrorServiceNotFound        = "ErrorServiceNotFound"
)
