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

const (
	// help
	LoginHelp                = "LoginHelp"
	ProjectHelp              = "ProjectHelp"
	ProjectListHelp          = "ProjectListHelp"
	ScopeHelp                = "ScopeHelp"
	ScopeProjectHelp         = "ScopeProjectHelp"
	ScopeResetHelp           = "ScopeResetHelp"
	ProjectDeleteHelp        = "ProjectDeleteHelp"
	ProjectImportHelp        = "ProjectImportHelp"
	ProjectServiceImportHelp = "ProjectServiceImportHelp"
	ServiceHelp              = "ServiceHelp"
	ServiceStartHelp         = "ServiceStartHelp"
	ServiceStopHelp          = "ServiceStopHelp"
	ServiceDeleteHelp        = "ServiceDeleteHelp"
	ServiceLogHelp           = "ServiceLogHelp"
	ServiceDeployHelp        = "ServiceDeployHelp"
	ServiceListHelp          = "ServiceListHelp"
	ServicePushHelp          = "ServicePushHelp"
	StatusShowDebugLogsHelp  = "StatusShowDebugLogsHelp"
	VersionHelp              = "VersionHelp"
	VpnHelp                  = "VpnHelp"
	VpnUpHelp                = "VpnUpHelp"
	VpnDownHelp              = "VpnDownHelp"

	// cmd short
	CmdDeployDesc          = "CmdDeployDesc"
	CmdPushDesc            = "CmdPushDesc"
	CmdLogin               = "CmdLogin"
	CmdStatusShowDebugLogs = "CmdStatusShowDebugLogs"
	CmdVersion             = "CmdVersion"
	CmdProject             = "CmdProject"
	CmdService             = "CmdService"
	CmdProjectList         = "CmdProjectList"
	CmdScope               = "CmdScope"
	CmdScopeProject        = "CmdScopeProject"
	CmdScopeReset          = "CmdScopeReset"
	CmdProjectDelete       = "CmdProjectDelete"
	CmdProjectImport       = "CmdProjectImport"
	CmdServiceList         = "CmdServiceList"
	CmdServiceImport       = "CmdServiceImport"
	CmdServiceStart        = "CmdServiceStart"
	CmdServiceStop         = "CmdServiceStop"
	CmdServiceDelete       = "CmdServiceDelete"
	CmdServiceLog          = "CmdServiceLog"
	CmdVpn                 = "CmdVpn"
	CmdVpnUp               = "CmdVpnUp"
	CmdVpnDown             = "CmdVpnDown"

	// cmd long
	CmdProjectImportLong = "CmdProjectImportLong"
	DeployDescLong       = "DeployDescLong"
	PushDescLong         = "PushDescLong"
	CmdServiceLogLong    = "CmdServiceLogLong"
	ServiceLogAdditional = "ServiceLogAdditional"

	// flags description
	RegionFlag            = "RegionFlag"
	RegionUrlFlag         = "RegionUrlFlag"
	BuildVersionName      = "BuildVersionName"
	SourceName            = "SourceName"
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

	// process
	ProcessInvalidState = "ProcessInvalidState"

	// archiveClient
	ArchClientWorkingDirectory  = "ArchClientWorkingDirectory"
	ArchClientMaxOneTilde       = "ArchClientMaxOneTilde"
	ArchClientPackingDirectory  = "ArchClientPackingDirectory"
	ArchClientPackingFile       = "ArchClientPackingFile"
	ArchClientFileAlreadyExists = "ArchClientFileAlreadyExists"

	// login
	LoginSuccess = "LoginSuccess"

	// region
	RegionNotFound        = "RegionNotFound"
	RegionTableColumnName = "RegionTableColumnName"

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

	// project + service
	ProjectDeleteConfirm = "ProjectDeleteConfirm"
	ServiceDeleteConfirm = "ServiceDeleteConfirm"
	ProjectDeleting      = "ProjectDeleting"
	ProjectDeleted       = "ProjectDeleted"
	ServiceStarting      = "ServiceStarting"
	ServiceStarted       = "ServiceStarted"
	ServiceStopping      = "ServiceStopping"
	ServiceStopped       = "ServiceStopped"
	ServiceDeleting      = "ServiceDeleting"
	ServiceDeleted       = "ServiceDeleted"
	ProjectImported      = "ProjectImported"
	ServiceImported      = "ServiceImported"

	// service logs
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

	// push
	PushRunning  = "PushRunning"
	PushFinished = "PushFinished"

	// deploy
	DeployHintPush                   = "DeployHintPush"
	BuildDeployCreatingPackageStart  = "BuildDeployCreatingPackageStart"
	BuildDeployCreatingPackageDone   = "BuildDeployCreatingPackageDone"
	BuildDeployPackageSavedInto      = "BuildDeployPackageSavedInto"
	BuildDeployUploadingPackageStart = "BuildDeployUploadingPackageStart"
	BuildDeployUploadingPackageDone  = "BuildDeployUploadingPackageDone"
	BuildDeployUploadPackageFailed   = "BuildDeployUploadPackageFailed"
	BuildDeployDeployingStart        = "BuildDeployDeployingStart"
	BuildDeployZeropsYamlEmpty       = "BuildDeployZeropsYamlEmpty"
	BuildDeployZeropsYamlTooLarge    = "BuildDeployZeropsYamlTooLarge"
	BuildDeployZeropsYamlFound       = "BuildDeployZeropsYamlFound"
	BuildDeployZeropsYamlNotFound    = "BuildDeployZeropsYamlNotFound"

	// status info
	StatusInfoCliDataFilePath  = "StatusInfoCliDataFilePath"
	StatusInfoLogFilePath      = "StatusInfoLogFilePath"
	StatusInfoWgConfigFilePath = "StatusInfoWgConfigFilePath"
	StatusInfoLoggedUser       = "StatusInfoLoggedUser"
	StatusInfoVpnStatus        = "StatusInfoVpnStatus"

	// debug logs
	DebugLogsNotFound = "DebugLogsNotFound"

	// vpn
	VpnUp                            = "VpnUp"
	VpnDown                          = "VpnDown"
	VpnConfigSaved                   = "VpnConfigSaved"
	VpnPrivateKeyCorrupted           = "VpnPrivateKeyCorrupted"
	VpnPrivateKeyCreated             = "VpnPrivateKeyCreated"
	VpnWgQuickIsNotInstalled         = "VpnWgQuickIsNotInstalled"
	VpnDisconnectionPrompt           = "VpnDisconnectionPrompt"
	VpnDisconnectionPromptNo         = "VpnDisconnectionPromptNo"
	VpnPingFailed                    = "VpnPingFailed"
	VpnCheckingConnection            = "VpnCheckingConnection"
	VpnCheckingConnectionIsActive    = "VpnCheckingConnectionIsActive"
	VpnCheckingConnectionIsNotActive = "VpnCheckingConnectionIsNotActive"

	////////////
	// global //
	////////////
	CliTerminalModeEnvVar = "TerminalModeEnv"
	CliLogFilePathEnvVar  = "CliLogFilePathEnvVar"
	CliDataFilePathEnvVar = "CliDataFilePathEnvVar"

	UnknownTerminalMode    = "UnknownTerminalMode"
	UnableToDecodeJsonFile = "UnableToDecodeJsonFile"
	UnableToWriteCliData   = "UnableToWriteCliData"
	UnableToWriteLogFile   = "UnableToWriteLogFile"

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
	ErrorInvalidServiceIdOrName = "ErrorInvalidServiceIdOrName"
)
