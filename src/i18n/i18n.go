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
	DisplayHelp       = "DisplayHelp"
	GroupHelp         = "GroupHelp"
	DeployHelp        = "DeployHelp"
	LogShowHelp       = "LogShowHelp"
	LoginHelp         = "LoginHelp"
	ProjectHelp       = "ProjectHelp"
	ProjectStartHelp  = "ProjectStartHelp"
	ProjectStopHelp   = "ProjectStopHelp"
	ProjectListHelp   = "ProjectListHelp"
	ScopeHelp         = "ScopeHelp"
	ScopeProjectHelp  = "ScopeProjectHelp"
	ScopeServiceHelp  = "ScopeServiceHelp"
	ScopeResetHelp    = "ScopeResetHelp"
	ProjectDeleteHelp = "ProjectDeleteHelp"
	ProjectImportHelp = "ProjectImportHelp"
	PushHelp          = "PushHelp"
	RegionListHelp    = "RegionListHelp"
	ServiceStartHelp  = "ServiceStartHelp"
	ServiceStopHelp   = "ServiceStopHelp"
	ServiceImportHelp = "ServiceImportHelp"
	ServiceDeleteHelp = "ServiceDeleteHelp"
	ServiceLogHelp    = "ServiceLogHelp"
	VersionHelp       = "VersionHelp"
	BucketCreateHelp  = "BucketCreateHelp"
	BucketDeleteHelp  = "BucketDeleteHelp"

	// cmd short
	CmdDeployDesc          = "CmdDeployDesc"
	CmdPushDesc            = "CmdPushDesc"
	CmdLogin               = "CmdLogin"
	CmdStatus              = "CmdStatus"
	CmdStatusInfo          = "CmdStatusInfo"
	CmdStatusShowDebugLogs = "CmdStatusShowDebugLogs"
	CmdVersion             = "CmdVersion"
	CmdRegion              = "CmdRegion"
	CmdRegionList          = "CmdRegionList"
	CmdProject             = "CmdProject"
	CmdService             = "CmdService"
	CmdProjectStart        = "CmdProjectStart"
	CmdProjectStop         = "CmdProjectStop"
	CmdProjectList         = "CmdProjectList"
	CmdScope               = "CmdScope"
	CmdScopeProject        = "CmdScopeProject"
	CmdScopeService        = "CmdScopeService"
	CmdScopeReset          = "CmdScopeReset"
	CmdProjectDelete       = "CmdProjectDelete"
	CmdProjectImport       = "CmdProjectImport"
	CmdServiceImport       = "CmdServiceImport"
	CmdServiceStart        = "CmdServiceStart"
	CmdServiceStop         = "CmdServiceStop"
	CmdServiceDelete       = "CmdServiceDelete"
	CmdServiceLog          = "CmdServiceLog"
	CmdBucket              = "CmdBucket"
	CmdBucketZerops        = "CmdBucketZerops"
	CmdBucketS3            = ""
	CmdBucketCreate        = "CmdBucketCreate"
	CmdBucketDelete        = "CmdBucketDelete"

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
	QuietModeFlag         = "QuietModeFlag"
	TerminalFlag          = "TerminalFlag"
	LogFilePathFlag       = "LogFilePathFlag"
	ConfirmFlag           = "ConfirmFlag"
	ServiceIdFlag         = "ServiceIdFlag"
	ProjectIdFlag         = "ProjectIdFlag"

	// prompt
	PromptEnterZeropsServiceName = "PromptEnterZeropsServiceName"
	PromptName                   = "PromptName"
	PromptInvalidInput           = "PromptInvalidInput"
	PromptInvalidHostname        = "PromptInvalidHostname"

	// process
	ProcessInvalidState        = "ProcessInvalidState"
	ProcessInvalidStateProcess = "ProcessInvalidStateProcess"
	ProcessStart               = "ProcessStart"
	ProcessEnd                 = "ProcessEnd"

	// archiveClient
	ArchClientWorkingDirectory  = "ArchClientWorkingDirectory"
	ArchClientMaxOneTilde       = "ArchClientMaxOneTilde"
	ArchClientPackingDirectory  = "ArchClientPackingDirectory"
	ArchClientPackingFile       = "ArchClientPackingFile"
	ArchClientFileAlreadyExists = "ArchClientFileAlreadyExists"

	// login
	LoginSuccess        = "LoginSuccess"
	LoginIncorrectToken = "LoginIncorrectToken"
	RegionUrl           = "RegionUrl"

	// region
	RegionNotFound = "RegionNotFound"

	// client ID
	MultipleClientIds  = "MultipleClientIds"
	AvailableClientIds = "AvailableClientIds"
	MissingClientId    = "MissingClientId"

	// import
	YamlCheck             = "YamlCheck"
	ImportYamlOk          = "ImportYamlOk"
	ImportYamlEmpty       = "ImportYamlEmpty"
	ImportYamlTooLarge    = "ImportYamlTooLarge"
	ImportYamlFound       = "ImportYamlFound"
	ImportYamlNotFound    = "ImportYamlNotFound"
	ImportYamlCorrupted   = "ImportYamlCorrupted"
	ServiceCount          = "ServiceCount"
	QueuedProcesses       = "QueuedProcesses"
	CoreServices          = "CoreServices"
	ReadyToImportServices = "ReadyToImportServices"

	// delete cmd
	DeleteCanceledByUser = "DeleteCanceledByUser"

	// project + service
	ProjectWrongId               = "ProjectWrongId"
	ProjectsWithSameName         = "ProjectsWithSameName"
	AvailableProjectIds          = "AvailableProjectIds"
	ProjectNameOrIdEmpty         = "ProjectNameOrIdEmpty"
	ProjectDeleteConfirm         = "ProjectDeleteConfirm"
	ServiceNameIsEmpty           = "ServiceNameIsEmpty"
	ServiceDeleteConfirm         = "ServiceDeleteConfirm"
	ProcessInit                  = "ProcessInit"
	ProjectStarting              = "ProjectStarting"
	ProjectStarted               = "ProjectStarted"
	ProjectStopping              = "ProjectStopping"
	ProjectStopped               = "ProjectStopped"
	ProjectDeleting              = "ProjectDeleting"
	ProjectDeleted               = "ProjectDeleted"
	ServiceStarting              = "ServiceStarting"
	ServiceStarted               = "ServiceStarted"
	ServiceStopping              = "ServiceStopping"
	ServiceStopped               = "ServiceStopped"
	ServiceDeleting              = "ServiceDeleting"
	ServiceDeleted               = "ServiceDeleted"
	ProjectImported              = "ProjectImported"
	ServiceImported              = "ServiceImported"
	LogLimitInvalid              = "LogLimitInvalid"
	LogMinSeverityInvalid        = "LogMinSeverityInvalid"
	LogMinSeverityStringLimitErr = "LogMinSeverityStringLimitErr"
	LogMinSeverityNumLimitErr    = "LogMinSeverityNumLimitErr"
	LogFormatInvalid             = "LogFormatInvalid"
	LogFormatTemplateMismatch    = "LogFormatTemplateMismatch"
	LogFormatStreamMismatch      = "LogFormatStreamMismatch"
	LogServiceNameInvalid        = "LogServiceNameInvalid"
	LogFormatTemplateInvalid     = "LogFormatTemplateInvalid"
	LogFormatTemplateNoSpace     = "LogFormatTemplateNoSpace"
	LogSuffixInvalid             = "LogSuffixInvalid"
	LogRuntimeOnly               = "LogRuntimeOnly"
	LogNoContainerFound          = "LogNoContainerFound"
	LogTooFewContainers          = "LogTooFewContainers"
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
	BuildDeployServiceStatus         = "BuildDeployServiceStatus"
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

	// S3
	BucketGenericXAmzAcl              = "BucketGenericXAmzAcl"
	BucketGenericXAmzAclInvalid       = "BucketGenericXAmzAclInvalid"
	BucketGenericOnlyForObjectStorage = "BucketGenericOnlyForObjectStorage"
	BucketGenericBucketNamePrefixed   = "BucketGenericBucketNamePrefixed"

	BucketCreated                 = "BucketCreated"
	BucketCreateCreatingDirect    = "BucketCreateCreatingDirect"
	BucketCreateCreatingZeropsApi = "BucketCreateCreatingZeropsApi"

	BucketDeleteConfirm           = "BucketDeleteConfirm"
	BucketDeleted                 = "BucketDeleted"
	BucketDeleteDeletingDirect    = "BucketDeleteDeletingDirect"
	BucketDeleteDeletingZeropsApi = "BucketDeleteDeletingZeropsApi"

	BucketS3AccessKeyId         = "AccessKeyId"
	BucketS3SecretAccessKey     = "SecretAccessKey"
	BucketS3FlagBothMandatory   = "FlagBothMandatory"
	BucketS3EnvBothMandatory    = "EnvBothMandatory"
	BucketS3RequestFailed       = "RequestFailed"
	BucketS3BucketAlreadyExists = "BucketAlreadyExists"

	// Status info
	StatusInfoCliDataFilePath = "StatusInfoCliDataFilePath"
	StatusInfoLogFilePath     = "StatusInfoLogFilePath"

	// Logger
	LoggerUnableToOpenLogFileWarning = "LoggerUnableToOpenLogFileWarning"

	// generic
	UnauthenticatedUser = "UnauthenticatedUser"

	HintChangeRegion = "HintChangeRegion"

	// UX helpers
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

	// Global
	SelectedProject       = "SelectedProject"
	SelectedService       = "SelectedService"
	ScopedProject         = "ScopedProject"
	ScopedProjectNotFound = "ScopedProjectNotFound"
	ScopedServiceNotFound = "ScopedServiceNotFound"

	ProjectIdInvalidFormat = "ProjectIdInvalidFormat"
	ProjectNotFound        = "ProjectNotFound"

	ServiceIdInvalidFormat = "ServiceIdInvalidFormat"
	ServiceNotFound        = "ServiceNotFound"

	DestructiveOperationConfirmationFailed = "DestructiveOperationConfirmationFailed"
)
