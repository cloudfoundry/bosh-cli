package podman

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/anchore/stereoscope/internal/log"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

var (
	ErrNoSocketAddress = errors.New("no socket address")
	ErrNoHostAddress   = errors.New("no host address")
)

func ClientOverSSH() (*client.Client, error) {
	var clientOpts = []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	host, identity := getSSHAddress(configPaths)

	if v, found := os.LookupEnv("CONTAINER_HOST"); found && v != "" {
		host = v
	}

	if v, found := os.LookupEnv("CONTAINER_SSHKEY"); found && v != "" {
		identity = v
	}

	passPhrase := ""
	if v, found := os.LookupEnv("CONTAINER_PASSPHRASE"); found {
		passPhrase = v
	}

	sshConf, err := newSSHConf(host, identity, passPhrase)
	if err != nil {
		return nil, err
	}

	httpClient, err := httpClientOverSSH(sshConf)
	if err != nil {
		return nil, fmt.Errorf("making http client: %w", err)
	}

	clientOpts = append(clientOpts, func(c *client.Client) error {
		return client.WithHTTPClient(httpClient)(c)
	})
	// This http path is defined by podman's docs: https://github.com/containers/podman/blob/main/pkg/api/server/docs.go#L31-L34
	clientOpts = append(clientOpts, client.WithHost("http://d"))

	c, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed create remote client for podman: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	_, err = c.Ping(ctx)

	return c, err
}

func ClientOverUnixSocket() (*client.Client, error) {
	var clientOpts = []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	addr := getUnixSocketAddress(configPaths)
	if v, found := os.LookupEnv("CONTAINER_HOST"); found && v != "" {
		addr = v
	}

	if addr == "" { // in some cases there might not be any config file
		// we can try guessing; podman CLI does that
		socketPath := fmt.Sprintf("/run/user/%d/podman/podman.sock", os.Getuid())
		log.Debugf("no socket address was found. Trying default address: %s", socketPath)
		_, err := os.Stat(socketPath)
		if err != nil {
			log.Debugf("looking for socket file: %v", err)
			return nil, ErrNoSocketAddress
		}

		addr = fmt.Sprintf("unix://%s", socketPath)
	}

	clientOpts = append(clientOpts, client.WithHost(addr))

	c, err := client.NewClientWithOpts(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating local client for podman: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*3)
	defer cancel()
	_, err = c.Ping(ctx)

	return c, err
}

func GetClient() (*client.Client, error) {
	c, err := ClientOverUnixSocket()
	if err == nil {
		return c, nil
	}

	return ClientOverSSH()
}
