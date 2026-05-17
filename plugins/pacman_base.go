package plugins

import "fmt"

type pacmanBase struct {
	name	string
	id	string
	icon	string
}

func (p pacmanBase) Name() string		{ return p.name }
func (p pacmanBase) ID() string			{ return p.id }
func (p pacmanBase) Icon() string		{ return p.icon }
func (p pacmanBase) PackageManager() string	{ return "pacman" }

func (p pacmanBase) UpdateSystem() []string {
	return []string{"pacman -Syu --noconfirm"}
}

func (p pacmanBase) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("pacman -S --noconfirm --needed %s", joinPkgs(pkgs)),
	}
}

var pacmanDEPkgs = map[string]string{
	"xfce":		"xfce4 xfce4-goodies",
	"lxqt":		"lxqt",
	"openbox":	"openbox obconf",
	"gnome":	"gnome gnome-extra",
	"kde":		"plasma kde-applications",
	"mate":		"mate mate-extra",
	"cinnamon":	"cinnamon",
}

func (p pacmanBase) InstallDesktop(de string) []string {
	pkg := pacmanDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return p.InstallPackages([]string{pkg})
}

var pacmanDisplayPkgs = map[string]string{
	"vnc":	"tigervnc",
	"x11":	"xorg-server xorg-xinit",
}

func (p pacmanBase) InstallDisplay(display string) []string {
	pkg := pacmanDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return p.InstallPackages([]string{pkg})
}

func (p pacmanBase) SetupUser(hostname string) []string {
	return []string{
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (p pacmanBase) SetupOpenGL() []string {
	return p.InstallPackages([]string{"mesa", "mesa-utils", "vulkan-mesa-layers", "lib32-mesa"})
}

func (p pacmanBase) PostInstall() []string {
	return []string{"pacman -Sc --noconfirm"}
}
