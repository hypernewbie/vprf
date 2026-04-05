package profile

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Load(path string) (*Profile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := any(file)
	var dec *json.Decoder
	if strings.HasSuffix(strings.ToLower(path), ".gz") {
		gz, err := gzip.NewReader(file)
		if err != nil {
			return nil, fmt.Errorf("open gzip profile: %w", err)
		}
		defer gz.Close()
		dec = json.NewDecoder(gz)
	} else {
		_ = reader
		dec = json.NewDecoder(file)
	}

	var profile Profile
	if err := dec.Decode(&profile); err != nil {
		return nil, fmt.Errorf("decode profile %s: %w", filepath.Base(path), err)
	}
	if sidecar, err := loadSidecar(path); err == nil && sidecar != nil {
		attachSidecarSymbols(&profile, sidecar)
	}
	return &profile, nil
}

func loadSidecar(profilePath string) (*SymbolSidecar, error) {
	sidecarPath := strings.TrimSuffix(profilePath, ".gz") + ".syms.json"
	file, err := os.Open(sidecarPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var sidecar SymbolSidecar
	if err := json.NewDecoder(file).Decode(&sidecar); err != nil {
		return nil, err
	}
	return &sidecar, nil
}
