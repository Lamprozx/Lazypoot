package app

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"lazypoot/plugins"
	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

type DeleteStartedMsg struct{}
type DeleteFinishedMsg struct{}
type DeleteErrorMsg struct{ Err string }
type DeleteProgressMsg struct {
	Level	types.LogLevel
	Message	string
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case FontCheckMsg:
		if !msg.NerdFontFound {
			m.showFontWarning = true
			m.fontWarnCursor = 0
		}
		return m, nil

	case InstalledDistrosLoadedMsg:
		m.installedDistros = msg.Distros
		return m, nil

	case DeleteStartedMsg:
		m.nav.Push(types.ScreenInstalling)
		return m, tea.Batch(waitForDeleteLog(m.logChan), tickAfterDelay())

	case DeleteProgressMsg:
		m.logs = append(m.logs, types.LogEntry{Level: msg.Level, Message: msg.Message})
		m.logScroll = len(m.logs)
		if m.logChan != nil {
			return m, tea.Batch(waitForDeleteLog(m.logChan), tickAfterDelay())
		}
		return m, nil

	case DeleteFinishedMsg:
		m.logChan = nil
		m.isDeleting = false
		if m.deleteCancelFn != nil {
			m.deleteCancelFn()
			m.deleteCancelFn = nil
		}
		m.nav.Reset()
		return m, loadInstalledDistrosCmd()

	case DeleteErrorMsg:
		m.logChan = nil
		m.isDeleting = false
		if m.deleteCancelFn != nil {
			m.deleteCancelFn()
			m.deleteCancelFn = nil
		}
		m.showError = true
		m.errorMsg = "Delete failed: " + msg.Err
		m.nav.Reset()
		return m, nil

	case FontInstallStartedMsg:
		return m, nil

	case FontInstallProgressMsg:
		m.fontInstallLogs = append(m.fontInstallLogs, types.LogEntry{Level: msg.Level, Message: msg.Line})
		m.fontInstallStep = msg.Step
		if m.fontInstallChan != nil {
			return m, tea.Batch(waitForFontLog(m.fontInstallChan), tickAfterDelay())
		}
		return m, nil

	case FontInstallDoneMsg:
		m.fontInstallChan = nil
		m.nav.Pop()
		m.nav.Push(types.ScreenFontSuccess)
		return m, nil

	case FontInstallErrorMsg:
		m.fontInstallChan = nil
		m.fontInstallLogs = append(m.fontInstallLogs, types.LogEntry{Level: types.LogError, Message: msg.Err})
		return m, nil

	case DoctorChecksMsg:
		m.doctorLoading = false
		m.doctorChecks = msg.Results
		return m, nil

	case SysInfoMsg:
		m.sysInfoLoading = false
		m.sysInfo = msg.Data
		return m, nil

	case SpinnerTickMsg:
		m.spinnerTick++
		if m.doctorLoading || m.sysInfoLoading || m.logChan != nil || m.fontInstallChan != nil {
			return m, tickAfterDelay()
		}
		return m, nil

	case InstallStartedMsg:
		m.nav.Push(types.ScreenInstalling)
		return m, tea.Batch(waitForLog(m.logChan), tickAfterDelay())

	case InstallProgressPctMsg:
		m.installPct = msg.Percent
		if m.logChan != nil {
			return m, tea.Batch(waitForLog(m.logChan), tickAfterDelay())
		}
		return m, nil

	case InstallProgressMsg:
		m.logs = append(m.logs, types.LogEntry{Level: msg.Level, Message: msg.Line})
		m.logScroll = len(m.logs)
		if m.logChan != nil {
			return m, tea.Batch(waitForLog(m.logChan), tickAfterDelay())
		}
		return m, nil

	case InstallFinishedMsg:
		_, _ = writeInstallLog(m.logs, m.config.Distro)
		if m.installCancelFn != nil {
			m.installCancelFn()
			m.installCancelFn = nil
		}
		m.logChan = nil
		m.config.VNCPassword = msg.VNCPassword
		m.nav.Reset()
		m.nav.Push(types.ScreenInstallSuccess)
		return m, nil

	case InstallErrorMsg:
		_, _ = writeInstallLog(m.logs, m.config.Distro)
		if m.installCancelFn != nil {
			m.installCancelFn()
			m.installCancelFn = nil
		}
		m.logChan = nil
		m.installErrMsg = msg.Err
		m.nav.Reset()
		m.nav.Push(types.ScreenInstallError)
		return m, nil
	}

	keyMsg, isKey := msg.(tea.KeyMsg)
	if !isKey {
		return m, nil
	}

	if m.showFontWarning {
		return handleFontWarning(m, keyMsg)
	}
	if m.showError {
		m.showError = false
		m.errorMsg = ""
		return m, nil
	}

	switch m.nav.Current() {
	case types.ScreenMainMenu:
		return handleMainMenu(m, keyMsg)
	case types.ScreenSelectDistro:
		return handleSelectDistro(m, keyMsg)
	case types.ScreenSetup:
		return handleSetup(m, keyMsg)
	case types.ScreenSetupDE:
		return handleSubList(m, keyMsg, deOptions, deIDs, func(m *Model, id string) {
			m.config.DE = id
			if id == "cli" {
				m.config.Display = ""
			}
		})
	case types.ScreenSetupDisplay:
		return handleSubList(m, keyMsg, displayOptions, displayIDs, func(m *Model, id string) {
			m.config.Display = id
		})
	case types.ScreenSetupDriver:
		return handleDriverScreen(m, keyMsg)
	case types.ScreenSetupProfileMenu:
		return handleProfileMenu(m, keyMsg)
	case types.ScreenSetupHostname:
		return handleNameInput(m, keyMsg, func(m *Model, val string) {
			m.config.Hostname = strings.TrimSpace(val)
		})
	case types.ScreenSetupAddUser:
		return handleAddUserName(m, keyMsg)
	case types.ScreenSetupAddUserPass:
		return handleAddUserPass(m, keyMsg)
	case types.ScreenSetupDeleteUser:
		return handleDeleteUser(m, keyMsg)
	case types.ScreenSetupPackages:
		return handleTextInput(m, keyMsg, func(m *Model, val string) {
			if val == "" {
				m.config.Packages = nil
				return
			}
			pkgs := strings.Fields(val)
			if err := plugins.ValidatePackages(pkgs); err != nil {
				m.showError = true
				m.errorMsg = err.Error()
				return
			}
			m.config.Packages = pkgs
		})
	case types.ScreenSetupTimezone:
		return handleSubList(m, keyMsg, tzOptions, tzOptions, func(m *Model, id string) {
			m.config.Timezone = id
		})
	case types.ScreenConfirmInstall:
		return handleConfirmInstall(m, keyMsg)
	case types.ScreenInstalling:
		return handleInstalling(m, keyMsg)
	case types.ScreenInstallSuccess, types.ScreenInstallError:
		return handlePostInstall(m, keyMsg)
	case types.ScreenNoDistro, types.ScreenDeleteList:
		return handleDelete(m, keyMsg)
	case types.ScreenDoctor:
		return handleDoctor(m, keyMsg)
	case types.ScreenDoctorLogs:
		return handleDoctorSub(m, keyMsg)
	case types.ScreenDoctorSysInfo:
		return handleDoctorSub(m, keyMsg)
	case types.ScreenFontInstalling:
		return handleFontInstalling(m, keyMsg)
	case types.ScreenFontSuccess:
		return handleFontSuccess(m, keyMsg)
	}
	return m, nil
}

