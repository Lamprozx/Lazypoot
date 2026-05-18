package app

import (
	"context"
	"strings"

	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

var deOptions = []string{
	"XFCE", "LXQt", "Openbox", "GNOME", "KDE Plasma", "MATE", "Cinnamon", "CLI Only",
}
var deIcons = []string{"󰧨", "󰧨", "󰧨", "󰨇", "󰚗", "󰧨", "󰧨", "\uf489"}
var deIDs = []string{"xfce", "lxqt", "openbox", "gnome", "kde", "mate", "cinnamon", "cli"}

var displayOptions = []string{"X11 (Termux X11)", "VNC (TigerVNC)"}
var displayIcons = []string{"󰹑", "󰕈"}
var displayIDs = []string{"x11", "vnc"}

var driverOptions = []string{"Install GPU acceleration (VirGL / OpenGL)", "Disable (software rendering only)"}
var driverIcons = []string{"󰾲", "󰒇"}

var tzOptions = []string{
	"Asia/Jakarta", "Asia/Singapore", "Asia/Tokyo", "Asia/Shanghai",
	"Asia/Kolkata", "Europe/London", "Europe/Berlin", "America/New_York",
	"America/Los_Angeles", "UTC",
}
var tzIcons = []string{"󰌎", "󰌎", "󰌎", "󰌎", "󰌎", "󰌎", "󰌎", "󰌎", "󰌎", "󰌎"}

var profileMenuLabels = []string{"Set Hostname", "Add User", "Remove User"}
var profileMenuIcons = []string{"󰟀", "󰀄", "󰀅"}

type Model struct {
	width	int
	height	int
	nav	NavStack

	config	types.InstallConfig

	distroItems		[]types.DistroItem
	installedDistros	[]types.InstalledDistro

	mainCursor	int
	distroCursor	int
	setupCursor	int
	subCursor	int
	profileCursor	int
	deleteCursor	int
	doctorCursor	int

	fontWarnCursor	int
	showFontWarning	bool
	fontInstallLogs	[]types.LogEntry
	fontInstallChan	chan tea.Msg
	fontInstallStep	int

	doctorLoading	bool
	doctorChecks	[]types.CheckResult
	spinnerTick	int

	sysInfoLoading	bool
	sysInfo		types.SysInfoData

	logs		[]types.LogEntry
	logScroll	int
	logTotal	int

	isDeleting	bool
	deleteCancelFn	context.CancelFunc
	installCancelFn	context.CancelFunc

	showError	bool
	errorMsg	string
	installErrMsg	string

	inputBuffer	string
	addUserNameBuf	string

	logChan		chan tea.Msg
	installPct	int
}

func NewModel() Model {
	return Model{
		width:		80,
		height:		24,
		nav:		newNavStack(),
		config:		types.InstallConfig{Timezone: "Asia/Jakarta"},
		distroItems:	defaultDistroItems(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(checkFontCmd(), loadInstalledDistrosCmd())
}

func defaultDistroItems() []types.DistroItem {
	return []types.DistroItem{
		{Name: "Adélie Linux", ID: "adelie", Icon: "\uf300"},
		{Name: "AlmaLinux", ID: "almalinux", Icon: "\uf31d"},
		{Name: "Alpine Linux", ID: "alpine", Icon: "\uf300"},
		{Name: "Arch Linux", ID: "archlinux", Icon: "\uf303"},
		{Name: "Artix Linux", ID: "artix", Icon: "\uf303"},
		{Name: "Chimera Linux", ID: "chimera", Icon: "\uf17c"},
		{Name: "Debian", ID: "debian", Icon: "\uf306"},
		{Name: "Deepin", ID: "deepin", Icon: "\uf17c"},
		{Name: "Fedora", ID: "fedora", Icon: "\uf30a"},
		{Name: "Manjaro", ID: "manjaro", Icon: "\uf312"},
		{Name: "OpenSUSE", ID: "opensuse", Icon: "\uf314"},
		{Name: "Oracle Linux", ID: "oracle", Icon: "\uf17c"},
		{Name: "Pardus", ID: "pardus", Icon: "\uf17c"},
		{Name: "Rocky Linux", ID: "rockylinux", Icon: "\uf31d"},
		{Name: "Trisquel GNU/Linux", ID: "trisquel", Icon: "\uf17c"},
		{Name: "Ubuntu (25.10)", ID: "ubuntu", Icon: "\uf31b"},
		{Name: "Void Linux", ID: "void", Icon: "\uf317"},
	}
}

func (m *Model) selectedDistroID() (string, string) {
	for _, d := range m.distroItems {
		if d.Selected {
			return d.ID, d.Name
		}
	}
	return "", ""
}

func (m *Model) validate() string {
	var errs []string
	if m.config.DistroID == "" {
		errs = append(errs, "No distro selected")
	}
	if m.config.DE == "" {
		errs = append(errs, "No desktop environment selected")
	}
	if m.config.DE != "cli" && m.config.Display == "" {
		errs = append(errs, "No display server selected")
	}
	if strings.TrimSpace(m.config.Hostname) == "" {
		errs = append(errs, "Hostname cannot be empty")
	}
	if len(errs) == 0 {
		return ""
	}
	return strings.Join(errs, "\n  ")
}

func (m *Model) logStrings() []string {
	lines := readLatestLog()
	if len(lines) == 0 && len(m.logs) > 0 {
		out := make([]string, len(m.logs))
		for i, e := range m.logs {
			out[i] = "[" + string(e.Level) + "] " + e.Message
		}
		return out
	}
	return lines
}
