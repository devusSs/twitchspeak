package system

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Gets the operating system
func GetOS(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "uname", "-s")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting os: %w", err)
	}
	osName := strings.TrimSpace(string(output))
	return osName, nil
}

// Gets the architecture
func GetArch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "uname", "-m")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("getting arch: %w", err)
	}
	arch := strings.TrimSpace(string(output))
	return arch, nil
}
