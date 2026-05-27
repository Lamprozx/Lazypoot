package plugins

import "fmt"

type parrotPlugin struct{}

func init() { Register(parrotPlugin{}) }

func (p parrotPlugin) Name() string          { return "Parrot OS" }
func (p parrotPlugin) ID() string            { return "parrot" }
func (p parrotPlugin) Icon() string           { return "\uf327" }
func (p parrotPlugin) PackageManager() string  { return "apt" }

func (p parrotPlugin) BootstrapVariants(arch string) ([]BootstrapVariant, error) {
	switch arch {
	case "arm64":
		return []BootstrapVariant{
			{ID: "core", Label: "Core (Minimum)", PackageSets: []string{"parrot-core"}, EstimatedSizeGB: 3},
			{ID: "home", Label: "Home Edition", PackageSets: []string{"parrot-interface", "parrot-interface-home"}, EstimatedSizeGB: 7},
			{ID: "full", Label: "Full (Security Tools)", PackageSets: []string{"parrot-tools-full"}, EstimatedSizeGB: 10},
		}, nil
	case "armhf":
		return []BootstrapVariant{
			{ID: "core", Label: "Core (Minimum)", PackageSets: []string{"parrot-core"}, EstimatedSizeGB: 3},
			{ID: "home", Label: "Home Edition", PackageSets: []string{"parrot-interface", "parrot-interface-home"}, EstimatedSizeGB: 7},
			{ID: "full", Label: "Full (Security Tools)", PackageSets: []string{"parrot-tools-full"}, EstimatedSizeGB: 10},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported arch for parrot: %s", arch)
	}
}

func (p parrotPlugin) UpdateSystem() []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get update -y",
		"DEBIAN_FRONTEND=noninteractive apt-get install -y sudo",
	}
}

func (p parrotPlugin) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", joinPkgs(pkgs)),
	}
}

var parrotDEPkgs = map[string]string{
	"xfce":    "parrot-desktop-xfce",
	"lxqt":    "lxqt",
	"openbox": "openbox",
	"gnome":   "parrot-desktop-gnome",
	"kde":     "parrot-desktop-kde",
	"awesome": "awesome",
	"i3":      "i3",
}

func (p parrotPlugin) InstallDesktop(de string) []string {
	pkg := parrotDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return p.InstallPackages([]string{pkg})
}

var parrotDisplayPkgs = map[string]string{
	"vnc": "tigervnc-standalone-server",
	"x11": "xorg x11-xserver-utils",
}

func (p parrotPlugin) InstallDisplay(display string) []string {
	pkg := parrotDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return p.InstallPackages([]string{pkg})
}

func (p parrotPlugin) SetupUser(hostname string) []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get install -y hostname",
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (p parrotPlugin) SetupOpenGL() []string {
	return p.InstallPackages([]string{"mesa-utils", "libgl1-mesa-dri", "mesa-vulkan-drivers", "libgles2-mesa"})
}

func (p parrotPlugin) PostInstall() []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get autoremove -y",
		"apt-get clean",
	}
}
