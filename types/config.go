package types

import "regexp"

var SafePkgRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._+-]{0,99}$`)
var SafeNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,30}$`)

const MaxInputLen = 512

type Screen int

const (
	ScreenMainMenu	Screen	= iota
	ScreenSelectDistro
	ScreenSetup
	ScreenSetupGUI
	ScreenSetupDE
	ScreenSetupWM
	ScreenSetupKaliVariant
	ScreenSetupDisplay
	ScreenSetupDriver
	ScreenSetupProfileMenu
	ScreenSetupHostname
	ScreenSetupAddUser
	ScreenSetupAddUserPass
	ScreenSetupDeleteUser
	ScreenSetupPackages
	ScreenSetupTimezone
	ScreenConfirmInstall
	ScreenInstalling
	ScreenInstallError
	ScreenInstallSuccess
	ScreenNoDistro
	ScreenDeleteList
	ScreenDoctor
	ScreenDoctorLogs
	ScreenDoctorSysInfo
	ScreenFontInstalling
	ScreenFontSuccess
)

type DistroItem struct {
	Name		string
	ID		string
	Icon		string
	Selected	bool
}

type UserAccount struct {
	Username	string
	Password	string
}

type InstallConfig struct {
	Distro		string
	DistroID	string
	DE		string
	Display		string
	InstallVirGL	bool
	Hostname	string

	VNCPassword	string
	Users		[]UserAccount
	Packages		[]string
	Timezone	string
	KaliVariant	string
	AvailableStorage	int64
}

type InstalledDistro struct {
	Name		string
	ID		string
	Icon		string
	Selected	bool
}

type LogLevel string

const (
	LogInfo		LogLevel	= "INFO"
	LogOK		LogLevel	= "OK"
	LogWarn		LogLevel	= "WARN"
	LogError	LogLevel	= "ERROR"
	LogStep		LogLevel	= "STEP"
)

type LogEntry struct {
	Level	LogLevel
	Message	string
}

type CheckResult struct {
	Tool	string
	Icon	string
	Found	bool
	Path	string
}

type SysInfoData struct {
	OS		string
	Kernel		string
	Arch		string
	CPU		string
	RAM		string
	Storage		string
	Shell		string
	ProotVer	string
}
