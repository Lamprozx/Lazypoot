package plugins

import "fmt"

type apkBase struct {
	name	string
	id	string
	icon	string
}

func (a apkBase) Name() string			{ return a.name }
func (a apkBase) ID() string			{ return a.id }
func (a apkBase) Icon() string			{ return a.icon }
func (a apkBase) PackageManager() string	{ return "apk" }

func (a apkBase) UpdateSystem() []string {
	return []string{"apk update", "apk upgrade"}
}

func (a apkBase) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{fmt.Sprintf("apk add %s", joinPkgs(pkgs))}
}

var apkDEPkgs = map[string]string{
	"xfce":		"xfce4 xfce4-extras",
	"lxqt":		"lxqt",
	"openbox":	"openbox",
	"gnome":	"gnome",
	"kde":		"plasma",
	"awesome":	"awesome",
	"i3":		"i3wm",
}

func (a apkBase) InstallDesktop(de string) []string {
	pkg := apkDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return a.InstallPackages([]string{pkg})
}

var apkDisplayPkgs = map[string]string{
	"vnc":	"tigervnc",
	"x11":	"xorg-server",
}

func (a apkBase) InstallDisplay(display string) []string {
	pkg := apkDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return a.InstallPackages([]string{pkg})
}

func (a apkBase) SetupUser(hostname string) []string {
	return []string{
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (a apkBase) SetupOpenGL() []string {
	return a.InstallPackages([]string{"mesa", "mesa-dri-gallium", "mesa-vulkan-gallium", "mesa-gl"})
}

func (a apkBase) PostInstall() []string {
	return []string{"apk cache clean"}
}
