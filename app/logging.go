package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lazypoot/types"
)

const (
	EvStepDownload   = "STEP_DOWNLOAD"
	EvStepExtract    = "STEP_EXTRACT"
	EvStepVerify     = "STEP_VERIFY"
	EvStepInstall    = "STEP_INSTALL"
	EvErrTLS         = "ERR_TLS_CERT"
	EvErrDNS         = "ERR_DNS"
	EvErrDownload    = "ERR_DOWNLOAD"
	EvErrSHA         = "ERR_SHA_MISMATCH"
	EvHookWriteOK    = "HOOK_WRITE_OK"
	EvHookWriteErr   = "HOOK_WRITE_ERR"
	EvManifestRead   = "MANIFEST_READ"
	EvManifestWrite  = "MANIFEST_WRITE"
	EvLockAcquired   = "LOCK_ACQUIRED"
	EvLockStale      = "LOCK_STALE"
	EvLockHeld       = "LOCK_HELD"
	EvSyncStart      = "SYNC_START"
	EvSyncDone       = "SYNC_DONE"
	EvBackupCreated  = "BACKUP_CREATED"
	EvBackupRestored = "BACKUP_RESTORED"
	EvStateRead      = "STATE_READ"
	EvStateWrite     = "STATE_WRITE"
)

func LogEvent(code, message string, data map[string]interface{}) {
	if os.Getenv("LAZYPOOT_DEBUG") == "" {
		return
	}
	if data == nil {
		data = map[string]interface{}{}
	}
	var extra string
	if len(data) > 0 {
		var pairs []string
		for k, v := range data {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}
		extra = " " + strings.Join(pairs, " ")
	}
	fmt.Fprintf(os.Stderr, "[lazypoot] [%s] %s%s\n", code, message, extra)
}

func logDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = "/data/data/com.termux/files/home"
	}
	return filepath.Join(home, ".lazypoot", "log")
}

func ensureLogDir() error {
	return os.MkdirAll(logDir(), 0755)
}

func writeInstallLog(logs []types.LogEntry, distro string) (string, error) {
	if err := ensureLogDir(); err != nil {
		return "", err
	}
	ts := time.Now().Format("2006-01-02_15-04-05")
	name := fmt.Sprintf("%s_%s.log", distro, ts)
	path := filepath.Join(logDir(), name)
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("LazyPoot install log — %s\n", time.Now().Format(time.RFC1123)))
	sb.WriteString(fmt.Sprintf("Distro: %s\n", distro))
	sb.WriteString(strings.Repeat("─", 60) + "\n")
	for _, e := range logs {
		sb.WriteString(fmt.Sprintf("[%s] %s\n", string(e.Level), e.Message))
	}
	return path, os.WriteFile(path, []byte(sb.String()), 0600)
}

func readLogFiles() []string {
	dir := logDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".log") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files
}

func readLatestLog() []string {
	files := readLogFiles()
	if len(files) == 0 {
		return nil
	}
	last := files[len(files)-1]
	data, err := os.ReadFile(last)
	if err != nil {
		return nil
	}
	return strings.Split(string(data), "\n")
}
