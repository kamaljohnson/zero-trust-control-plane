package security

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadPEM_InlinePEM(t *testing.T) {
	inlinePEM := testPrivateKeyPEM
	pemBytes, err := LoadPEM(inlinePEM)
	if err != nil {
		t.Fatalf("LoadPEM: %v", err)
	}
	if len(pemBytes) == 0 {
		t.Error("LoadPEM returned empty bytes")
	}
	// Check that it contains PEM markers
	pemStr := string(pemBytes)
	if !strings.Contains(pemStr, "-----BEGIN") {
		t.Error("LoadPEM did not return PEM content")
	}
}

func TestLoadPEM_InlinePEMWithLiteralNewlines(t *testing.T) {
	// Test that literal \n is converted to actual newlines
	inlinePEM := "-----BEGIN PRIVATE KEY-----\\nMII...\\n-----END PRIVATE KEY-----"
	pemBytes, err := LoadPEM(inlinePEM)
	if err != nil {
		// This might fail if not valid PEM, but should handle \n conversion
		if err == ErrInvalidKey {
			// Expected for invalid PEM content
			return
		}
		t.Fatalf("LoadPEM: %v", err)
	}
	// Check that \n was converted
	if len(pemBytes) > 0 {
		// Should have actual newlines, not literal \n
		hasNewline := false
		for _, b := range pemBytes {
			if b == '\n' {
				hasNewline = true
				break
			}
		}
		if !hasNewline {
			t.Error("LoadPEM should convert \\n to actual newlines")
		}
	}
}

func TestLoadPEM_FilePath(t *testing.T) {
	// Create a temporary file with PEM content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.pem")
	if err := os.WriteFile(tmpFile, []byte(testPrivateKeyPEM), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	pemBytes, err := LoadPEM(tmpFile)
	if err != nil {
		t.Fatalf("LoadPEM: %v", err)
	}
	if len(pemBytes) == 0 {
		t.Error("LoadPEM returned empty bytes")
	}
	pemStr := string(pemBytes)
	if !strings.Contains(pemStr, "-----BEGIN") {
		t.Error("LoadPEM did not read file content")
	}
}

func TestLoadPEM_EmptyString(t *testing.T) {
	_, err := LoadPEM("")
	if err != ErrInvalidKey {
		t.Errorf("LoadPEM empty string: want ErrInvalidKey, got %v", err)
	}
}

func TestLoadPEM_WhitespaceOnly(t *testing.T) {
	_, err := LoadPEM("   ")
	if err != ErrInvalidKey {
		t.Errorf("LoadPEM whitespace: want ErrInvalidKey, got %v", err)
	}
}

func TestLoadPEM_InvalidFile(t *testing.T) {
	_, err := LoadPEM("/nonexistent/file.pem")
	if err == nil {
		t.Error("LoadPEM should return error for nonexistent file")
	}
}

func TestParsePrivateKey_RSA(t *testing.T) {
	key, err := ParsePrivateKey(testPrivateKeyPEM)
	if err != nil {
		t.Fatalf("ParsePrivateKey: %v", err)
	}
	if key == nil {
		t.Fatal("ParsePrivateKey returned nil key")
	}
}

func TestParsePrivateKey_InvalidPEM(t *testing.T) {
	// Use a string that looks like inline PEM but is invalid
	invalidPEM := "-----BEGIN PRIVATE KEY-----\ninvalid\n-----END PRIVATE KEY-----"
	_, err := ParsePrivateKey(invalidPEM)
	if err == nil {
		t.Error("ParsePrivateKey should return error for invalid PEM")
	}
	// Error should be ErrInvalidKey or from parsing
	if err != ErrInvalidKey && !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "asn1") {
		t.Errorf("ParsePrivateKey invalid PEM: want ErrInvalidKey or parsing error, got %v", err)
	}
}

func TestParsePrivateKey_InvalidKeyType(t *testing.T) {
	invalidPEM := `-----BEGIN CERTIFICATE-----
MII...
-----END CERTIFICATE-----`
	_, err := ParsePrivateKey(invalidPEM)
	if err == nil {
		t.Error("ParsePrivateKey should return error for non-key PEM")
	}
}

func TestParsePublicKey_RSA(t *testing.T) {
	key, err := ParsePublicKey(testPublicKeyPEM)
	if err != nil {
		t.Fatalf("ParsePublicKey: %v", err)
	}
	if key == nil {
		t.Fatal("ParsePublicKey returned nil key")
	}
}

func TestParsePublicKey_InvalidPEM(t *testing.T) {
	// Use a string that looks like inline PEM but is invalid
	invalidPEM := "-----BEGIN PUBLIC KEY-----\ninvalid\n-----END PUBLIC KEY-----"
	_, err := ParsePublicKey(invalidPEM)
	if err == nil {
		t.Error("ParsePublicKey should return error for invalid PEM")
	}
	// Error should be ErrInvalidKey or from parsing
	if err != ErrInvalidKey && !strings.Contains(err.Error(), "invalid") && !strings.Contains(err.Error(), "asn1") {
		t.Errorf("ParsePublicKey invalid PEM: want ErrInvalidKey or parsing error, got %v", err)
	}
}

