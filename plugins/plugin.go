package plugins

type DistroPlugin interface {
	Name() string
	ID() string
	Icon() string
	PackageManager() string

	UpdateSystem() []string
	InstallPackages(pkgs []string) []string
	InstallDesktop(de string) []string
	InstallDisplay(display string) []string
	SetupUser(hostname string) []string
	SetupOpenGL() []string
	PostInstall() []string
}
