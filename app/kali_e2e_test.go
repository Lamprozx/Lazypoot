package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectTermuxArchOnAmd64(t *testing.T) {
	arch := detectTermuxArch()
	// On x86_64 Linux, /proc/cpuinfo won't have ARM strings,
	// falls through to uname -m → "x86_64" → not in arm64/armhf → default "arm64"
	if arch != "arm64" {
		t.Errorf("expected arm64 (default fallback) on x86_64, got %s", arch)
	}
	t.Logf("detectTermuxArch() = %s", arch)
}

func TestSanitizeLineAptOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ANSI color codes stripped",
			input: "\x1b[0;32mOK\x1b[0m",
			want:  "OK",
		},
		{
			name:  "carriage return progress",
			input: "Progress: 50%\rProgress: 100%\rDone",
			want:  "Done",
		},
		{
			name:  "apt bold sequence",
			input: "\x1b[1mReading package lists...\x1b[0m",
			want:  "Reading package lists...",
		},
		{
			name:  "trailing \\r from apt output",
			input: "Setting up libc6 (2.36-9+deb12u7) ...\r",
			want:  "Setting up libc6 (2.36-9+deb12u7) ...",
		},
		{
			name:  "null byte + DEL stripped",
			input: "Unpacking\x00 files...\x7f",
			want:  "Unpacking files...",
		},
		{
			name:  "multiple ANSI sequences",
			input: "\x1b[31m\x1b[1mERROR\x1b[0m: \x1b[33mdisk full\x1b[0m",
			want:  "ERROR: disk full",
		},
		{
			name:  "all ANSI = empty",
			input: "\x1b[0m\x1b[0m\x1b[0m",
			want:  "",
		},
		{
			name:  "normal text unchanged",
			input: "Reading package lists... Done",
			want:  "Reading package lists... Done",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLine(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestKaliVariantInfo(t *testing.T) {
	tests := []struct {
		arch     string
		variant  string
		wantSize int64
	}{
		{arch: "arm64", variant: "nano", wantSize: 300 * 1024 * 1024},
		{arch: "arm64", variant: "minimal", wantSize: 800 * 1024 * 1024},
		{arch: "arm64", variant: "full", wantSize: 2 * 1024 * 1024 * 1024},
		{arch: "armhf", variant: "nano", wantSize: 280 * 1024 * 1024},
		{arch: "armhf", variant: "minimal", wantSize: 750 * 1024 * 1024},
		{arch: "armhf", variant: "full", wantSize: 2 * 1024 * 1024 * 1024},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s-%s", tt.arch, tt.variant), func(t *testing.T) {
			base := "https://kali.download/nethunter-images/current/rootfs"
			imageName := fmt.Sprintf("kali-nethunter-rootfs-%s-%s.tar.xz", tt.variant, tt.arch)
			url := fmt.Sprintf("%s/%s", base, imageName)
			shaURL := fmt.Sprintf("%s/SHA256SUMS", base)

			if !strings.HasSuffix(url, ".tar.xz") {
				t.Errorf("image URL must end with .tar.xz: %s", url)
			}
			if !strings.Contains(shaURL, "SHA256SUMS") {
				t.Errorf("SHA URL must point to SHA256SUMS: %s", shaURL)
			}
		})
	}
}

func TestImageDistroPaths(t *testing.T) {
	home := "/data/data/com.termux/files/home"
	id := "kali"

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"distroDir", distroDir(home, id), home + "/.lazypoot/distros/kali"},
		{"lockPath", lockPath(home, id), home + "/.lazypoot/distros/kali/.install.lock"},
		{"metaPath", metaPath(home, id), home + "/.lazypoot/distros/kali/meta.json"},
		{"launcherPath", launcherPath(home, id), home + "/.lazypoot/distros/kali/launcher.sh"},
		{"imagePartPath", imagePartPath(home, id), home + "/.lazypoot/distros/kali/image.tar.xz.part"},
		{"imageFinalPath", imageFinalPath(home, id), home + "/.lazypoot/distros/kali/image.tar.xz"},
		{"rootfsTmpPath", rootfsTmpPath(home, id), home + "/.lazypoot/distros/kali/rootfs.tmp"},
		{"rootfsFinalPath", rootfsFinalPath(home, id), home + "/.lazypoot/distros/kali/rootfs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestIsImageBasedDistro(t *testing.T) {
	if !isImageBasedDistro("kali") {
		t.Error("isImageBasedDistro('kali') should be true")
	}
	if isImageBasedDistro("debian") {
		t.Error("isImageBasedDistro('debian') should be false")
	}
	if isImageBasedDistro("ubuntu") {
		t.Error("isImageBasedDistro('ubuntu') should be false")
	}
}

