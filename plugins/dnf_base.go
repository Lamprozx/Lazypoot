package plugins

import "fmt"

type dnfBase struct {
	name	string
	id	string
	icon	string
}

func (d dnfBase) Name() string			{ return d.name }
func (d dnfBase) ID() string			{ return d.id }
func (d dnfBase) Icon() string			{ return d.icon }
func (d dnfBase) PackageManager() string	{ return "dnf" }

func (d dnfBase) UpdateSystem() []string {
	return []string{"dnf update -y"}
}

func (d dnfBase) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("dnf install -y %s", joinPkgs(pkgs))}
}

var dnfDEPkgs = map[string]string{
	"xfce":		"@xfce-desktop",
	"lxqt":		"@lxqt-desktop",
	"openbox":	"openbox",
	"gnome":	"@gnome-desktop",
	"kde":		"@kde-desktop",
	"awesome":	"awesome",
	"i3":		"i3",
}

func (d dnfBase) InstallDesktop(de string) []string {
	pkg := dnfDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return d.InstallPackages([]string{pkg})
}

var dnfDisplayPkgs = map[string]string{
	"vnc":	"tigervnc-server",
	"x11":	"xorg-x11-server-Xorg",
}

func (d dnfBase) InstallDisplay(display string) []string {
	pkg := dnfDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return d.InstallPackages([]string{pkg})
}

func (d dnfBase) SetupUser(hostname string) []string {
	return []string{
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (d dnfBase) SetupOpenGL() []string {
	return d.InstallPackages([]string{"mesa-dri-drivers", "mesa-libGL", "mesa-vulkan-drivers", "mesa-libEGL"})
}

func (d dnfBase) PostInstall() []string {
	return []string{"dnf clean all"}
}
