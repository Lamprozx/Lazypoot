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
	Size      int64
	Extracted int64
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

type BootstrapVariant struct {
	ID              string
	Label           string
	PackageSets     []string
	EstimatedSizeGB int
}

type MutableRootfsDistro interface {
	DistroPlugin
	BootstrapVariants(arch string) ([]BootstrapVariant, error)
}
