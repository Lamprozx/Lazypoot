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

type ImageVariant struct {
	ID        string
	Label     string
	Size      int64 // compressed bytes
	Extracted int64 // estimated extracted bytes
	ImageName string
	ImageURL  string
	SHAURL    string
	Mirrors   []string
}

type ImageBasedDistro interface {
	DistroPlugin
	Variants(arch string) ([]ImageVariant, error)
	SupportedArchs() []string
}