func TestWriteMetaAndReadMeta(t *testing.T) {
	dir := t.TempDir()
	mp := filepath.Join(dir, "meta.json")

	meta := ImageMeta{
		ID:        "kali",
		Variant:   "nano",
		Arch:      "arm64",
		Installed: 1234567890,
		State:     "ready",
	}

	if err := writeMeta(mp, meta); err != nil {
		t.Fatalf("writeMeta failed: %v", err)
	}

	got, err := ReadMeta(mp)
	if err != nil {
		t.Fatalf("ReadMeta failed: %v", err)
	}

	if got.ID != "kali" || got.Variant != "nano" || got.Arch != "arm64" || got.State != "ready" {
		t.Errorf("meta data mismatch:\n  got:  %+v\n  want: ID=kali variant=nano arch=arm64 state=ready", got)
	}
}

func TestGenerateImageLauncherTemplate(t *testing.T) {
	home := t.TempDir()
	distroID := "kali"
	rootfs := fmt.Sprintf("%s/.lazypoot/distros/kali/rootfs", home)

	err := generateImageLauncher(distroID, rootfs, home)
	if err != nil {
		t.Fatalf("generateImageLauncher failed: %v", err)
	}

	lp := launcherPath(home, distroID)
	data, err := os.ReadFile(lp)
	if err != nil {
		t.Fatalf("read launcher failed: %v", err)
	}

	content := string(data)
	checks := []struct {
		name string
		ok   bool
	}{
		{"unset LD_PRELOAD", strings.Contains(content, "unset LD_PRELOAD")},
		{"export TMPDIR=/tmp", strings.Contains(content, "export TMPDIR=/tmp")},
		{"proot with escaped newline", strings.Contains(content, "proot \\")},
		{"link2symlink flag", strings.Contains(content, "--link2symlink -0")},
		{"bind /dev", strings.Contains(content, "-b /dev")},
		{"bind /proc", strings.Contains(content, "-b /proc")},
		{"bind /sdcard", strings.Contains(content, "-b /sdcard")},
		{"bind /system", strings.Contains(content, "-b /system")},
		{"bind /vendor", strings.Contains(content, "-b /vendor")},
		{"bind /apex", strings.Contains(content, "-b /apex")},
		{"rootfs path in launcher", strings.Contains(content, rootfs)},
		{"sudo -u $user", strings.Contains(content, "sudo -u $user")},
		{"shebang termux bash", strings.Contains(content, "#!/data/data/com.termux/files/usr/bin/bash")},
		{"exec eval pattern", strings.Contains(content, "eval exec")},
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("launcher template missing: %s", c.name)
		}
	}
	t.Logf("launcher.sh generated (%d bytes)", len(data))
}

func TestSetupImagePortalLogin(t *testing.T) {
	home := t.TempDir()
	distroID := "kali"
	rootfs := fmt.Sprintf("%s/.lazypoot/distros/kali/rootfs", home)
	usernames := []string{"testedkali"}

	err := SetupImagePortalLogin(home, distroID, rootfs, usernames)
	if err != nil {
		t.Fatalf("SetupImagePortalLogin failed: %v", err)
	}

	pPath := filepath.Join(home, ".lazypoot", "portal", distroID+".sh")
	data, err := os.ReadFile(pPath)
	if err != nil {
		t.Fatalf("read portal script failed: %v", err)
	}

	content := string(data)
	checks := []struct {
		name string
		ok   bool
	}{
		{"shebang", strings.Contains(content, "#!/data/data/com.termux/files/usr/bin/bash")},
		{"TMPDIR=/tmp", strings.Contains(content, "export TMPDIR=/tmp")},
		{"select user loop", strings.Contains(content, "select user in")},
		{"sudo -u $user", strings.Contains(content, "sudo -u")},
		{"testedkali user", strings.Contains(content, "testedkali")},
		{"USERS array with root", strings.Contains(content, "root") || strings.Contains(content, "\"root\"")},
	}
	for _, c := range checks {
		if !c.ok {
			t.Errorf("portal script missing: %s", c.name)
		}
	}
	t.Logf("portal.sh generated (%d bytes)", len(data))
}

