package plugins

import "fmt"

type kaliPlugin struct{}

func init() { Register(kaliPlugin{}) }

func (k kaliPlugin) Name() string             { return "Kali Linux" }
func (k kaliPlugin) ID() string               { return "kali" }
func (k kaliPlugin) Icon() string             { return "\uf327" }
func (k kaliPlugin) PackageManager() string    { return "apt" }

func (k kaliPlugin) SupportedArchs() []string { return []string{"arm64", "armhf"} }

func (k kaliPlugin) Variants(arch string) ([]ImageVariant, error) {
	base := "https://kali.download/nethunter-images/current/rootfs"
	switch arch {
	case "arm64":
		return []ImageVariant{
			{
				ID: "nano", Label: "Nano", Size: 300 * 1024 * 1024, Extracted: 1 * 1024 * 1024 * 1024,
				ImageName: "kali-nethunter-rootfs-nano-arm64.tar.xz",
				ImageURL:  base + "/kali-nethunter-rootfs-nano-arm64.tar.xz",
				SHAURL:    base + "/SHA256SUMS",
				Mirrors:   nil,
			},
			{
				ID: "minimal", Label: "Minimal", Size: 800 * 1024 * 1024, Extracted: 3 * 1024 * 1024 * 1024,
				ImageName: "kali-nethunter-rootfs-minimal-arm64.tar.xz",
				ImageURL:  base + "/kali-nethunter-rootfs-minimal-arm64.tar.xz",
				SHAURL:    base + "/SHA256SUMS",
				Mirrors:   nil,
			},
			{
				ID: "full", Label: "Full", Size: 2 * 1024 * 1024 * 1024, Extracted: 6 * 1024 * 1024 * 1024,
				ImageName: "kali-nethunter-rootfs-full-arm64.tar.xz",
				ImageURL:  base + "/kali-nethunter-rootfs-full-arm64.tar.xz",
				SHAURL:    base + "/SHA256SUMS",
				Mirrors:   nil,
			},
		}, nil
	case "armhf":
		return []ImageVariant{
			{
				ID: "nano", Label: "Nano", Size: 280 * 1024 * 1024, Extracted: 900 * 1024 * 1024,
				ImageName: "kali-nethunter-rootfs-nano-armhf.tar.xz",
				ImageURL:  base + "/kali-nethunter-rootfs-nano-armhf.tar.xz",
				SHAURL:    base + "/SHA256SUMS",
				Mirrors:   nil,
			},
			{
				ID: "minimal", Label: "Minimal", Size: 750 * 1024 * 1024, Extracted: 3 * 1024 * 1024 * 1024,
				ImageName: "kali-nethunter-rootfs-minimal-armhf.tar.xz",
				ImageURL:  base + "/kali-nethunter-rootfs-minimal-armhf.tar.xz",
				SHAURL:    base + "/SHA256SUMS",
				Mirrors:   nil,
			},
			{
				ID: "full", Label: "Full", Size: 2 * 1024 * 1024 * 1024, Extracted: 6 * 1024 * 1024 * 1024,
				ImageName: "kali-nethunter-rootfs-full-armhf.tar.xz",
				ImageURL:  base + "/kali-nethunter-rootfs-full-armhf.tar.xz",
				SHAURL:    base + "/SHA256SUMS",
				Mirrors:   nil,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported arch for kali: %s", arch)
	}
}

func (k kaliPlugin) UpdateSystem() []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get update -y",
		"DEBIAN_FRONTEND=noninteractive apt-get install -y sudo",
	}
}

func (k kaliPlugin) InstallPackages(pkgs []string) []string {
	if len(pkgs) == 0 {
		return nil
	}
	return []string{
		fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", joinPkgs(pkgs)),
	}
}

var kaliDEPkgs = map[string]string{
	"xfce":    "kali-desktop-xfce",
	"lxqt":    "kali-desktop-lxqt",
	"openbox": "openbox",
	"gnome":   "kali-desktop-gnome",
	"kde":     "kali-desktop-kde",
	"awesome": "awesome",
	"i3":      "i3",
}

func (k kaliPlugin) InstallDesktop(de string) []string {
	pkg := kaliDEPkgs[de]
	if pkg == "" {
		return nil
	}
	return k.InstallPackages([]string{pkg})
}

var kaliDisplayPkgs = map[string]string{
	"vnc": "tigervnc-standalone-server",
	"x11": "xorg x11-xserver-utils",
}

func (k kaliPlugin) InstallDisplay(display string) []string {
	pkg := kaliDisplayPkgs[display]
	if pkg == "" {
		return nil
	}
	return k.InstallPackages([]string{pkg})
}

func (k kaliPlugin) SetupUser(hostname string) []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get install -y hostname",
		fmt.Sprintf("hostname %s", hostname),
		fmt.Sprintf("echo '%s' > /etc/hostname", hostname),
	}
}

func (k kaliPlugin) SetupOpenGL() []string {
	return k.InstallPackages([]string{"mesa-utils", "libgl1-mesa-dri", "mesa-vulkan-drivers", "libgles2-mesa"})
}

func (k kaliPlugin) PostInstall() []string {
	return []string{
		"DEBIAN_FRONTEND=noninteractive apt-get autoremove -y",
		"apt-get clean",
	}
}