func handleDelete(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch m.nav.Current() {
	case types.ScreenNoDistro:
		if key.String() == "esc" {
			m.nav.Pop()
		}
	case types.ScreenDeleteList:
		n := len(m.installedDistros)
		switch key.String() {
		case "j", "down":
			m.deleteCursor = clampCursor(m.deleteCursor+1, n)
		case "k", "up":
			m.deleteCursor = clampCursor(m.deleteCursor-1, n)
		case " ":
			if m.deleteCursor < n {
				m.installedDistros[m.deleteCursor].Selected = !m.installedDistros[m.deleteCursor].Selected
			}
		case "enter":
			var ids []string
			for _, d := range m.installedDistros {
				if d.Selected {
					ids = append(ids, d.ID)
				}
			}
			if len(ids) == 0 {
				m.nav.Pop()
				return m, nil
			}
			ctx, cancel := context.WithCancel(context.Background())
			m.deleteCancelFn = cancel
			m.isDeleting = true
			m.logs = nil
			m.logScroll = 0
			m.logTotal = len(ids) * 5
			ch := runDeleteDistroCmd(ctx, ids)
			m.logChan = ch
			return m, func() tea.Msg { return DeleteStartedMsg{} }
		case "esc":
			m.nav.Pop()
		}
	}
	return m, nil
}

