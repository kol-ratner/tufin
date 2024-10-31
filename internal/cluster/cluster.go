package cluster

import (
	"bufio"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

//go:embed bin/k3d-darwin-arm64
var k3dDarwinArm64 []byte

//go:embed bin/k3d-darwin-amd64
var k3dDarwinAmd64 []byte

//go:embed bin/k3d-linux-amd64
var k3dLinuxAmd64 []byte

//go:embed bin/k3d-linux-arm64
var k3dLinuxArm64 []byte

type k3dHostInfo struct {
	os   string
	arch string
	bin  *os.File
}

// getK3d provides info on the host's runtime env and based on that,
// selects a k3d binary that is appropriate for the host
func getK3d() (*k3dHostInfo, error) {
	var k3dBinary []byte

	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			k3dBinary = k3dDarwinArm64
		} else {
			k3dBinary = k3dDarwinAmd64
		}
	case "linux":
		if runtime.GOARCH == "arm64" {
			k3dBinary = k3dLinuxArm64
		} else {
			k3dBinary = k3dLinuxAmd64
		}
	default:
		return nil, errors.New("your operating system is not currently supported")
	}

	tmpFile, err := os.CreateTemp("", "k3d-*")
	if err != nil {
		return nil, err
	}

	// Write binary content and make executable
	if _, err := tmpFile.Write(k3dBinary); err != nil {
		return nil, err
	}
	if err := tmpFile.Chmod(0755); err != nil {
		return nil, err
	}
	defer tmpFile.Close()

	return &k3dHostInfo{
		os:   runtime.GOOS,
		arch: runtime.GOARCH,
		bin:  tmpFile,
	}, nil
}

// parseK3dOutput parses the output of the k3d command and returns the message as a string
// the trick here is to strip the ansi color code from the k3d output so that we can also
// strip the formatting of the k3d logger - this gives me complete control over the log output for this application
func parseK3dOutput(line string) string {
	// Strip ANSI color codes
	var stripAnsi *regexp.Regexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	cleanLine := stripAnsi.ReplaceAllString(line, "")

	var infoRegex *regexp.Regexp = regexp.MustCompile(`^INFO\[\d+\]\s+(.+)$`)
	var fataRegex *regexp.Regexp = regexp.MustCompile(`^FATA\[\d+\]\s+(.+)$`)

	if matches := infoRegex.FindStringSubmatch(cleanLine); matches != nil {
		return matches[1]
	}

	if matches := fataRegex.FindStringSubmatch(cleanLine); matches != nil {
		return matches[1]
	}
	// Default to stdout for other messages
	return line
}

// clusterExists checks if a k3d cluster exists
func clusterExists(bin, clusterName string) (bool, error) {
	cmd := exec.Command(bin, "cluster", "list", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	var clusters []struct{ Name string }
	if err := json.Unmarshal(output, &clusters); err != nil {
		return false, err
	}

	for _, c := range clusters {
		if c.Name == clusterName {
			return true, nil
		}
	}
	return false, nil
}

// create creates a k3d cluster
func Create(msgChan chan<- string) error {
	k3d, err := getK3d()
	if err != nil {
		return err
	}
	defer os.Remove(k3d.bin.Name())
	msgChan <- fmt.Sprintf("Detected OS: %s, ARCH: %s", k3d.os, k3d.arch)

	if exists, err := clusterExists(k3d.bin.Name(), "k3s-default"); err != nil {
		return err
	} else if exists {
		msgChan <- "cluster already exists, skipping creation"
		return nil
	}

	command := exec.Command(k3d.bin.Name(), "cluster", "create")
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		// defer close(msgChan)
		return err
	}

	// writing to the channel via these goroutines is necessary
	// because of how exec.Command's StdoutPipe() and StderrPipe() work..
	// they need to be continuously read to prevent the command from blocking
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			text := scanner.Text()

			// only sending specific status updates
			if strings.Contains(text, "Starting cluster") {
				msgChan <- parseK3dOutput(text)
			} else if strings.Contains(text, "created successfully") {
				msgChan <- parseK3dOutput(text)
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			text := scanner.Text()

			// only sending specific status updates
			if strings.Contains(text, "Failed") {
				msgChan <- parseK3dOutput(text)
			}
		}
	}()

	return command.Wait()
}
