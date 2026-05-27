package app

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)


func LockPath(home string) string {
	return filepath.Join(home, ".lazypoot", ".lock")
}

func ManifestPath(home string) string {
	return filepath.Join(home, ".lazypoot", "portal", "manifest")
}

func StatePath(home string) string {
	return filepath.Join(home, ".lazypoot", "state.json")
}

func BackupDir(home string) string {
	return filepath.Join(home, ".lazypoot", "backup")
}

func binaryHash() string {
	exe, err := os.Executable()
	if err != nil {
		return "unknown"
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		return "unknown"
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func AcquireLock(home string) (release func(), err error) {
	path := LockPath(home)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err == nil {
		li, parseErr := ParseLockInfo(string(data))
		if parseErr == nil && li.PID > 0 {
			proc, _ := os.FindProcess(li.PID)
			if proc != nil && proc.Signal(syscall.Signal(0)) == nil {
				return nil, fmt.Errorf("lock held by process %d (pid %d, started %s)",
					li.PID, li.PID, li.StartTime)
			}
		}
	}

	pid := os.Getpid()
	now := time.Now().UTC().Format(time.RFC3339)
	binHash := binaryHash()
	lockData := FormatLockInfo(pid, now, binHash)

	if err := os.WriteFile(path, []byte(lockData+"\n"), 0644); err != nil {
		return nil, err
	}

	var released bool
	release = func() {
		if !released {
			released = true
			os.Remove(path)
		}
	}
	return release, nil
}

func ReadManifest(home string) []string {
	data, err := os.ReadFile(ManifestPath(home))
	if err != nil {
		return nil
	}
	return ParseManifest(string(data))
}

func WriteManifest(home string, ids []string) error {
	path := ManifestPath(home)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data := FormatManifest(ids)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(data), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}


func ReadState(home string) *PortalState {
	data, err := os.ReadFile(StatePath(home))
	if err != nil {
		return &PortalState{DistroIDs: []string{}, HookVersion: PortalVersion}
	}
	s, err := ParseState(string(data))
	if err != nil {
		return &PortalState{DistroIDs: []string{}, HookVersion: PortalVersion}
	}
	return s
}

func WriteState(home string, s *PortalState) error {
	path := StatePath(home)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(FormatState(s)), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func SyncStateFromManifest(home string) error {
	ids := ReadManifest(home)
	s := ReadState(home)
	if !stringSliceEqual(s.DistroIDs, ids) {
		s.DistroIDs = ids
		s.HookVersion = PortalVersion
		s.LastSyncHash = ComputeSyncHash(ids)
		return WriteState(home, s)
	}
	return nil
}


func BackupRCFiles(home string) (rollback func(), err error) {
	dir := BackupDir(home)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	type entry struct {
		path   string
		backup string
	}
	var saved []entry

	for _, t := range RCTargets(home) {
		data, rerr := os.ReadFile(t.Path)
		if rerr != nil {
			continue
		}
		name := filepath.Base(t.Path) + "." + strconv.FormatInt(time.Now().UnixNano(), 36)
		bp := filepath.Join(dir, name)
		if rerr := os.WriteFile(bp, data, 0644); rerr != nil {
			return nil, rerr
		}
		saved = append(saved, entry{t.Path, bp})
	}

	var restored bool
	rollback = func() {
		if restored {
			return
		}
		restored = true
		for _, e := range saved {
			data, rerr := os.ReadFile(e.backup)
			if rerr == nil {
				os.WriteFile(e.path, data, 0644)
			}
		}
	}
	return rollback, nil
}


func SyncPortalHooks(home string) error {
	ids := ReadManifest(home)
	for _, t := range RCTargets(home) {
		block := t.BlockFn(ids)
		if err := os.MkdirAll(filepath.Dir(t.Path), 0755); err != nil {
			return err
		}
		data, err := os.ReadFile(t.Path)
		if err != nil {
			data = []byte{}
		}
		out := UpdateRCFile(string(data), block)
		if ContentDiff(string(data), out) {
			if err := os.WriteFile(t.Path, []byte(out), 0644); err != nil {
				return err
			}
		}
	}
	if err := SyncStateFromManifest(home); err != nil {
		LogEvent("STATE_SYNC_ERR", "failed to sync state", map[string]interface{}{"error": err.Error()})
	}
	return nil
}

func SyncDryRun(home string) (string, error) {
	var report strings.Builder
	ids := ReadManifest(home)
	report.WriteString(fmt.Sprintf("Manifest: %d distro(s)\n", len(ids)))
	for _, id := range ids {
		report.WriteString(fmt.Sprintf("  - %s\n", id))
	}
	state := ReadState(home)
	report.WriteString(fmt.Sprintf("Hook version: %d\n", state.HookVersion))
	report.WriteString(fmt.Sprintf("Last sync hash: %s\n", state.LastSyncHash))

	for _, t := range RCTargets(home) {
		block := t.BlockFn(ids)
		data, err := os.ReadFile(t.Path)
		if err != nil {
			data = []byte{}
		}
		out := UpdateRCFile(string(data), block)
		if ContentDiff(string(data), out) {
			report.WriteString(fmt.Sprintf("\n[CHANGED] %s\n", t.Path))
		} else {
			report.WriteString(fmt.Sprintf("\n[OK] %s (no change)\n", t.Path))
		}
	}
	return report.String(), nil
}


func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
