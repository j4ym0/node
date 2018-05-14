package openvpn

import (
	"bufio"
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"strings"
)

// ValidateConfig is function which takes VPNConfig as argument, checks it and returns error if validation fails
type ValidateConfig func(config *VPNConfig) error

// ConfigValidator represents structure which contains list of validating functions
type ConfigValidator struct {
	validators []ValidateConfig
}

// NewDefaultValidator returns ConfigValidator with predefined list of validating functions
func NewDefaultValidator() *ConfigValidator {
	return &ConfigValidator{
		validators: []ValidateConfig{
			validProtocol,
			validPort,
			validIPFormat,
			validTLSPresharedKey,
			validCACertificate,
		},
	}
}

// IsValid function checks if provided config is valid against given config validator and returns first encountered error
func (v *ConfigValidator) IsValid(config *VPNConfig) error {
	for _, validator := range v.validators {
		if err := validator(config); err != nil {
			return err
		}
	}
	return nil
}

func validProtocol(config *VPNConfig) error {
	switch config.RemoteProtocol {
	case
		"udp",
		"tcp":
		return nil
	}
	return errors.New("invalid protocol: " + config.RemoteProtocol)
}

func validPort(config *VPNConfig) error {
	if config.RemotePort > 65535 || config.RemotePort < 1024 {
		return errors.New("invalid port range, should fall within 1024 .. 65535 range")
	}
	return nil
}

func validIPFormat(config *VPNConfig) error {
	parsed := net.ParseIP(config.RemoteIP)
	if parsed == nil {
		return errors.New("unable to parse ip address " + config.RemoteIP)
	}
	if parsed.To4() == nil {
		return errors.New("IPv4 address is expected")
	}
	return nil
}

// preshared key format (PEM blocks with data encoded to hex) are taken from
// openvpn --genkey --secret static.key, which is openvpn specific
func validTLSPresharedKey(config *VPNConfig) error {
	contentScanner := bufio.NewScanner(bytes.NewBufferString(config.TLSPresharedKey))
	for contentScanner.Scan() {
		line := contentScanner.Text()
		//skip empty lines or comments
		if len(line) > 0 || strings.HasPrefix(line, "#") {
			break
		}
	}
	if err := contentScanner.Err(); err != nil {
		return contentScanner.Err()
	}
	header := contentScanner.Text()
	if header != "-----BEGIN OpenVPN Static key V1-----" {
		return errors.New("Invalid key header: " + header)
	}

	var key string
	for contentScanner.Scan() {
		line := contentScanner.Text()
		if line == "-----END OpenVPN Static key V1-----" {
			break
		} else {
			key = key + line
		}
	}
	if err := contentScanner.Err(); err != nil {
		return err
	}
	// 256 bytes key is 512 bytes if encoded to hex
	if len(key) != 512 {
		return errors.New("invalid key length")
	}
	return nil
}

func validCACertificate(config *VPNConfig) error {
	pemBlock, _ := pem.Decode([]byte(config.CACertificate))
	if pemBlock.Type != "CERTIFICATE" {
		return errors.New("invalid CA certificate. Certificate block expected")
	}
	//if we parse it correctly - at least structure is right
	_, err := x509.ParseCertificate(pemBlock.Bytes)
	if err != nil {
		return err
	}
	return nil
}
