// supabase-secrets is an entrypoint wrapper that derives service-scoped
// Supabase credentials from ROOT_SECRET and replaces itself with the requested
// command.
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/6ixfalls/supabase/internal/secrets"
	"golang.org/x/sys/unix"
)

// serviceEnvironments maps each exported variable to the value returned by
// secrets.Derive that supplies it.
var serviceEnvironments = map[string]map[string]string{
	"studio": {
		"SUPABASE_ANON_KEY":        "ANON_KEY",
		"SUPABASE_SERVICE_KEY":     "SERVICE_ROLE_KEY",
		"AUTH_JWT_SECRET":          "JWT_SECRET",
		"SUPABASE_PUBLISHABLE_KEY": "SUPABASE_PUBLISHABLE_KEY",
		"SUPABASE_SECRET_KEY":      "SUPABASE_SECRET_KEY",
	},
	"gateway": {
		"ANON_KEY":                    "ANON_KEY",
		"SERVICE_ROLE_KEY":            "SERVICE_ROLE_KEY",
		"SUPABASE_PUBLISHABLE_KEY":    "SUPABASE_PUBLISHABLE_KEY",
		"SUPABASE_SECRET_KEY":         "SUPABASE_SECRET_KEY",
		"ANON_KEY_ASYMMETRIC":         "ANON_KEY_ASYMMETRIC",
		"SERVICE_ROLE_KEY_ASYMMETRIC": "SERVICE_ROLE_KEY_ASYMMETRIC",
		"SUPABASE_ANON_KEY":           "ANON_KEY",
		"SUPABASE_SERVICE_KEY":        "SERVICE_ROLE_KEY",
	},
	"auth": {
		"GOTRUE_JWT_SECRET": "JWT_SECRET",
		"GOTRUE_JWT_KEYS":   "JWT_KEYS",
	},
	"rest": {
		"PGRST_JWT_SECRET":              "JWT_JWKS",
		"PGRST_APP_SETTINGS_JWT_SECRET": "JWT_SECRET",
	},
	"realtime": {
		"API_JWT_SECRET":     "JWT_SECRET",
		"API_JWT_JWKS":       "JWT_JWKS",
		"METRICS_JWT_SECRET": "JWT_SECRET",
	},
	"storage": {
		"ANON_KEY":        "ANON_KEY",
		"SERVICE_KEY":     "SERVICE_ROLE_KEY",
		"AUTH_JWT_SECRET": "JWT_SECRET",
		"JWT_JWKS":        "JWT_JWKS",
	},
	"functions": {
		"JWT_SECRET":                "JWT_SECRET",
		"SUPABASE_JWKS":             "JWT_JWKS",
		"SUPABASE_ANON_KEY":         "ANON_KEY",
		"SUPABASE_SERVICE_ROLE_KEY": "SERVICE_ROLE_KEY",
		"SUPABASE_PUBLISHABLE_KEYS": "SUPABASE_PUBLISHABLE_KEY",
		"SUPABASE_SECRET_KEYS":      "SUPABASE_SECRET_KEY",
	},
	"db": {
		"JWT_SECRET": "JWT_SECRET",
	},
	"supavisor": {
		"API_JWT_SECRET":     "JWT_SECRET",
		"METRICS_JWT_SECRET": "JWT_SECRET",
	},
}

func main() {
	os.Exit(run(os.Args[1:], os.Environ()))
}

func run(args, environment []string) int {
	service, command, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "supabase-secrets: %v\n%s\n", err, usage())
		return 2
	}

	values, err := credentialValues(service, environment)
	if err != nil {
		fmt.Fprintf(os.Stderr, "supabase-secrets: %v\n", err)
		return 1
	}

	path, err := exec.LookPath(command[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "supabase-secrets: %v\n", err)
		return 127
	}
	err = unix.Exec(path, command, childEnvironment(environment, values, serviceEnvironments[service]))
	fmt.Fprintf(os.Stderr, "supabase-secrets: exec %s: %v\n", command[0], err)
	return 126
}

func credentialValues(service string, environment []string) (map[string]string, error) {
	rootSecret := lookup(environment, "ROOT_SECRET")
	values, err := secrets.Derive(rootSecret)
	if err != nil {
		return nil, err
	}
	if service == "gateway" {
		gatewayValues, err := secrets.MintGatewayTokens(rootSecret)
		if err != nil {
			return nil, err
		}
		for name, value := range gatewayValues {
			values[name] = value
		}
	}
	return values, nil
}

func parseArgs(args []string) (string, []string, error) {
	var service string
	for len(args) > 0 {
		switch {
		case args[0] == "--":
			args = args[1:]
			goto parsed
		case args[0] == "--service":
			if service != "" {
				return "", nil, errors.New("--service may only be specified once")
			}
			if len(args) < 2 || args[1] == "" {
				return "", nil, errors.New("--service requires a value")
			}
			service = args[1]
			args = args[2:]
		case strings.HasPrefix(args[0], "--service="):
			if service != "" {
				return "", nil, errors.New("--service may only be specified once")
			}
			service = strings.TrimPrefix(args[0], "--service=")
			if service == "" {
				return "", nil, errors.New("--service requires a value")
			}
			args = args[1:]
		case strings.HasPrefix(args[0], "-"):
			return "", nil, fmt.Errorf("unknown option %s", args[0])
		default:
			goto parsed
		}
	}

parsed:
	if service == "" {
		return "", nil, errors.New("--service is required")
	}
	if _, ok := serviceEnvironments[service]; !ok {
		return "", nil, fmt.Errorf("unknown service %q", service)
	}
	if len(args) == 0 {
		return "", nil, errors.New("command is required")
	}
	return service, args, nil
}

func usage() string {
	services := make([]string, 0, len(serviceEnvironments))
	for service := range serviceEnvironments {
		services = append(services, service)
	}
	sort.Strings(services)
	return "usage: supabase-secrets --service <" + strings.Join(services, "|") + "> [--] command [argument ...]"
}

func childEnvironment(environment []string, values, selected map[string]string) []string {
	replaced := managedEnvironmentNames()

	result := make([]string, 0, len(environment)+len(selected))
	for _, entry := range environment {
		name, _, _ := strings.Cut(entry, "=")
		if !replaced[name] {
			result = append(result, entry)
		}
	}
	names := make([]string, 0, len(selected))
	for name := range selected {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		result = append(result, name+"="+environmentValue(name, values[selected[name]]))
	}
	return result
}

func managedEnvironmentNames() map[string]bool {
	names := make(map[string]bool, len(secrets.Names)+len(secrets.GatewayNames)+1)
	names["ROOT_SECRET"] = true
	for _, name := range secrets.Names {
		names[name] = true
	}
	for _, name := range secrets.GatewayNames {
		names[name] = true
	}
	for _, environment := range serviceEnvironments {
		for name := range environment {
			names[name] = true
		}
	}
	return names
}

func environmentValue(name, value string) string {
	if name == "SUPABASE_PUBLISHABLE_KEYS" || name == "SUPABASE_SECRET_KEYS" {
		return defaultKeyMap(value)
	}
	return value
}

func defaultKeyMap(value string) string {
	encoded, err := json.Marshal(map[string]string{"default": value})
	if err != nil {
		panic(err) // strings are always JSON-encodable
	}
	return string(encoded)
}

func lookup(environment []string, name string) string {
	prefix := name + "="
	for i := len(environment) - 1; i >= 0; i-- {
		if strings.HasPrefix(environment[i], prefix) {
			return strings.TrimPrefix(environment[i], prefix)
		}
	}
	return ""
}