func runDeleteDistroCmd(ctx context.Context, ids []string) chan tea.Msg {
	ch := make(chan tea.Msg, 256)
	go func() {
		defer close(ch)
		for _, id := range ids {
			select {
			case <-ctx.Done():
				ch <- DeleteProgressMsg{Level: types.LogWarn, Message: "Delete cancelled by user"}
				return
			default:
			}

			ch <- DeleteProgressMsg{Level: types.LogStep, Message: fmt.Sprintf("Removing distro: %s...", id)}

			cmd := exec.CommandContext(ctx, "proot-distro", "remove", id)
			pr, pw := io.Pipe()
			cmd.Stdout = pw
			cmd.Stderr = pw

			if err := cmd.Start(); err != nil {
				pw.Close()
				ch <- DeleteProgressMsg{Level: types.LogError, Message: fmt.Sprintf("Failed to start removal of %s: %v", id, err)}
				continue
			}

			var waitErr error
			done := make(chan struct{})
			go func() {
				waitErr = cmd.Wait()
				pw.Close()
				close(done)
			}()

			scanner := bufio.NewScanner(pr)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					ch <- DeleteProgressMsg{Level: types.LogInfo, Message: "  " + line}
				}
			}
			<-done

			if ctx.Err() != nil {
				ch <- DeleteProgressMsg{Level: types.LogWarn, Message: "Delete cancelled by user"}
			} else if waitErr != nil {
				ch <- DeleteProgressMsg{Level: types.LogError, Message: fmt.Sprintf("Error removing %s: %v", id, waitErr)}
			} else {
				home := os.Getenv("HOME")
				if home != "" {
					if err := RemovePortalHooks(home, id); err != nil {
						ch <- DeleteProgressMsg{Level: types.LogWarn, Message: fmt.Sprintf("Hook cleanup: %v", err)}
					} else {
						ch <- DeleteProgressMsg{Level: types.LogInfo, Message: "  Removed shell hooks for " + id}
					}
				}
				ch <- DeleteProgressMsg{Level: types.LogOK, Message: fmt.Sprintf("Distro %s removed successfully.", id)}
			}
		}
	}()
	return ch
}

func waitForDeleteLog(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return DeleteFinishedMsg{}
		}
		return msg
	}
}

func handleFontWarning(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "j", "down", "l", "right":
		m.fontWarnCursor = 1
	case "k", "up", "h", "left":
		m.fontWarnCursor = 0
	case "enter":
		m.showFontWarning = false
		if m.fontWarnCursor == 1 {
			ch, cmd := runFontInstallCmd()
			m.fontInstallChan = ch
			m.fontInstallLogs = nil
			m.fontInstallStep = 0
			m.nav.Push(types.ScreenFontInstalling)
			return m, cmd
		}
	case "esc", "q":
		m.showFontWarning = false
	}
	return m, nil
}

func handleFontInstalling(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	if key.String() == "esc" && m.fontInstallChan == nil {
		m.nav.Pop()
	}
	return m, nil
}

func handleFontSuccess(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	if key.String() == "esc" || key.String() == "enter" || key.String() == "q" {
		m.nav.Reset()
	}
	return m, nil
}

