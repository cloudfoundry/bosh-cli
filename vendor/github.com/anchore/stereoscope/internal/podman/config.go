package podman

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/homedir"
	"github.com/pelletier/go-toml"
)

type containersConfig struct {
	Engine engine `toml:"engine"`
}

type engine struct {
	ActiveService       string                        `toml:"active_service"`
	ServiceDestinations map[string]serviceDestination `toml:"service_destinations"`
}

type serviceDestination struct {
	URI      string `toml:"uri"`
	Identity string `toml:"identity"`
}

func findUnixAddressFromFile(path string) string {
	cc, err := parseContainerConfig(path)
	if err != nil {
		return ""
	}

	if cc == nil {
		return ""
	}

	return findUnixAddress(*cc)
}

func findDestinationOfType(cc containersConfig, ty string) *serviceDestination {
	// always attempt what the config says is the current service
	if destination, ok := cc.Engine.ServiceDestinations[cc.Engine.ActiveService]; ok {
		if isScheme(destination.URI, ty) {
			return &destination
		}
	}

	// fallback to looking at all configured services if the active service is not found or is not unix
	for _, destination := range cc.Engine.ServiceDestinations {
		if destination.URI == "" {
			continue
		}
		if isScheme(destination.URI, ty) {
			return &destination
		}
	}
	return nil
}

func findSSHConnectionInfoFromFile(path string) (string, string) {
	cc, err := parseContainerConfig(path)
	if err != nil {
		return "", ""
	}

	if cc == nil {
		return "", ""
	}

	return findSSHConnectionInfo(*cc)
}
func findSSHConnectionInfo(cc containersConfig) (string, string) {
	dest := findDestinationOfType(cc, "ssh")
	if dest == nil {
		return "", ""
	}

	return dest.URI, dest.Identity
}

func findUnixAddress(cc containersConfig) string {
	dest := findDestinationOfType(cc, "unix")
	if dest == nil {
		return ""
	}
	return parseUnixAddress(dest.URI)
}

func parseUnixAddress(uri string) string {
	u, err := url.Parse(uri)
	if err != nil {
		return ""
	}

	if u.Scheme == "unix" {
		return fmt.Sprintf("unix://%s", u.Path)
	}
	return ""
}

func isScheme(uri, scheme string) bool {
	u, err := url.Parse(uri)
	if err != nil {
		return false
	}

	return u.Scheme == scheme
}

func parseContainerConfig(path string) (*containersConfig, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cc containersConfig
	if err := toml.Unmarshal(configBytes, &cc); err != nil {
		return nil, err
	}
	return &cc, nil
}

var (
	// configFile is the default dir + container config used by podman.
	configFile = filepath.Join("containers", "containers.conf")

	// configPaths holds a list of config files, they are sorted from
	// the least to the most relevant during reading.
	configPaths = []string{
		// holds the default containers config path
		filepath.Join("usr", "share", configFile),
		// holds the default config path overridden by the root user
		filepath.Join("etc", configFile),
		// holds the container config path overridden by the rootless user
		filepath.Join(homedir.Get(), ".config", configFile),
	}
)

func getUnixSocketAddress(paths []string) (address string) {
	for _, p := range paths {
		if a := findUnixAddressFromFile(p); a != "" {
			// overwriting here is intentional, as a way to
			// prioritize different config files
			address = a
		}
	}

	return
}

func getSSHAddress(paths []string) (address, identity string) {
	for _, p := range paths {
		a, id := findSSHConnectionInfoFromFile(p)
		// overwriting here is intentional, as a way to
		// prioritize different config files
		if a != "" && id != "" {
			address = a
			identity = id
			break
		}
	}

	return
}
