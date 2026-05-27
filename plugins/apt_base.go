package plugins

import "fmt"

type aptBase struct {
	name	string
	id	string
	icon	string
}

func (a aptBase) Name() string			{ return a.name }
func (a aptBase) ID() string			{ return a.id }
func (a aptBase) Icon() string			{ return a.icon }
func (a aptBase) PackageManager() string	{ return "apt" }

func (a aptBase) UpdateSystem() []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get update -y",
		"DEBIAN_FRONTEND=noninteractive apt-get upgrade -y",
	}
}

func (a aptBase) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", joinPkgs(pkgs)),
	}
}

var aptDEPkgs = map[string]string{
	"xfce":		"xfce4 xfce4-goodies",
	"lxqt":		"lxqt",
	"openbox":	"openbox obconf",
	"gnome":	"gnome",
	"kde":		"kde-standard",
	"awesome":	"awesome",
	"i3":		"i3",
}

func (a aptBase) InstallDesktop(de string) []string {
	pkg := aptDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return a.InstallPackages([]string{pkg})
}

var aptDisplayPkgs = map[string]string{
	"vnc":	"tigervnc-standalone-server",
	"x11":	"xorg x11-xserver-utils",
}

func (a aptBase) InstallDisplay(display string) []string {
	pkg := aptDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return a.InstallPackages([]string{pkg})
}

func (a aptBase) SetupUser(hostname string) []string {
	return []string{
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (a aptBase) SetupOpenGL() []string {
	return a.InstallPackages([]string{"mesa-utils", "libgl1-mesa-dri", "mesa-vulkan-drivers", "libgles2-mesa"})
}

func (a aptBase) PostInstall() []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get autoremove -y",
		"apt-get clean",
	}
}
