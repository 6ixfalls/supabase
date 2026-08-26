package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/6ixfalls/supabase/internal/secrets"
)

func TestChildEnvironmentReplacesSecretsAndDropsRoot(t *testing.T) {
	values, err := secrets.Derive("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatal(err)
	}
	environment := childEnvironment([]string{
		"PATH=/bin", "ROOT_SECRET=do-not-pass", "ANON_KEY=stale",
		"GOTRUE_JWT_SECRET=stale", "PGRST_JWT_SECRET=stale",
		"SUPABASE_PUBLISHABLE_KEYS=stale", "OTHER=kept",
	}, values, serviceEnvironments["auth"])
	joined := "\n" + strings.Join(environment, "\n") + "\n"
	if strings.Contains(joined, "\nROOT_SECRET=") {
		t.Error("ROOT_SECRET leaked into child environment")
	}
	if strings.Contains(joined, "\nANON_KEY=stale\n") {
		t.Error("stale derived credential was retained")
	}
	if strings.Contains(joined, "\nANON_KEY=") {
		t.Error("unselected ANON_KEY was passed to auth")
	}
	if !strings.Contains(joined, "\nOTHER=kept\n") {
		t.Error("unrelated environment variable was removed")
	}

	got := environmentMap(t, environment)
	for name, want := range map[string]string{
		"GOTRUE_JWT_SECRET": values["JWT_SECRET"],
		"GOTRUE_JWT_KEYS":   values["JWT_KEYS"],
	} {
		if got[name] != want {
			t.Errorf("%s = %q, want derived value", name, got[name])
		}
	}
	if len(got) != 4 {
		t.Fatalf("auth environment contains %d variables, want PATH, OTHER, and two auth secrets", len(got))
	}
}

func TestEveryServiceExportsOnlyItsSelectedSecrets(t *testing.T) {
	const rootSecret = "0123456789abcdef0123456789abcdef"
	managed := managedEnvironmentNames()
	allSecrets := make([]string, 0, len(managed)+1)
	for name := range managed {
		allSecrets = append(allSecrets, name+"=stale")
	}
	allSecrets = append(allSecrets, "OTHER=kept")

	for service, selected := range serviceEnvironments {
		t.Run(service, func(t *testing.T) {
			values, err := credentialValues(service, []string{"ROOT_SECRET=" + rootSecret})
			if err != nil {
				t.Fatal(err)
			}
			got := environmentMap(t, childEnvironment(allSecrets, values, selected))
			if got["OTHER"] != "kept" {
				t.Error("unrelated variable was not preserved")
			}
			if len(got) != len(selected)+1 {
				t.Fatalf("got %d variables, want %d selected variables plus OTHER", len(got), len(selected))
			}
			for name, source := range selected {
				want, exists := values[source]
				if !exists {
					t.Fatalf("profile maps %s to unknown derived variable %s", name, source)
				}
				want = environmentValue(name, want)
				if got[name] != want {
					t.Errorf("%s was not set to its derived value", name)
				}
			}
		})
	}

	values, err := credentialValues("functions", []string{"ROOT_SECRET=" + rootSecret})
	if err != nil {
		t.Fatal(err)
	}
	for plural, singular := range map[string]string{
		"SUPABASE_PUBLISHABLE_KEYS": "SUPABASE_PUBLISHABLE_KEY",
		"SUPABASE_SECRET_KEYS":      "SUPABASE_SECRET_KEY",
	} {
		var keyMap map[string]string
		got := environmentMap(t, childEnvironment(nil, values, serviceEnvironments["functions"]))
		if err := json.Unmarshal([]byte(got[plural]), &keyMap); err != nil {
			t.Fatalf("%s is not JSON: %v", plural, err)
		}
		if keyMap["default"] != values[singular] {
			t.Errorf("%s default key does not match %s", plural, singular)
		}
	}
}

func TestOnlyGatewayMintsAsymmetricRoleTokens(t *testing.T) {
	const rootSecret = "0123456789abcdef0123456789abcdef"
	auth, err := credentialValues("auth", []string{"ROOT_SECRET=" + rootSecret})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range secrets.GatewayNames {
		if _, exists := auth[name]; exists {
			t.Errorf("auth unexpectedly minted %s", name)
		}
	}

	first, err := credentialValues("gateway", []string{"ROOT_SECRET=" + rootSecret})
	if err != nil {
		t.Fatal(err)
	}
	second, err := credentialValues("gateway", []string{"ROOT_SECRET=" + rootSecret})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range secrets.GatewayNames {
		if first[name] == "" || second[name] == "" {
			t.Fatalf("gateway did not mint %s", name)
		}
		if first[name] == second[name] {
			t.Errorf("gateway reused randomized %s", name)
		}
	}
	for _, name := range secrets.Names {
		if first[name] != second[name] {
			t.Errorf("persistent credential %s changed across gateway starts", name)
		}
	}
}

func TestCredentialValuesRequireRootSecret(t *testing.T) {
	if _, err := credentialValues("auth", nil); err == nil {
		t.Fatal("expected missing ROOT_SECRET to fail")
	}
}

func TestParseArgs(t *testing.T) {
	service, command, err := parseArgs([]string{"--service", "auth", "--", "server", "arg"})
	if err != nil {
		t.Fatal(err)
	}
	if service != "auth" || strings.Join(command, " ") != "server arg" {
		t.Fatalf("got service %q and command %q", service, command)
	}

	for _, args := range [][]string{
		{"server"},
		{"--service", "unknown", "server"},
		{"--service", "auth"},
		{"--service", "auth", "--service", "rest", "server"},
		{"--bad", "auth", "server"},
	} {
		if _, _, err := parseArgs(args); err == nil {
			t.Errorf("parseArgs(%q) unexpectedly succeeded", args)
		}
	}
}

func environmentMap(t *testing.T, environment []string) map[string]string {
	t.Helper()
	result := make(map[string]string, len(environment))
	for _, entry := range environment {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			t.Fatalf("invalid environment entry %q", entry)
		}
		if _, exists := result[name]; exists {
			t.Fatalf("duplicate environment variable %s", name)
		}
		result[name] = value
	}
	return result
}
