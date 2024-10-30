package cluster

import (
	"bytes"
	_ "embed"
	"errors"
	"log"
	"os"
	"os/exec"
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

func Create() error {
	var k3dBinary []byte

	log.Println("Detecting your operating system...")
	log.Printf("Detected OS: %s, ARCH: %s \n", runtime.GOOS, runtime.GOARCH)

	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			k3dBinary = k3dDarwinArm64
		}
		k3dBinary = k3dDarwinAmd64
	case "linux":
		if runtime.GOARCH == "arm64" {
			k3dBinary = k3dLinuxArm64
		}
		k3dBinary = k3dLinuxAmd64
	default:
		return errors.New("Your operating system is not currently supported")
	}

	tmpFile, err := os.CreateTemp("", "k3d-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// Write binary content and make executable
	if _, err := tmpFile.Write(k3dBinary); err != nil {
		return err
	}
	if err := tmpFile.Chmod(0755); err != nil {
		return err
	}
	tmpFile.Close()

	// Execute k3d cluster create
	command := exec.Command(tmpFile.Name(), "cluster", "create")
	var stderr bytes.Buffer
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		errorMsg := stderr.String()
		// Find everything after the first ]
		if idx := strings.Index(errorMsg, "]"); idx != -1 {
			cleanError := strings.TrimSpace(errorMsg[idx+1:])
			return errors.New(cleanError)
		}
		// log.Fatal(strings.Split(stderr.String(), " ")[0])
		// log.Fatal(err)
	}

	return nil
}
