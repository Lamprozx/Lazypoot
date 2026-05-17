package app

import (
	"fmt"

	"lazypoot/screens"
	"lazypoot/types"
	"lazypoot/ui"
)

func (m Model) View() string {
	if m.showFontWarning {
		return screens.RenderFontWarnInteractive(m.fontWarnCursor, m.width, m.height)
	}
	if m.showError {
		return screens.RenderErrorPopup(m.errorMsg, m.width, m.height)
	}

	switch m.nav.Current() {
	case types.ScreenMainMenu:
		return screens.RenderMainMenu(m.mainCursor, m.width, m.height)
	case types.ScreenSelectDistro:
		return screens.RenderSelectDistro(m.distroItems, m.distroCursor, m.width, m.height)
	case types.ScreenSetup:
		return screens.RenderSetup(m.config, m.setupCursor, m.width, m.height)
	case types.ScreenSetupDE:
		return screens.RenderSetupSubscreen("  Desktop Environment", deOptions, deIcons, m.subCursor, m.width, m.height)
	case types.ScreenSetupDisplay:
		return screens.RenderSetupSubscreen("󰹑  Display Server", displayOptions, displayIcons, m.subCursor, m.width, m.height)
	case types.ScreenSetupDriver:
		return screens.RenderDriverScreen(driverOptions, driverIcons, m.subCursor, m.width, m.height)
	case types.ScreenSetupProfileMenu:
		return screens.RenderProfileMenu(profileMenuLabels, profileMenuIcons, m.profileCursor, m.config, m.width, m.height)
	case types.ScreenSetupHostname:
		return screens.RenderInputScreen("󰟀  Hostname", "Hostname", m.inputBuffer, false, m.width, m.height)
	case types.ScreenSetupAddUser:
		return screens.RenderInputScreen("󰀄  Add User  ›  Username", "Username", m.inputBuffer, false, m.width, m.height)
	case types.ScreenSetupAddUserPass:
		return screens.RenderInputScreen("󰀄  Add User  ›  Password for "+m.addUserNameBuf, "Password", m.inputBuffer, true, m.width, m.height)
	case types.ScreenSetupDeleteUser:
		return screens.RenderDeleteUserList(m.config.Users, m.subCursor, m.width, m.height)
	case types.ScreenSetupPackages:
		return screens.RenderInputScreen("󰏗  Extra Packages", "Packages (space separated)", m.inputBuffer, false, m.width, m.height)
	case types.ScreenSetupTimezone:
		return screens.RenderSetupSubscreen("󰌎  Timezone", tzOptions, tzIcons, m.subCursor, m.width, m.height)
	case types.ScreenConfirmInstall:
		return screens.RenderConfirmInstall(m.config, m.width, m.height)
	case types.ScreenInstalling:
		label := fmt.Sprintf("Installing %s", m.config.Distro)
		if m.isDeleting {
			label = "Removing distro(s)"
		}
		spinnerChar := spinnerFrame(m.spinnerTick)
		return screens.RenderInstalling(m.logs, m.logScroll, m.logTotal, label, spinnerChar, m.width, m.height, m.installPct)
	case types.ScreenInstallSuccess:
		return screens.RenderInstallSuccess(m.config.Distro, m.config.DistroID, m.config.DE, m.config.Display, m.config.VNCPassword, m.width, m.height)
	case types.ScreenInstallError:
		return screens.RenderInstallError(m.installErrMsg, m.width, m.height)
	case types.ScreenNoDistro:
		return screens.RenderNoDistro(m.width, m.height)
	case types.ScreenDeleteList:
		return screens.RenderDeleteList(m.installedDistros, m.deleteCursor, m.width, m.height)
	case types.ScreenDoctor:
		return screens.RenderDoctor(m.doctorCursor, m.doctorLoading, m.doctorChecks, spinnerFrame(m.spinnerTick), m.width, m.height)
	case types.ScreenDoctorLogs:
		return screens.RenderDoctorLogs(m.logStrings(), m.width, m.height)
	case types.ScreenDoctorSysInfo:
		return screens.RenderDoctorSysInfo(m.sysInfo, m.sysInfoLoading, spinnerFrame(m.spinnerTick), m.width, m.height)
	case types.ScreenFontInstalling:
		return screens.RenderFontInstalling(m.fontInstallLogs, m.fontInstallStep, m.width, m.height)
	case types.ScreenFontSuccess:
		return screens.RenderFontSuccess(m.width, m.height)
	}

	return ui.StyleRed.Render("  Unknown screen — press q to quit")
}