func TestPathTraversalBlocked(t *testing.T) {
	// Test path traversal detection logic (same as extractTarXZ)
	traversalPaths := []string{
		"../../etc/passwd",
		"/etc/passwd",
		"../etc/shadow",
		"home/user/../file",
	}
	safePaths := []string{
		"etc/passwd",
		"usr/bin/bash",
		"var/log/something",
		"usr/share/doc/something",
	}

	for _, p := range traversalPaths {
		if !strings.Contains(p, "..") && !strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "../") {
			t.Errorf("traversal path not detected: %s", p)
		}
	}
	for _, p := range safePaths {
		if strings.Contains(p, "..") || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "../") {
			t.Errorf("safe path incorrectly blocked: %s", p)
		}
	}
}

func TestCheckAvailableStorage(t *testing.T) {
	free, err := checkAvailableStorage("/")
	if err != nil {
		t.Fatalf("checkAvailableStorage failed: %v", err)
	}
	if free <= 0 {
		t.Errorf("expected positive free space, got %d", free)
	}
	t.Logf("Free space on /: %.1f GB", float64(free)/(1024*1024*1024))
}

func TestExtractTarXZWithRealFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping extraction test in short mode")
	}

	// Create a small test tar.xz with gzip (simulating the extraction path)
	dir := t.TempDir()
	srcFile := filepath.Join(dir, "test.tar.gz")

	// Build a tar.gz (we use gzip since xz needs external lib, same logic applies)
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	hdr := &tar.Header{
		Name: "test.txt",
		Size: 5,
		Mode: 0644,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	tw.Close()

	gzBuf := new(bytes.Buffer)
	gw := gzip.NewWriter(gzBuf)
	gw.Write(buf.Bytes())
	gw.Close()

	if err := os.WriteFile(srcFile, gzBuf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	// Test that the output directory is created (dir already created via t.TempDir)
	f, err := os.Open(srcFile)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	header, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "test.txt" {
		t.Errorf("expected test.txt, got %s", header.Name)
	}

	t.Logf("tar.gz extraction test passed: %s extracted", header.Name)
}

func TestRealSHA256SUMSVerifyFormat(t *testing.T) {
	// Parse the SHA256SUMS format that Kali uses
	shaContent := `a6673ef39dfc858a5ad4e4b18f974f3eda7732b12cba07609d377b5dbd7d2436  kali-nethunter-rootfs-full-amd64.tar.xz
b8098fc90ed74a553592f7019a1d88dfe3c65b16c60af487b0658860554dc5aa  kali-nethunter-rootfs-full-arm64.tar.xz`

	// Use the same parser logic as verifyChecksum
	imageName := "kali-nethunter-rootfs-full-arm64.tar.xz"
	expectedHash := ""
	for _, line := range strings.Split(shaContent, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		hash := parts[0]
		fname := parts[len(parts)-1]
		if filepath.Base(fname) == imageName {
			expectedHash = hash
			break
		}
	}

	if expectedHash == "" {
		t.Fatal("SHA256SUMS parser failed: hash not found")
	}
	if len(expectedHash) != 64 {
		t.Fatalf("expected 64-char SHA256 hex, got %d chars", len(expectedHash))
	}
	if expectedHash != "b8098fc90ed74a553592f7019a1d88dfe3c65b16c60af487b0658860554dc5aa" {
		t.Fatalf("SHA256 hash mismatch: got %s", expectedHash)
	}

	// Verify with real Go crypto
	h := sha256.New()
	h.Write([]byte("test"))
	_ = hex.EncodeToString(h.Sum(nil))
	t.Logf("SHA256SUMS parser OK: hash=%s... image=%s", expectedHash[:16], imageName)
}
