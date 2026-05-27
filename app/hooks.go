package app

import (
	"fmt"
	"os"
	"path/filepath"
)

func SetupPortalLogin(home, distroID string, usernames []string) error {
	path, err := generatePortalScript(home, distroID, usernames)
	if err != nil {
		LogEvent("ERR_PORTAL_SCRIPT", "failed to generate portal script", map[string]interface{}{
			"distroID": distroID, "error": err.Error(),
		})
		return err
	}
	LogEvent("PORTAL_SCRIPT_OK", "portal script generated", map[string]interface{}{
		"distroID": distroID, "path": path,
	})
	return InjectPortalHooks(home, distroID)
}

func InjectPortalHooks(home, distroID string) (err error) {
	LogEvent("HOOK_INJECT_START", "injecting hooks", map[string]interface{}{
		"distroID": distroID,
	})

	release, err := AcquireLock(home)
	if err != nil {
		LogEvent(EvLockHeld, err.Error(), nil)
		return fmt.Errorf("lock: %w", err)
	}
	defer release()
	LogEvent(EvLockAcquired, "lock acquired", map[string]interface{}{"distroID": distroID})

	rollback, err := BackupRCFiles(home)
	if err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			rollback()
			LogEvent(EvBackupRestored, "backup restored after panic", nil)
			err = fmt.Errorf("panic in InjectPortalHooks: %v", r)
		}
		if err != nil {
			rollback()
			LogEvent(EvBackupRestored, "backup restored after error", map[string]interface{}{"error": err.Error()})
		}
	}()

	ids := ReadManifest(home)
	found := false
	for _, id := range ids {
		if id == distroID {
			found = true
			break
		}
	}
	if !found {
		ids = append(ids, distroID)
	}

	if err := WriteManifest(home, ids); err != nil {
		LogEvent(EvManifestWrite, "manifest write failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("manifest: %w", err)
	}
	LogEvent(EvManifestWrite, "manifest written", map[string]interface{}{"distros": len(ids)})

	if err := SyncPortalHooks(home); err != nil {
		LogEvent(EvHookWriteErr, "hook sync failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("sync: %w", err)
	}
	LogEvent(EvHookWriteOK, "hooks synced", map[string]interface{}{"distroID": distroID})

	return nil
}

func RemovePortalHooks(home, distroID string) (err error) {
	LogEvent("HOOK_REMOVE_START", "removing hooks", map[string]interface{}{
		"distroID": distroID,
	})

	release, err := AcquireLock(home)
	if err != nil {
		LogEvent(EvLockHeld, err.Error(), nil)
		return fmt.Errorf("lock: %w", err)
	}
	defer release()
	LogEvent(EvLockAcquired, "lock acquired", map[string]interface{}{"distroID": distroID})

	rollback, err := BackupRCFiles(home)
	if err != nil {
		return fmt.Errorf("backup: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			rollback()
			LogEvent(EvBackupRestored, "backup restored after panic", nil)
			err = fmt.Errorf("panic in RemovePortalHooks: %v", r)
		}
		if err != nil {
			rollback()
			LogEvent(EvBackupRestored, "backup restored after error", map[string]interface{}{"error": err.Error()})
		}
	}()

	ids := ReadManifest(home)
	var kept []string
	for _, id := range ids {
		if id != distroID {
			kept = append(kept, id)
		}
	}

	if err := WriteManifest(home, kept); err != nil {
		LogEvent(EvManifestWrite, "manifest write failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("manifest: %w", err)
	}

	if err := SyncPortalHooks(home); err != nil {
		LogEvent(EvHookWriteErr, "hook sync failed", map[string]interface{}{"error": err.Error()})
		return fmt.Errorf("sync: %w", err)
	}
	LogEvent(EvHookWriteOK, "hooks synced after removal", map[string]interface{}{"distroID": distroID})

	script := filepath.Join(home, ".lazypoot", "portal", distroID+".sh")
	os.Remove(script)
	LogEvent("PORTAL_SCRIPT_REMOVED", "portal script removed", map[string]interface{}{"path": script})

	return nil
}
