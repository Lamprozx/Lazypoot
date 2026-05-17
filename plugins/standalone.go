package plugins

import "fmt"

func init()	{ Register(opensusePlugin{}) }

type opensusePlugin struct{}

func (o opensusePlugin) Name() string		{ return "OpenSUSE" }
func (o opensusePlugin) ID() string		{ return "opensuse" }
func (o opensusePlugin) Icon() string		{ return "\uf314" }
func (o opensusePlugin) PackageManager() string	{ return "zypper" }

func (o opensusePlugin) UpdateSystem() []string {
	return []string{"zypper refresh", "zypper update -y"}
}

func (o opensusePlugin) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("zypper install -y %s", joinPkgs(pkgs))}
}

var opensuseDEPkgs = map[string]string{
	"xfce":		"patterns-xfce-xfce",
	"lxqt":		"lxqt",
	"openbox":	"openbox",
	"gnome":	"patterns-gnome-gnome",
	"kde":		"patterns-kde-kde",
	"mate":		"mate",
	"cinnamon":	"cinnamon",
}

func (o opensusePlugin) InstallDesktop(de string) []string {
	pkg := opensuseDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return o.InstallPackages([]string{pkg})
}

var opensuseDisplayPkgs = map[string]string{
	"vnc":	"tigervnc",
	"x11":	"xorg-x11-server",
}

func (o opensusePlugin) InstallDisplay(display string) []string {
	pkg := opensuseDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return o.InstallPackages([]string{pkg})
}

func (o opensusePlugin) SetupUser(hostname string) []string {
	return []string{
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (o opensusePlugin) SetupOpenGL() []string {
	return o.InstallPackages([]string{"Mesa", "Mesa-libGL1", "Mesa-dri", "Mesa-vulkan-drivers"})
}

func (o opensusePlugin) PostInstall() []string {
	return []string{"zypper clean -a"}
}

func init()	{ Register(voidPlugin{}) }

type voidPlugin struct{}

func (v voidPlugin) Name() string		{ return "Void Linux" }
func (v voidPlugin) ID() string			{ return "void" }
func (v voidPlugin) Icon() string		{ return "\uf317" }
func (v voidPlugin) PackageManager() string	{ return "xbps" }

func (v voidPlugin) UpdateSystem() []string {
	return []string{"xbps-install -Su xbps", "xbps-install -Su"}
}

func (v voidPlugin) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("xbps-install -y %s", joinPkgs(pkgs))}
}

var voidDEPkgs = map[string]string{
	"xfce":		"xfce4",
	"lxqt":		"lxqt",
	"openbox":	"openbox",
	"gnome":	"gnome",
	"kde":		"kde5",
	"mate":		"mate",
	"cinnamon":	"cinnamon",
}

func (v voidPlugin) InstallDesktop(de string) []string {
	pkg := voidDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return v.InstallPackages([]string{pkg})
}

var voidDisplayPkgs = map[string]string{
	"vnc":	"tigervnc",
	"x11":	"xorg-minimal",
}

func (v voidPlugin) InstallDisplay(display string) []string {
	pkg := voidDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return v.InstallPackages([]string{pkg})
}

func (v voidPlugin) SetupUser(hostname string) []string {
	return []string{
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (v voidPlugin) SetupOpenGL() []string {
	return v.InstallPackages([]string{"mesa", "mesa-dri", "mesa-vulkan-radeon", "mesa-vulkan-intel"})
}

func (v voidPlugin) PostInstall() []string {
	return []string{"xbps-remove -Oo"}
}
