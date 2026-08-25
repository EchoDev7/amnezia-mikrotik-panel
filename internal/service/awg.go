package service

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GeneratePrivateKey generates a new WireGuard/AmneziaWG private key
func GeneratePrivateKey() (string, error) {
	cmd := exec.Command("awg", "genkey")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %v", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// GeneratePublicKey generates a public key from a given private key
func GeneratePublicKey(privateKey string) (string, error) {
	cmd := exec.Command("awg", "pubkey")
	cmd.Stdin = strings.NewReader(privateKey)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to generate public key: %v", err)
	}
	return strings.TrimSpace(out.String()), nil
}

// AddPeer adds a peer to the given interface
func AddPeer(interfaceName, publicKey, allowedIPs string) error {
	cmd := exec.Command("awg", "set", interfaceName, "peer", publicKey, "allowed-ips", allowedIPs)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to add peer: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

// RemovePeer removes a peer from the given interface
func RemovePeer(interfaceName, publicKey string) error {
	cmd := exec.Command("awg", "set", interfaceName, "peer", publicKey, "remove")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to remove peer: %v, stderr: %s", err, stderr.String())
	}
	return nil
}

// ShowDump returns the output of `awg show <interface> dump`
func ShowDump(interfaceName string) (string, error) {
	cmd := exec.Command("awg", "show", interfaceName, "dump")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to dump interface stats: %v, stderr: %s", err, stderr.String())
	}
	return out.String(), nil
}

// GetServerPublicKey returns the public key of the server interface
func GetServerPublicKey(interfaceName string) (string, error) {
	cmd := exec.Command("awg", "show", interfaceName, "public-key")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to get server public key: %v, stderr: %s", err, stderr.String())
	}
	return strings.TrimSpace(out.String()), nil
}
