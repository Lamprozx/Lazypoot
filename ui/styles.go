package ui

import "github.com/charmbracelet/lipgloss"

const (
	ColBg		= "#1a1b26"
	ColBg2		= "#16161e"
	ColBlue		= "#7aa2f7"
	ColCyan		= "#7dcfff"
	ColGreen	= "#9ece6a"
	ColYellow	= "#e0af68"
	ColOrange	= "#ff9e64"
	ColRed		= "#f7768e"
	ColPurple	= "#9d7cd8"
	ColMagenta	= "#bb9af7"
	ColComment	= "#565f89"
	ColFg		= "#c0caf5"
	ColFgDim	= "#a9b1d6"
	ColSelected	= "#2d3f76"
	ColBorder	= "#3b4261"
	ColTitle	= "#7aa2f7"
)

var (
	StyleBox	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColBorder))

	StyleBoxBlue	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColBlue))

	StyleBoxCyan	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColCyan))

	StyleBoxGreen	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColGreen))

	StyleBoxRed	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColRed))

	StyleBoxPurple	= lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColPurple))
)

var (
	StyleTitle	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColBlue)).
			Bold(true)

	StyleSubtitle	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColComment))

	StyleBold	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColFg)).
			Bold(true)

	StyleDim	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColComment))

	StyleGreen	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColGreen)).Bold(true)

	StyleBlue	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColBlue))

	StyleCyan	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColCyan))

	StyleOrange	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColOrange))

	StylePurple	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColPurple))

	StyleMagenta	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColMagenta))

	StyleRed	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColRed))

	StyleYellow	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColYellow))

	StyleKey	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColOrange)).
			Bold(true)

	StyleCursor	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColCyan)).
			Bold(true)

	StyleSelected	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColCyan)).
			Background(lipgloss.Color(ColSelected)).
			Bold(true)

	StyleNormal	= lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColFgDim))
)

var LogStyles = map[string]lipgloss.Style{
	"INFO":		lipgloss.NewStyle().Foreground(lipgloss.Color(ColBlue)),
	"OK":		lipgloss.NewStyle().Foreground(lipgloss.Color(ColGreen)).Bold(true),
	"WARN":		lipgloss.NewStyle().Foreground(lipgloss.Color(ColYellow)),
	"ERROR":	lipgloss.NewStyle().Foreground(lipgloss.Color(ColRed)).Bold(true),
	"STEP":		lipgloss.NewStyle().Foreground(lipgloss.Color(ColMagenta)).Bold(true),
}

var StyleFooter = lipgloss.NewStyle().
	Foreground(lipgloss.Color(ColComment)).
	BorderTop(true).
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color(ColBorder))
