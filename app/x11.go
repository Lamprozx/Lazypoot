package app

import (
	"fmt"
	"os"
	"path/filepath"

	"lazypoot/types"
)

func buildX11UserCmds(de, username, distroID string) []string {
	deCmd := deStartCmd(de)
	if deCmd == "" {
		return nil
	}
	src := fmt.Sprintf(`. ~/.lazypoot/x11.sh`)
	rc := userRcFile(distroID)
	return []string{
		fmt.Sprintf("mkdir -p /home/%s/.lazypoot", username),
		fmt.Sprintf(`cat > /home/%s/.lazypoot/x11.sh << 'X11EOF'
#!/bin/bash
export DISPLAY=:0
%s &
X11EOF`, username, deCmd),
		fmt.Sprintf(`grep -q 'lazypoot/x11.sh' /home/%s/%s 2>/dev/null || echo '%s' >> /home/%s/%s`, username, rc, src, username, rc),
	}
}

func GenerateHostX11Launcher(cfg types.InstallConfig) (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", fmt.Errorf("HOME not set")
	}

	deCmd := deStartCmd(cfg.DE)
	if deCmd == "" {
		return "", fmt.Errorf("unknown DE: %s", cfg.DE)
	}

	dir := filepath.Join(home, ".lazypoot", "launchers")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	shutdown := `kill -9 $(pgrep -f "virgl_test_server|termux.x11") 2>/dev/null`

	var pulseBlock string
	if cfg.InstallAudio {
		pulseBlock = `# Audio (PulseAudio)
pulseaudio --start --load="module-native-protocol-tcp auth-ip-acl=127.0.0.1 auth-anonymous=1" --exit-idle-time=-1
export PULSE_SERVER=127.0.0.1
`
	}

	var gpuBlock string
	switch cfg.AccelMode {
	case "virgl":
		gpuBlock = `# VirGL GPU acceleration
virgl_test_server_android &
`
	case "zink":
		gpuBlock = `# Zink GPU acceleration (Vulkan)
MESA_NO_ERROR=1 MESA_GL_VERSION_OVERRIDE=4.3COMPAT MESA_GLES_VERSION_OVERRIDE=3.2 GALLIUM_DRIVER=zink ZINK_DESCRIPTORS=lazy virgl_test_server --use-egl-surfaceless --use-gles &
`
	}

	prootEnv := `DISPLAY=:0`
	if cfg.AccelMode != "software" {
		prootEnv = `DISPLAY=:0 GALLIUM_DRIVER=virpipe MESA_GL_VERSION_OVERRIDE=4.0`
	}

	script := fmt.Sprintf(`#!/data/data/com.termux/files/usr/bin/bash
# X11 launcher for %s (LazyPoot)
%s

export XDG_RUNTIME_DIR=${TMPDIR}

%s# Display server
termux-x11 :0 >/dev/null &
sleep 1
am start --user 0 -n com.termux.x11/com.termux.x11.MainActivity >/dev/null 2>&1

%s
sleep 1

# Enter distro and launch desktop
proot-distro login '%s' -- bash -c "%s dbus-launch --exit-with-session %s"

exit 0
`, cfg.Distro, shutdown, pulseBlock, gpuBlock, cfg.DistroID, prootEnv, deCmd)

	path := filepath.Join(dir, cfg.DistroID+".sh")
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		return "", err
	}

	return path, nil
}
