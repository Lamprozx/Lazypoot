package plugins

func init()	{ Register(debianPlugin{aptBase: aptBase{name: "Debian", id: "debian", icon: "\uf306"}}) }

type debianPlugin struct{ aptBase }

func init()	{ Register(ubuntuPlugin{aptBase: aptBase{name: "Ubuntu", id: "ubuntu", icon: "\uf31b"}}) }

type ubuntuPlugin struct{ aptBase }

func init()	{ Register(pardusPlugin{aptBase: aptBase{name: "Pardus", id: "pardus", icon: "\uf17c"}}) }

type pardusPlugin struct{ aptBase }

func init() {
	Register(trisquelPlugin{aptBase: aptBase{name: "Trisquel GNU/Linux", id: "trisquel", icon: "\uf17c"}})
}

type trisquelPlugin struct{ aptBase }

func init()	{ Register(deepinPlugin{aptBase: aptBase{name: "Deepin", id: "deepin", icon: "\uf17c"}}) }

type deepinPlugin struct{ aptBase }

func init() {
	Register(archPlugin{pacmanBase: pacmanBase{name: "Arch Linux", id: "archlinux", icon: "\uf303"}})
}

type archPlugin struct{ pacmanBase }

func init() {
	Register(manjaroPlugin{pacmanBase: pacmanBase{name: "Manjaro", id: "manjaro", icon: "\uf312"}})
}

type manjaroPlugin struct{ pacmanBase }

func init() {
	Register(artixPlugin{pacmanBase: pacmanBase{name: "Artix Linux", id: "artix", icon: "\uf303"}})
}

type artixPlugin struct{ pacmanBase }

func init() {
	Register(alpinePlugin{apkBase: apkBase{name: "Alpine Linux", id: "alpine", icon: "\uf300"}})
}

type alpinePlugin struct{ apkBase }

func init() {
	Register(adeliePlugin{apkBase: apkBase{name: "Adelie Linux", id: "adelie", icon: "\uf300"}})
}

type adeliePlugin struct{ apkBase }

func init() {
	Register(chimeraPlugin{apkBase: apkBase{name: "Chimera Linux", id: "chimera", icon: "\uf17c"}})
}

type chimeraPlugin struct{ apkBase }

func init() {
	Register(fedoraPlugin{dnfBase: dnfBase{name: "Fedora", id: "fedora", icon: "\uf30a"}})
}

type fedoraPlugin struct{ dnfBase }

func init() {
	Register(oraclePlugin{dnfBase: dnfBase{name: "Oracle Linux", id: "oracle", icon: "\uf17c"}})
}

type oraclePlugin struct{ dnfBase }

func init() {
	Register(almaPlugin{dnfBase: dnfBase{name: "AlmaLinux", id: "almalinux", icon: "\uf17c"}})
}

type almaPlugin struct{ dnfBase }

func init() {
	Register(rockyPlugin{dnfBase: dnfBase{name: "Rocky Linux", id: "rockylinux", icon: "\uf17c"}})
}

type rockyPlugin struct{ dnfBase }