func handleMainMenu(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	const itemCount = 4
	switch key.String() {
	case "j", "down":
		m.mainCursor = clampCursor(m.mainCursor+1, itemCount)
	case "k", "up":
		m.mainCursor = clampCursor(m.mainCursor-1, itemCount)
	case "enter":
		switch m.mainCursor {
		case 0:
			m.nav.Push(types.ScreenSelectDistro)
		case 1:
			if len(m.installedDistros) == 0 {
				m.nav.Push(types.ScreenNoDistro)
			} else {
				m.nav.Push(types.ScreenDeleteList)
			}
		case 2:
			m.doctorLoading = true
			m.doctorChecks = nil
			m.nav.Push(types.ScreenDoctor)
			return m, tea.Batch(runDoctorChecksCmd(), tickAfterDelay())
		case 3:
			return m, tea.Quit
		}
	case "i":
		m.nav.Push(types.ScreenSelectDistro)
	case "d":
		if len(m.installedDistros) == 0 {
			m.nav.Push(types.ScreenNoDistro)
		} else {
			m.nav.Push(types.ScreenDeleteList)
		}
	case "p":
		m.doctorLoading = true
		m.doctorChecks = nil
		m.nav.Push(types.ScreenDoctor)
		return m, tea.Batch(runDoctorChecksCmd(), tickAfterDelay())
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func handleSelectDistro(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	n := len(m.distroItems)
	switch key.String() {
	case "j", "down":
		m.distroCursor = clampCursor(m.distroCursor+1, n)
	case "k", "up":
		m.distroCursor = clampCursor(m.distroCursor-1, n)
	case " ":
		for i := range m.distroItems {
			m.distroItems[i].Selected = i == m.distroCursor && !m.distroItems[i].Selected
		}
		id, name := m.selectedDistroID()
		m.config.DistroID = id
		m.config.Distro = name
	case "enter":
		id, name := m.selectedDistroID()
		if id == "" {
			m.showError = true
			m.errorMsg = "Please select a distro first (press Space to toggle)"
			return m, nil
		}
		m.config.DistroID = id
		m.config.Distro = name
		m.nav.Push(types.ScreenSetup)
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func handleSetup(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	const itemCount = 7
	switch key.String() {
	case "j", "down":
		m.setupCursor = clampCursor(m.setupCursor+1, itemCount)
	case "k", "up":
		m.setupCursor = clampCursor(m.setupCursor-1, itemCount)
	case "enter":
		switch m.setupCursor {
		case 0:
			m.subCursor = 0
			m.nav.Push(types.ScreenSetupDE)
		case 1:
			if m.config.DE == "cli" {
				m.showError = true
				m.errorMsg = "Display server is not needed for CLI Only mode"
				return m, nil
			}
			m.subCursor = 0
			m.nav.Push(types.ScreenSetupDisplay)
		case 2:
			if m.config.DE == "cli" {
				m.showError = true
				m.errorMsg = "GPU driver is not needed for CLI Only mode"
				return m, nil
			}
			m.subCursor = 0
			m.nav.Push(types.ScreenSetupDriver)
		case 3:
			m.profileCursor = 0
			m.nav.Push(types.ScreenSetupProfileMenu)
		case 4:
			m.inputBuffer = strings.Join(m.config.Packages, " ")
			m.nav.Push(types.ScreenSetupPackages)
		case 5:
			m.subCursor = 0
			for i, tz := range tzOptions {
				if tz == m.config.Timezone {
					m.subCursor = i
					break
				}
			}
			m.nav.Push(types.ScreenSetupTimezone)
		case 6:
			if errMsg := m.validate(); errMsg != "" {
				m.showError = true
				m.errorMsg = errMsg
				return m, nil
			}
			m.nav.Push(types.ScreenConfirmInstall)
		}
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func handleDriverScreen(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "j", "down":
		m.subCursor = clampCursor(m.subCursor+1, 2)
	case "k", "up":
		m.subCursor = clampCursor(m.subCursor-1, 2)
	case "enter":
		m.config.InstallVirGL = m.subCursor == 0
		m.nav.Pop()
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func handleProfileMenu(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	const itemCount = 3
	switch key.String() {
	case "j", "down":
		m.profileCursor = clampCursor(m.profileCursor+1, itemCount)
	case "k", "up":
		m.profileCursor = clampCursor(m.profileCursor-1, itemCount)
	case "enter":
		switch m.profileCursor {
		case 0:
			m.inputBuffer = m.config.Hostname
			m.nav.Push(types.ScreenSetupHostname)
		case 1:
			m.addUserNameBuf = ""
			m.inputBuffer = ""
			m.nav.Push(types.ScreenSetupAddUser)
		case 2:
			if len(m.config.Users) == 0 {
				m.showError = true
				m.errorMsg = "No users added yet"
				return m, nil
			}
			m.subCursor = 0
			m.nav.Push(types.ScreenSetupDeleteUser)
		}
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func validateNameAllowlist(s string) string {
	if s == "" {
		return "Name cannot be empty"
	}
	if !types.SafeNameRe.MatchString(s) {
		return "Only lowercase letters a-z, digits 0-9, and hyphens allowed.\nMust start with a letter. Max 31 characters."
	}
	return ""
}

func handleNameInput(m Model, key tea.KeyMsg, apply func(*Model, string)) (Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		if errMsg := validateNameAllowlist(m.inputBuffer); errMsg != "" {
			m.showError = true
			m.errorMsg = errMsg
			return m, nil
		}
		apply(&m, m.inputBuffer)
		m.inputBuffer = ""
		m.nav.Pop()
	case "esc":
		m.inputBuffer = ""
		m.nav.Pop()
	case "backspace", "ctrl+h":
		if len(m.inputBuffer) > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:len(runes)-1])
		}
	default:
		if len(key.String()) == 1 && len([]rune(m.inputBuffer)) < types.MaxInputLen {
			m.inputBuffer += key.String()
		}
	}
	return m, nil
}

func handlePasswordInput(m Model, key tea.KeyMsg, apply func(*Model, string)) (Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		if strings.TrimSpace(m.inputBuffer) == "" {
			m.showError = true
			m.errorMsg = "Password cannot be empty"
			return m, nil
		}
		if len([]rune(m.inputBuffer)) < 6 {
			m.showError = true
			m.errorMsg = "Password must be at least 6 characters"
			return m, nil
		}
		apply(&m, m.inputBuffer)
		m.inputBuffer = ""
		m.nav.Pop()
	case "esc":
		m.inputBuffer = ""
		m.nav.Pop()
	case "backspace", "ctrl+h":
		if len(m.inputBuffer) > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:len(runes)-1])
		}
	default:
		if len(key.String()) == 1 && len([]rune(m.inputBuffer)) < types.MaxInputLen {
			m.inputBuffer += key.String()
		}
	}
	return m, nil
}

func handleAddUserName(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		name := strings.TrimSpace(m.inputBuffer)
		if errMsg := validateNameAllowlist(name); errMsg != "" {
			m.showError = true
			m.errorMsg = errMsg
			return m, nil
		}
		m.addUserNameBuf = name
		m.inputBuffer = ""
		m.nav.Pop()
		m.nav.Push(types.ScreenSetupAddUserPass)
	case "esc":
		m.inputBuffer = ""
		m.nav.Pop()
	case "backspace", "ctrl+h":
		if len(m.inputBuffer) > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:len(runes)-1])
		}
	default:
		if len(key.String()) == 1 && len([]rune(m.inputBuffer)) < types.MaxInputLen {
			m.inputBuffer += key.String()
		}
	}
	return m, nil
}

