// Package secrets derives persistent Supabase credentials from one root secret
// and mints the gateway's internal asymmetric API-key JWTs. Derivation labels
// and persistent output formats are part of the package's public compatibility
// contract: changing either rotates the derived credentials.
package secrets

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
)

const (
	MinimumRootSecretBytes = 32
	defaultIssuedAt        = int64(1_700_000_000) // 2023-11-14
	defaultExpiresAt       = int64(4_102_444_800) // 2100-01-01
	projectRef             = "supabase-self-hosted"
	derivationVersion      = "supabase-secrets/v1"
	p256OrderHex           = "FFFFFFFF00000000FFFFFFFFFFFFFFFFBCE6FAADA7179E84F3B9CAC2FC632551"
)

// Names contains credentials whose exact values are stable for a given
// root secret and derivation version.
var Names = []string{
	"JWT_SECRET",
	"ANON_KEY",
	"SERVICE_ROLE_KEY",
	"SUPABASE_PUBLISHABLE_KEY",
	"SUPABASE_SECRET_KEY",
	"JWT_KEYS",
	"JWT_JWKS",
}

// GatewayNames contains internal, pre-signed JWTs minted independently each
// time the gateway starts. Their signatures are intentionally not stable.
var GatewayNames = []string{
	"ANON_KEY_ASYMMETRIC",
	"SERVICE_ROLE_KEY_ASYMMETRIC",
}

type claims struct {
	Role string `json:"role"`
	Iss  string `json:"iss"`
	Iat  int64  `json:"iat"`
	Exp  int64  `json:"exp"`
}

type ecJWK struct {
	Kty    string   `json:"kty"`
	Kid    string   `json:"kid"`
	Use    string   `json:"use"`
	KeyOps []string `json:"key_ops"`
	Alg    string   `json:"alg"`
	Ext    bool     `json:"ext"`
	Crv    string   `json:"crv"`
	X      string   `json:"x"`
	Y      string   `json:"y"`
	D      string   `json:"d,omitempty"`
}

type octJWK struct {
	Kty string `json:"kty"`
	K   string `json:"k"`
	Alg string `json:"alg"`
}

// Derive returns persistent credentials derived from ROOT_SECRET. Results are
// stable for a given root secret and derivation version.
func Derive(rootSecret string) (map[string]string, error) {
	if err := validateRootSecret(rootSecret); err != nil {
		return nil, err
	}

	jwtSecret := base64.StdEncoding.EncodeToString(derive(rootSecret, "jwt-secret", 30))
	anonClaims := claims{Role: "anon", Iss: "supabase", Iat: defaultIssuedAt, Exp: defaultExpiresAt}
	serviceClaims := claims{Role: "service_role", Iss: "supabase", Iat: defaultIssuedAt, Exp: defaultExpiresAt}

	anonKey, err := signHS256(anonClaims, jwtSecret)
	if err != nil {
		return nil, err
	}
	serviceRoleKey, err := signHS256(serviceClaims, jwtSecret)
	if err != nil {
		return nil, err
	}

	privateKey, err := deriveP256Key(rootSecret)
	if err != nil {
		return nil, err
	}
	kid := deriveUUID(rootSecret, "jwt-kid")
	privateBytes, err := privateKey.Bytes()
	if err != nil {
		return nil, err
	}
	publicKey, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("derived key is not ECDSA")
	}
	publicBytes, err := publicKey.Bytes()
	if err != nil {
		return nil, err
	}

	privateJWK := ecJWK{
		Kty: "EC", Kid: kid, Use: "sig", KeyOps: []string{"sign", "verify"},
		Alg: "ES256", Ext: true, Crv: "P-256",
		X: base64URL(publicBytes[1:33]), Y: base64URL(publicBytes[33:]),
		D: base64URL(privateBytes),
	}
	publicJWK := privateJWK
	publicJWK.KeyOps = []string{"verify"}
	publicJWK.D = ""
	octKey := octJWK{Kty: "oct", K: base64URL([]byte(jwtSecret)), Alg: "HS256"}

	// GOTRUE_JWT_KEYS expects the signing keys as a bare JSON array, matching
	// Supabase's docker/utils/add-new-auth-keys.sh output.
	jwtKeys, err := json.Marshal([]any{privateJWK, octKey})
	if err != nil {
		return nil, err
	}
	jwtJWKS, err := json.Marshal(struct {
		Keys []any `json:"keys"`
	}{Keys: []any{publicJWK, octKey}})
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"JWT_SECRET":               jwtSecret,
		"ANON_KEY":                 anonKey,
		"SERVICE_ROLE_KEY":         serviceRoleKey,
		"SUPABASE_PUBLISHABLE_KEY": opaqueKey(rootSecret, "publishable", "sb_publishable_"),
		"SUPABASE_SECRET_KEY":      opaqueKey(rootSecret, "secret", "sb_secret_"),
		"JWT_KEYS":                 string(jwtKeys),
		"JWT_JWKS":                 string(jwtJWKS),
	}, nil
}

