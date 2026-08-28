package secrets

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sort"
	"strings"
	"testing"
)

const testRootSecret = "0123456789abcdef0123456789abcdef"

func TestPersistentDerivationIsDeterministicAndComplete(t *testing.T) {
	first, err := Derive(testRootSecret)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Derive(testRootSecret)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range Names {
		if first[name] == "" {
			t.Errorf("%s is empty", name)
		}
		if first[name] != second[name] {
			t.Errorf("%s is not deterministic", name)
		}
	}
	if len(first) != len(Names) {
		t.Fatalf("got %d credentials, want %d", len(first), len(Names))
	}

	// Locks the version-one derivation contract. An intentional format or label
	// change must use a new derivation version because it rotates credentials.
	names := append([]string(nil), Names...)
	sort.Strings(names)
	var canonical strings.Builder
	for _, name := range names {
		canonical.WriteString(name + "=" + first[name] + "\n")
	}
	fingerprint := sha256.Sum256([]byte(canonical.String()))
	if got, want := base64URL(fingerprint[:]), "CCYIpAzelYFOFxscaxlTSyQt4ku7PSKFrm1qnPuD8UY"; got != want {
		t.Errorf("derivation fingerprint = %s, want %s", got, want)
	}
}

func TestDerivedLegacyJWTsVerify(t *testing.T) {
	values, err := Derive(testRootSecret)
	if err != nil {
		t.Fatal(err)
	}
	verifyHS256(t, values["ANON_KEY"], values["JWT_SECRET"], "anon")
	verifyHS256(t, values["SERVICE_ROLE_KEY"], values["JWT_SECRET"], "service_role")
}

func TestMintedGatewayJWTsVerifyAndAreNotStable(t *testing.T) {
	values, err := Derive(testRootSecret)
	if err != nil {
		t.Fatal(err)
	}
	first, err := MintGatewayTokens(testRootSecret)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MintGatewayTokens(testRootSecret)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range GatewayNames {
		if first[name] == "" || second[name] == "" {
			t.Fatalf("%s is empty", name)
		}
		if first[name] == second[name] {
			t.Errorf("%s unexpectedly has a stable randomized signature", name)
		}
	}
	publicKey := publicKeyFromJWKS(t, values["JWT_JWKS"])
	for _, minted := range []map[string]string{first, second} {
		verifyES256(t, minted["ANON_KEY_ASYMMETRIC"], publicKey, "anon")
		verifyES256(t, minted["SERVICE_ROLE_KEY_ASYMMETRIC"], publicKey, "service_role")
	}
}

func TestDerivedJWKShapesMatchSupabaseDocker(t *testing.T) {
	values, err := Derive(testRootSecret)
	if err != nil {
		t.Fatal(err)
	}

	var signingKeys []json.RawMessage
	if err := json.Unmarshal([]byte(values["JWT_KEYS"]), &signingKeys); err != nil {
		t.Fatalf("JWT_KEYS must be a JSON array: %v", err)
	}
	if len(signingKeys) != 2 {
		t.Fatalf("JWT_KEYS contains %d keys, want 2", len(signingKeys))
	}
	var privateJWK ecJWK
	if err := json.Unmarshal(signingKeys[0], &privateJWK); err != nil {
		t.Fatal(err)
	}
	if privateJWK.D == "" {
		t.Error("JWT_KEYS signing key is missing private component")
	}

	var publicKeys struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.Unmarshal([]byte(values["JWT_JWKS"]), &publicKeys); err != nil {
		t.Fatalf("JWT_JWKS must be a JSON object with keys: %v", err)
	}
	if len(publicKeys.Keys) != 2 {
		t.Fatalf("JWT_JWKS contains %d keys, want 2", len(publicKeys.Keys))
	}
	var publicJWK ecJWK
	if err := json.Unmarshal(publicKeys.Keys[0], &publicJWK); err != nil {
		t.Fatal(err)
	}
	if publicJWK.D != "" {
		t.Error("JWT_JWKS exposes private key component")
	}
	if privateJWK.Kid != publicJWK.Kid || privateJWK.X != publicJWK.X || privateJWK.Y != publicJWK.Y {
		t.Error("JWT_KEYS and JWT_JWKS contain different EC keys")
	}

	var symmetric octJWK
	if err := json.Unmarshal(signingKeys[1], &symmetric); err != nil {
		t.Fatal(err)
	}
	if symmetric.K != base64URL([]byte(values["JWT_SECRET"])) {
		t.Error("JWT_KEYS symmetric JWK does not encode JWT_SECRET")
	}
}

func TestDeriveAndMintRejectWeakRootSecret(t *testing.T) {
	if _, err := Derive("too short"); err == nil {
		t.Error("expected Derive to reject a short ROOT_SECRET")
	}
	if _, err := MintGatewayTokens("too short"); err == nil {
		t.Error("expected MintGatewayTokens to reject a short ROOT_SECRET")
	}
}

func publicKeyFromJWKS(t *testing.T, encoded string) *ecdsa.PublicKey {
	t.Helper()
	var keys struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.Unmarshal([]byte(encoded), &keys); err != nil {
		t.Fatal(err)
	}
	var jwk ecJWK
	if err := json.Unmarshal(keys.Keys[0], &jwk); err != nil {
		t.Fatal(err)
	}
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		t.Fatal(err)
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		t.Fatal(err)
	}
	encodedPublicKey := append([]byte{4}, xBytes...)
	encodedPublicKey = append(encodedPublicKey, yBytes...)
	publicKey, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), encodedPublicKey)
	if err != nil {
		t.Fatal(err)
	}
	return publicKey
}

func verifyHS256(t *testing.T, token, secret, role string) {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid JWT parts: %d", len(parts))
	}
	signature := mac([]byte(secret), []byte(parts[0]+"."+parts[1]))
	if base64URL(signature) != parts[2] {
		t.Error("invalid HS256 signature")
	}
	verifyRole(t, parts[1], role)
}

func mac(key []byte, parts ...[]byte) []byte {
	h := hmac.New(sha256.New, key)
	for _, part := range parts {
		h.Write(part)
	}
	return h.Sum(nil)
}

func verifyES256(t *testing.T, token string, publicKey *ecdsa.PublicKey, role string) {
	t.Helper()
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("invalid JWT parts: %d", len(parts))
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(signature) != 64 {
		t.Fatalf("invalid ES256 signature encoding")
	}
	digest := sha256.Sum256([]byte(parts[0] + "." + parts[1]))
	if !ecdsa.Verify(publicKey, digest[:], new(big.Int).SetBytes(signature[:32]), new(big.Int).SetBytes(signature[32:])) {
		t.Error("invalid ES256 signature")
	}
	verifyRole(t, parts[1], role)
}

func verifyRole(t *testing.T, encodedPayload, want string) {
	t.Helper()
	payload, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		t.Fatal(err)
	}
	var claims claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	if claims.Role != want {
		t.Errorf("role = %q, want %q", claims.Role, want)
	}
}
