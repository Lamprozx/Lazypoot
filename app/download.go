package app

import (
	"archive/tar"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lazypoot/types"

	"github.com/ulikunitz/xz"
	tea "github.com/charmbracelet/bubbletea"
)

type ImageInstallProgressMsg struct {
	Percent int
}

type DownloadProgressMsg struct {
	Downloaded int64
	Total      int64
	Label      string
}

type DownloadDoneMsg struct{}

type DownloadErrorMsg struct {
	Err string
}

type ExtractProgressMsg struct {
	Label  string
	Count  int
	Total  int
}

type ExtractDoneMsg struct{}

type ExtractErrorMsg struct {
	Err string
}

func detectTermuxArch() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		info := string(data)
		if strings.Contains(info, "v7l") || strings.Contains(info, "v7") || strings.Contains(info, "ARMv7") || strings.Contains(info, "CPU architecture: 7") {
			return "armhf"
		}
		if strings.Contains(info, "v8") || strings.Contains(info, "AArch64") || strings.Contains(info, "aarch64") || strings.Contains(info, "ARMv8") {
			return "arm64"
		}
	}
	u, uerr := exec.Command("uname", "-m").Output()
	if uerr == nil {
		um := strings.TrimSpace(string(u))
		switch um {
		case "aarch64", "arm64":
			return "arm64"
		case "armv7l", "armv7", "armhf":
			return "armhf"
		}
	}
	return "arm64"
}

func distroDir(home, distroID string) string {
	return filepath.Join(home, ".lazypoot", "distros", distroID)
}

func lockPath(home, distroID string) string {
	return filepath.Join(distroDir(home, distroID), ".install.lock")
}

func metaPath(home, distroID string) string {
	return filepath.Join(distroDir(home, distroID), "meta.json")
}

func launcherPath(home, distroID string) string {
	return filepath.Join(distroDir(home, distroID), "launcher.sh")
}

func imagePartPath(home, distroID string) string {
	return filepath.Join(distroDir(home, distroID), "image.tar.xz.part")
}

func imageFinalPath(home, distroID string) string {
	return filepath.Join(distroDir(home, distroID), "image.tar.xz")
}

func rootfsTmpPath(home, distroID string) string {
	return filepath.Join(distroDir(home, distroID), "rootfs.tmp")
}

func rootfsFinalPath(home, distroID string) string {
	return filepath.Join(distroDir(home, distroID), "rootfs")
}

var downloadClient = &http.Client{
	Timeout: 30 * time.Minute,
	Transport: &http.Transport{
		ResponseHeaderTimeout: 30 * time.Second,
	},
}

func downloadWithResume(ctx context.Context, ch chan<- tea.Msg, url, dest string) (int64, error) {
	existing := int64(0)
	if stat, err := os.Stat(dest); err == nil {
		existing = stat.Size()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	if existing > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existing))
	}

	resp, err := downloadClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	total := resp.ContentLength + existing
	if total <= 0 {
		total = 1
	}

	flag := os.O_WRONLY | os.O_CREATE
	if existing > 0 {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	f, err := os.OpenFile(dest, flag, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	written := existing
	lastReport := time.Now()

	for {
		select {
		case <-ctx.Done():
			f.Close()
			return written, ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := f.Write(buf[:n]); writeErr != nil {
				return written, writeErr
			}
			written += int64(n)
			if time.Since(lastReport) > 200*time.Millisecond {
				ch <- DownloadProgressMsg{
					Downloaded: written,
					Total:      total,
					Label:      fmt.Sprintf("%.1f MB / %.1f MB", float64(written)/(1024*1024), float64(total)/(1024*1024)),
				}
				lastReport = time.Now()
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return written, readErr
		}
	}

	return written, nil
}

func verifyChecksum(imagePath, shaURL string) error {
	req, err := http.NewRequest("GET", shaURL, nil)
	if err != nil {
		return fmt.Errorf("sha request: %w", err)
	}
	resp, err := downloadClient.Do(req)
	if err != nil {
		return fmt.Errorf("sha download: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sha read: %w", err)
	}

	expectedHash := ""
	imageName := filepath.Base(imagePath)
	for _, line := range strings.Split(string(body), "\n") {
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
		fname = strings.TrimPrefix(fname, "*")
		fname = strings.TrimPrefix(fname, " ")
		if filepath.Base(fname) == imageName || fname == imageName {
			expectedHash = hash
			break
		}
	}

	if expectedHash == "" {
		return fmt.Errorf("checksum not found for %s in sha file", imageName)
	}

	var h hash.Hash
	switch len(expectedHash) {
	case 64:
		h = sha256.New()
	case 128:
		h = sha512.New()
	default:
		return fmt.Errorf("unsupported hash length: %d (expected 64 for SHA256 or 128 for SHA512)", len(expectedHash))
	}

	f, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(h, f); err != nil {
		return fmt.Errorf("hash compute: %w", err)
	}
	got := hex.EncodeToString(h.Sum(nil))

	if got != expectedHash {
		return fmt.Errorf("checksum mismatch: got %s, expected %s", got[:16], expectedHash[:16])
	}
	return nil
}

func extractTarXZ(ch chan<- tea.Msg, src, tmpDir string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	xzReader, err := xz.NewReader(f)
	if err != nil {
		return fmt.Errorf("xz init: %w", err)
	}

	tarReader := tar.NewReader(xzReader)
	count := 0

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar read: %w", err)
		}

		name := header.Name
		if strings.Contains(name, "..") || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") {
			return fmt.Errorf("path traversal blocked: %s", name)
		}

		target := filepath.Join(tmpDir, name)
		dir := filepath.Dir(target)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		}

		count++
		if count%100 == 0 {
			ch <- ExtractProgressMsg{Label: fmt.Sprintf("extracted %d files", count), Count: count}
		}
	}

	ch <- ExtractProgressMsg{Label: fmt.Sprintf("extracted %d files", count), Count: count}
	return nil
}

func checkAvailableStorage(path string) (int64, error) {
	var stat types.Statfs
	if err := stat.Syscall(path); err != nil {
		return 0, err
	}
	return stat.Available(), nil
}