func TestKeyAlg_RSA(t *testing.T) {
	pub, err := ParsePublicKey(testPublicKeyPEM)
	if err != nil {
		t.Fatalf("ParsePublicKey: %v", err)
	}
	alg := KeyAlg(pub)
	if alg != "RS256" {
		t.Errorf("KeyAlg RSA: want RS256, got %q", alg)
	}
}

func TestKeyAlg_Unsupported(t *testing.T) {
	// Test with nil or unsupported key type
	alg := KeyAlg(nil)
	if alg != "" {
		t.Errorf("KeyAlg nil: want empty string, got %q", alg)
	}
}

func TestParsePrivateKey_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "key.pem")
	if err := os.WriteFile(tmpFile, []byte(testPrivateKeyPEM), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	key, err := ParsePrivateKey(tmpFile)
	if err != nil {
		t.Fatalf("ParsePrivateKey with file: %v", err)
	}
	if key == nil {
		t.Fatal("ParsePrivateKey returned nil key")
	}
}

func TestParsePublicKey_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "pub.pem")
	if err := os.WriteFile(tmpFile, []byte(testPublicKeyPEM), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	key, err := ParsePublicKey(tmpFile)
	if err != nil {
		t.Fatalf("ParsePublicKey with file: %v", err)
	}
	if key == nil {
		t.Fatal("ParsePublicKey returned nil key")
	}
}

// Key Parsing Edge Case Tests

func TestParsePrivateKey_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name string
		pem  string
	}{
		{"not PEM format", "not a pem format"},
		{"missing BEGIN marker", "-----END PRIVATE KEY-----\ncontent\n-----END PRIVATE KEY-----"},
		{"missing END marker", "-----BEGIN PRIVATE KEY-----\ncontent"},
		{"empty PEM block", "-----BEGIN PRIVATE KEY-----\n-----END PRIVATE KEY-----"},
		{"invalid base64", "-----BEGIN PRIVATE KEY-----\n!!!invalid base64!!!\n-----END PRIVATE KEY-----"},
		{"file path that doesn't exist", "/nonexistent/private_key.pem"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePrivateKey(tc.pem)
			if err == nil {
				t.Errorf("ParsePrivateKey %q: want error, got nil", tc.name)
			}
		})
	}
}

func TestParsePrivateKey_WrongKeyType(t *testing.T) {
	// Test with a public key (wrong type for ParsePrivateKey)
	_, err := ParsePrivateKey(testPublicKeyPEM)
	if err == nil {
		t.Error("ParsePrivateKey with public key: want error, got nil")
	}
	if err != ErrInvalidKey && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("ParsePrivateKey wrong type: want ErrInvalidKey or parsing error, got %v", err)
	}

	// Test with certificate (wrong type)
	certPEM := `-----BEGIN CERTIFICATE-----
MIIC...
-----END CERTIFICATE-----`
	_, err = ParsePrivateKey(certPEM)
	if err == nil {
		t.Error("ParsePrivateKey with certificate: want error, got nil")
	}
}

func TestParsePublicKey_InvalidFormat(t *testing.T) {
	testCases := []struct {
		name string
		pem  string
	}{
		{"not PEM format", "not a pem format"},
		{"missing BEGIN marker", "-----END PUBLIC KEY-----\ncontent\n-----END PUBLIC KEY-----"},
		{"missing END marker", "-----BEGIN PUBLIC KEY-----\ncontent"},
		{"empty PEM block", "-----BEGIN PUBLIC KEY-----\n-----END PUBLIC KEY-----"},
		{"invalid base64", "-----BEGIN PUBLIC KEY-----\n!!!invalid base64!!!\n-----END PUBLIC KEY-----"},
		{"file path that doesn't exist", "/nonexistent/public_key.pem"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParsePublicKey(tc.pem)
			if err == nil {
				t.Errorf("ParsePublicKey %q: want error, got nil", tc.name)
			}
		})
	}
}

func TestParsePublicKey_WrongKeyType(t *testing.T) {
	// Test with a private key (wrong type for ParsePublicKey)
	_, err := ParsePublicKey(testPrivateKeyPEM)
	if err == nil {
		t.Error("ParsePublicKey with private key: want error, got nil")
	}
	if err != ErrInvalidKey && !strings.Contains(err.Error(), "invalid") {
		t.Errorf("ParsePublicKey wrong type: want ErrInvalidKey or parsing error, got %v", err)
	}

	// Test with certificate (wrong type)
	certPEM := `-----BEGIN CERTIFICATE-----
MIIC...
-----END CERTIFICATE-----`
	_, err = ParsePublicKey(certPEM)
	if err == nil {
		t.Error("ParsePublicKey with certificate: want error, got nil")
	}
}