func handleAddUserPass(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		if strings.TrimSpace(m.inputBuffer) == "" {
			m.showError = true
			m.errorMsg = "Password cannot be empty"
			return m, nil
		}
		if len([]rune(m.inputBuffer)) < 6 {
			m.showError = true
			m.errorMsg = "Password must be at least 6 characters"
			return m, nil
		}
		m.config.Users = append(m.config.Users, types.UserAccount{
			Username:	m.addUserNameBuf,
			Password:	m.inputBuffer,
		})
		m.addUserNameBuf = ""
		m.inputBuffer = ""
		m.nav.Pop()
	case "esc":
		m.addUserNameBuf = ""
		m.inputBuffer = ""
		m.nav.Pop()
	case "backspace", "ctrl+h":
		if len(m.inputBuffer) > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:len(runes)-1])
		}
	default:
		if len(key.String()) == 1 && len([]rune(m.inputBuffer)) < types.MaxInputLen {
			m.inputBuffer += key.String()
		}
	}
	return m, nil
}

func handleDeleteUser(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	n := len(m.config.Users)
	switch key.String() {
	case "j", "down":
		m.subCursor = clampCursor(m.subCursor+1, n)
	case "k", "up":
		m.subCursor = clampCursor(m.subCursor-1, n)
	case "enter":
		if m.subCursor < n {
			m.config.Users = append(m.config.Users[:m.subCursor], m.config.Users[m.subCursor+1:]...)
			m.subCursor = clampCursor(m.subCursor, len(m.config.Users))
		}
		m.nav.Pop()
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func handleTextInput(m Model, key tea.KeyMsg, apply func(*Model, string)) (Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		apply(&m, m.inputBuffer)
		m.inputBuffer = ""
		m.nav.Pop()
	case "esc":
		m.inputBuffer = ""
		m.nav.Pop()
	case "backspace", "ctrl+h":
		if len(m.inputBuffer) > 0 {
			runes := []rune(m.inputBuffer)
			m.inputBuffer = string(runes[:len(runes)-1])
		}
	default:
		if len(key.String()) == 1 && len([]rune(m.inputBuffer)) < types.MaxInputLen {
			m.inputBuffer += key.String()
		}
	}
	return m, nil
}

func handleSubList(m Model, key tea.KeyMsg, options []string, ids []string, apply func(*Model, string)) (Model, tea.Cmd) {
	n := len(options)
	switch key.String() {
	case "j", "down":
		m.subCursor = clampCursor(m.subCursor+1, n)
	case "k", "up":
		m.subCursor = clampCursor(m.subCursor-1, n)
	case "enter":
		if m.subCursor < len(ids) {
			apply(&m, ids[m.subCursor])
		}
		m.nav.Pop()
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func handleConfirmInstall(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		if m.config.Display == "vnc" && m.config.VNCPassword == "" {
			pw, err := generateVNCPassword()
			if err != nil {
				m.showError = true
				m.errorMsg = "Failed to generate VNC password: " + err.Error()
				return m, nil
			}
			m.config.VNCPassword = pw
		}
		m.logTotal = 80
		ctx, cancel := context.WithCancel(context.Background())
		m.installCancelFn = cancel
		ch, cmd := ExecuteInstall(ctx, m)
		m.logChan = ch
		m.logs = nil
		m.logScroll = 0
		return m, cmd
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func handleInstalling(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "j", "down":
		if m.logScroll < len(m.logs) {
			m.logScroll++
		}
	case "k", "up":
		if m.logScroll > 0 {
			m.logScroll--
		}
	case "ctrl+c":
		if m.isDeleting && m.deleteCancelFn != nil {
			m.deleteCancelFn()
			m.deleteCancelFn = nil
			return m, nil
		}
		if m.installCancelFn != nil {
			m.installCancelFn()
			m.installCancelFn = nil
		}
		if m.logChan != nil {
			ch := m.logChan
			m.logChan = nil
			go func() {
				for range ch {
				}
			}()
		}
		m.nav.Reset()
		return m, loadInstalledDistrosCmd()
	}
	return m, nil
}

func handlePostInstall(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	switch key.String() {
	case "enter", "q", "esc":
		m.nav.Reset()
	}
	return m, nil
}

func handleDoctor(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	const itemCount = 3
	switch key.String() {
	case "j", "down":
		m.doctorCursor = clampCursor(m.doctorCursor+1, itemCount)
	case "k", "up":
		m.doctorCursor = clampCursor(m.doctorCursor-1, itemCount)
	case "enter":
		switch m.doctorCursor {
		case 0:
			m.nav.Push(types.ScreenDoctorLogs)
		case 1:
			m.sysInfoLoading = true
			m.sysInfo = types.SysInfoData{}
			m.nav.Push(types.ScreenDoctorSysInfo)
			return m, tea.Batch(loadSysInfoCmd(), tickAfterDelay())
		case 2:
			m.nav.Pop()
		}
	case "esc":
		m.nav.Pop()
	}
	return m, nil
}

func handleDoctorSub(m Model, key tea.KeyMsg) (Model, tea.Cmd) {
	if key.String() == "esc" {
		m.sysInfoLoading = false
		m.nav.Pop()
	}
	return m, nil
}

func tickAfterDelay() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(_ time.Time) tea.Msg {
		return SpinnerTickMsg{}
	})
}
