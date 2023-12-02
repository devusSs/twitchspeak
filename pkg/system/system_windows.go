package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Gets the operating system
func GetOS(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "ver")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting os: %w", err)
	}
	osVersion := strings.TrimSpace(string(output))
	if strings.Contains(osVersion, "Windows") {
		osVersion = "Windows"
	}
	return osVersion, nil
}

// Gets the architecture
func GetArch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/c", "wmic", "os", "get", "osarchitecture")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting arch: %w", err)
	}
	arch := strings.Replace(strings.TrimSpace(string(output)), "OSArchitecture", "", 1)
	if strings.Contains(arch, "64") {
		return "x86_64", nil
	}
	return arch, nil
}
