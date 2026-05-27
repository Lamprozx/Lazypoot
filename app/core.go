package app

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)


var distroIDRe = regexp.MustCompile(`^[a-z][a-z0-9-]{0,30}$`)

const (
	PortalBlockStart = "# --- LazyPoot Portal START ---"
	PortalBlockEnd   = "# --- LazyPoot Portal END ---"
	PortalVersion    = 3
)


func BlockChecksum(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func BinaryChecksum(bin []byte) string {
	h := sha256.Sum256(bin)
	return hex.EncodeToString(h[:])

}


func ParseManifest(data string) []string {
	var ids []string
	for _, line := range strings.Split(data, "\n") {
		id := strings.TrimSpace(line)
		if id != "" && distroIDRe.MatchString(id) {
			ids = append(ids, id)
		}
	}
	return ids
}

func FormatManifest(ids []string) string {
	return strings.Join(ids, "\n") + "\n"
}


type PortalState struct {
	DistroIDs    []string `json:"distroIDs"`
	HookVersion  int      `json:"hookVersion"`
	LastSyncHash string   `json:"lastSyncHash"`
}

func ParseState(data string) (*PortalState, error) {
	var s PortalState
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func FormatState(s *PortalState) string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

func ComputeSyncHash(ids []string) string {
	return BlockChecksum(strings.Join(ids, ","))
}


func GenerateSHBlock(ids []string) string {
	var buf strings.Builder
	buf.WriteString(PortalBlockStart + "\n")
	buf.WriteString(fmt.Sprintf("# lazypoot_version=%d\n", PortalVersion))

	var content strings.Builder
	for _, id := range ids {
		rel := filepath.Join("$HOME", ".lazypoot", "portal", id+".sh")
		content.WriteString(fmt.Sprintf("# LazyPoot Portal: %s\n%s(){ \"%s\" \"$@\"; }\n", id, id, rel))
	}

	buf.WriteString(fmt.Sprintf("# block_version=%s\n", BlockChecksum(content.String())))
	buf.WriteString(content.String())
	buf.WriteString(PortalBlockEnd + "\n")
	return buf.String()
}

func GenerateFishBlock(ids []string) string {
	var buf strings.Builder
	buf.WriteString(PortalBlockStart + "\n")
	buf.WriteString(fmt.Sprintf("# lazypoot_version=%d\n", PortalVersion))

	var content strings.Builder
	for _, id := range ids {
		rel := filepath.Join("$HOME", ".lazypoot", "portal", id+".sh")
		content.WriteString(fmt.Sprintf("# LazyPoot Portal: %s\nfunction %s\n    \"%s\" $argv\nend\n", id, id, rel))
	}

	buf.WriteString(fmt.Sprintf("# block_version=%s\n", BlockChecksum(content.String())))
	buf.WriteString(content.String())
	buf.WriteString(PortalBlockEnd + "\n")
	return buf.String()
}


func UpdateRCFile(content, block string) string {
	start := strings.Index(content, PortalBlockStart)
	if start < 0 {
		return strings.TrimRight(content, "\n") + "\n" + block
	}
	lineStart := start
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}
	return strings.TrimRight(content[:lineStart], "\n") + "\n" + block
}


func ContentDiff(oldContent, newContent string) bool {
	return oldContent != newContent
}


type LockInfo struct {
	PID        int    `json:"pid"`
	StartTime  string `json:"startTime"`
	BinaryHash string `json:"binaryHash"`
}

func FormatLockInfo(pid int, startTime, binHash string) string {
	li := LockInfo{PID: pid, StartTime: startTime, BinaryHash: binHash}
	data, _ := json.Marshal(li)
	return string(data)
}

func ParseLockInfo(data string) (*LockInfo, error) {
	var li LockInfo
	if err := json.Unmarshal([]byte(data), &li); err != nil {
		return nil, err
	}
	return &li, nil
}