// MintGatewayTokens creates the gateway's internal ES256 JWT API keys. The
// claims and signing key are stable, but ECDSA signatures are randomized and
// callers must not rely on the returned token bytes surviving a restart.
func MintGatewayTokens(rootSecret string) (map[string]string, error) {
	if err := validateRootSecret(rootSecret); err != nil {
		return nil, err
	}
	privateKey, err := deriveP256Key(rootSecret)
	if err != nil {
		return nil, err
	}
	kid := deriveUUID(rootSecret, "jwt-kid")
	anonClaims := claims{Role: "anon", Iss: "supabase", Iat: defaultIssuedAt, Exp: defaultExpiresAt}
	serviceClaims := claims{Role: "service_role", Iss: "supabase", Iat: defaultIssuedAt, Exp: defaultExpiresAt}
	anonKey, err := signES256(anonClaims, privateKey, kid)
	if err != nil {
		return nil, err
	}
	serviceRoleKey, err := signES256(serviceClaims, privateKey, kid)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"ANON_KEY_ASYMMETRIC":         anonKey,
		"SERVICE_ROLE_KEY_ASYMMETRIC": serviceRoleKey,
	}, nil
}

func validateRootSecret(rootSecret string) error {
	if len(rootSecret) < MinimumRootSecretBytes {
		return errors.New("ROOT_SECRET must contain at least 32 bytes")
	}
	return nil
}

// derive is RFC 5869 HKDF-SHA256 with a versioned salt and domain-separated
// info. Current outputs need at most one SHA-256 block, but this supports more.
func derive(rootSecret, label string, length int) []byte {
	extract := hmac.New(sha256.New, []byte(derivationVersion))
	extract.Write([]byte(rootSecret))
	prk := extract.Sum(nil)
	var output, previous []byte
	for counter := byte(1); len(output) < length; counter++ {
		expand := hmac.New(sha256.New, prk)
		expand.Write(previous)
		expand.Write([]byte(label))
		expand.Write([]byte{counter})
		previous = expand.Sum(nil)
		output = append(output, previous...)
	}
	return output[:length]
}

func deriveP256Key(rootSecret string) (*ecdsa.PrivateKey, error) {
	n := new(big.Int).Sub(p256Order(), big.NewInt(1))
	d := new(big.Int).SetBytes(derive(rootSecret, "jwt-es256-private-key", 32))
	d.Mod(d, n)
	d.Add(d, big.NewInt(1))
	return ecdsa.ParseRawPrivateKey(elliptic.P256(), paddedBytes(d, 32))
}

func deriveUUID(rootSecret, label string) string {
	b := derive(rootSecret, label, 16)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	hex := "0123456789abcdef"
	encoded := make([]byte, 36)
	positions := []int{8, 13, 18, 23}
	for _, position := range positions {
		encoded[position] = '-'
	}
	for source, target := 0, 0; source < len(b); source++ {
		for encoded[target] == '-' {
			target++
		}
		encoded[target] = hex[b[source]>>4]
		encoded[target+1] = hex[b[source]&0x0f]
		target += 2
	}
	return string(encoded)
}

func opaqueKey(rootSecret, label, prefix string) string {
	random := base64URL(derive(rootSecret, "opaque-"+label, 17))[:22]
	intermediate := prefix + random
	checksum := sha256.Sum256([]byte(projectRef + "|" + intermediate))
	return intermediate + "_" + base64URL(checksum[:])[:8]
}

func signHS256(payload claims, secret string) (string, error) {
	header := base64URL([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	content := header + "." + base64URL(body)
	signature := hmac.New(sha256.New, []byte(secret))
	signature.Write([]byte(content))
	return content + "." + base64URL(signature.Sum(nil)), nil
}

func signES256(payload claims, privateKey *ecdsa.PrivateKey, kid string) (string, error) {
	header, err := json.Marshal(struct {
		Alg string `json:"alg"`
		Typ string `json:"typ"`
		Kid string `json:"kid"`
	}{Alg: "ES256", Typ: "JWT", Kid: kid})
	if err != nil {
		return "", err
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	content := base64URL(header) + "." + base64URL(body)
	digest := sha256.Sum256([]byte(content))
	encodedSignature, err := privateKey.Sign(rand.Reader, digest[:], crypto.SHA256)
	if err != nil {
		return "", err
	}
	var values struct {
		R, S *big.Int
	}
	rest, err := asn1.Unmarshal(encodedSignature, &values)
	if err != nil {
		return "", err
	}
	if len(rest) != 0 || values.R == nil || values.S == nil {
		return "", errors.New("invalid ECDSA signature encoding")
	}
	signature := append(paddedBytes(values.R, 32), paddedBytes(values.S, 32)...)
	return content + "." + base64URL(signature), nil
}

func paddedBytes(value *big.Int, length int) []byte {
	result := make([]byte, length)
	value.FillBytes(result)
	return result
}

func p256Order() *big.Int {
	order, ok := new(big.Int).SetString(p256OrderHex, 16)
	if !ok {
		panic("invalid P-256 order")
	}
	return order
}

func base64URL(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}
