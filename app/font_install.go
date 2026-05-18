package app

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lazypoot/types"

	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	os.Setenv("GODEBUG", "netdns=go")
}

const nerdFontReleaseURL = "https://github.com/ryanoasis/nerd-fonts/releases/download/v3.4.0/FiraCode.zip"

const expectedFontSHA256 = "29b619655612cb273e034737408b9508a04beb63c1ddbdfaa9a6846c409c7a2e"

func runFontInstallCmd() (chan tea.Msg, tea.Cmd) {
	ch := make(chan tea.Msg, 128)
	go fontInstallPipeline(ch)
	return ch, tea.Batch(
		func() tea.Msg { return FontInstallStartedMsg{} },
		waitForFontLog(ch),
	)
}

func waitForFontLog(ch chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}

func fontInstallPipeline(ch chan tea.Msg) {
	defer close(ch)

	send := func(step int, lvl types.LogLevel, msg string) {
		ch <- FontInstallProgressMsg{Line: msg, Level: lvl, Step: step}
	}

	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	termuxDir := filepath.Join(home, ".termux")
	os.MkdirAll(termuxDir, 0755)

	send(1, types.LogStep, "Downloading FiraCode Nerd Font...")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	zipData, err := downloadBytes(ctx, nerdFontReleaseURL)
	if err != nil {
		if strings.Contains(err.Error(), "tls") || strings.Contains(err.Error(), "certificate") {
			exec.Command("pkg", "install", "ca-certificates", "-y").Run()
			zipData, err = downloadBytes(ctx, nerdFontReleaseURL)
		}
		if err != nil {
			ch <- FontInstallErrorMsg{Err: "Download failed: " + err.Error()}
			return
		}
	}
	send(1, types.LogOK, fmt.Sprintf("Downloaded (%d KB)", len(zipData)/1024))

	send(2, types.LogStep, "Extracting font...")
	fontData, fontName, err := extractFontFromZip(zipData)
	if err != nil {
		ch <- FontInstallErrorMsg{Err: "Extract failed: " + err.Error()}
		return
	}
	send(2, types.LogOK, fmt.Sprintf("Found: %s", fontName))

	sum := sha256.Sum256(fontData)
	got := hex.EncodeToString(sum[:])
	if got != expectedFontSHA256 {
		ch <- FontInstallErrorMsg{Err: fmt.Sprintf("SHA256 mismatch: expected %s, got %s", expectedFontSHA256, got)}
		return
	}
	send(2, types.LogOK, "SHA256 verified")

	fontPath := filepath.Join(termuxDir, "font.ttf")
	if err := os.WriteFile(fontPath, fontData, 0644); err != nil {
		ch <- FontInstallErrorMsg{Err: "Write failed: " + err.Error()}
		return
	}

	hashPath := fontPath + ".sha256"
	_ = os.WriteFile(hashPath, []byte(hex.EncodeToString(sum[:])+"\n"), 0644)

	send(3, types.LogOK, "Font installed to ~/.termux/font.ttf")

	send(4, types.LogStep, "Refreshing Termux...")
	_ = exec.Command("termux-reload-settings").Run()
	send(4, types.LogOK, "Nerd Font installed! Restart Termux if needed.")
	ch <- FontInstallDoneMsg{}
}

var fallbackDNS = &net.Resolver{
	PreferGo: true,
	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{
			Timeout: 3 * time.Second,
		}

		return d.DialContext(ctx, "udp4", "8.8.8.8:53")
	},
}

func certPool() *x509.CertPool {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	caCert, err := os.ReadFile("/data/data/com.termux/files/usr/etc/tls/cert.pem")
	if err == nil {
		pool.AppendCertsFromPEM(caCert)
	}
	return pool
}

var httpClient = &http.Client{
	Timeout: 2 * time.Minute,
	Transport: &http.Transport{
		ForceAttemptHTTP2: true,
		TLSClientConfig: &tls.Config{
			RootCAs: certPool(),
		},
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			Resolver:  fallbackDNS,
		}).DialContext,
	},
}

const maxFontDownloadBytes = 50 * 1024 * 1024

func downloadBytes(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned HTTP %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, maxFontDownloadBytes)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(data) >= maxFontDownloadBytes {
		return nil, fmt.Errorf("download exceeded maximum size of %d MB", maxFontDownloadBytes/(1024*1024))
	}
	return data, nil
}

func extractFontFromZip(zipData []byte) (data []byte, name string, err error) {
	r, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, "", err
	}
	for _, f := range r.File {
		if !strings.HasSuffix(strings.ToLower(f.Name), ".ttf") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, "", err
		}
		d, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, "", err
		}

		if strings.EqualFold(filepath.Base(f.Name), "FiraCodeNerdFont-Regular.ttf") {
			return d, filepath.Base(f.Name), nil
		}
	}
	return nil, "", fmt.Errorf("FiraCodeNerdFont-Regular.ttf not found in archive")
}
