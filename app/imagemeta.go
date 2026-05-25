package app

import (
	"encoding/json"
	"os"
)

type ImageMeta struct {
	ID        string `json:"id"`
	Variant   string `json:"variant"`
	Arch      string `json:"arch"`
	Installed int64  `json:"installed_at"`
	State     string `json:"state"`
}

type ImageInstallDoneMsg struct {
	Distro      string
	DistroID    string
	RootfsDir   string
	VNCPassword string
}

func writeMeta(path string, meta ImageMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func ReadMeta(path string) (*ImageMeta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var meta ImageMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}
