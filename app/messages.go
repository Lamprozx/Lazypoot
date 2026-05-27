package app

import "lazypoot/types"

type InstallStartedMsg struct{}

type FontInstallStartedMsg struct{}

type InstallProgressMsg struct {
	Line	string
	Level	types.LogLevel
}

type InstallFinishedMsg struct {
	Distro		string
	VNCPassword	string
}

type InstallErrorMsg struct {
	Err string
}

type FontCheckMsg struct {
	NerdFontFound bool
}

type InstalledDistrosLoadedMsg struct {
	Distros []types.InstalledDistro
}

type FontInstallProgressMsg struct {
	Line	string
	Level	types.LogLevel
	Step	int
}

type FontInstallDoneMsg struct{}

type FontInstallErrorMsg struct {
	Err string
}

type DoctorChecksMsg struct {
	Results []types.CheckResult
}

type SysInfoMsg struct {
	Data types.SysInfoData
}

type InstallProgressPctMsg struct {
	Percent int
}

type SpinnerTickMsg struct{}
