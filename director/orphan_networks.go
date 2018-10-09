package director

import (
	"fmt"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type OrphanNetworkImpl struct {
	client     Client
	tpe        string
	name       string
	createdAt  time.Time
	orphanedAt time.Time
}

func (n OrphanNetworkImpl) Name() string          { return n.name }
func (n OrphanNetworkImpl) CreatedAt() time.Time  { return n.createdAt }
func (n OrphanNetworkImpl) OrphanedAt() time.Time { return n.orphanedAt }
func (n OrphanNetworkImpl) Type() string          { return n.tpe }
func (n OrphanNetworkImpl) Delete() error {
	err := n.client.DeleteOrphanNetwork(n.name)
	if err != nil {
		resps, listErr := n.client.OrphanNetworks()
		if listErr != nil {
			return err
		}

		for _, resp := range resps {
			if resp.Name == n.name {
				return err
			}
		}
	}

	return nil
}

type OrphanNetworkResp struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	CreatedAt  string `json:"created_at"`
	OrphanedAt string `json:"orphaned_at"` // e.g. "2016-01-09 06:23:25 +0000"
}

func (d DirectorImpl) FindOrphanNetwork(name string) (OrphanNetwork, error) {
	return OrphanNetworkImpl{client: d.client, name: name}, nil
}

func (d DirectorImpl) OrphanNetworks() ([]OrphanNetwork, error) {
	var networks []OrphanNetwork

	resps, err := d.client.OrphanNetworks()
	if err != nil {
		return networks, err
	}

	for _, r := range resps {
		orphanedAt, err := TimeParser{}.Parse(r.OrphanedAt)
		if err != nil {
			return networks, bosherr.WrapErrorf(err, "Converting orphaned at '%s' to time", r.OrphanedAt)
		}

		createdAt, err := TimeParser{}.Parse(r.CreatedAt)
		if err != nil {
			return networks, bosherr.WrapErrorf(err, "Converting created at '%s' to time", r.OrphanedAt)
		}

		network := OrphanNetworkImpl{
			client:     d.client,
			name:       r.Name,
			tpe:        r.Type,
			createdAt:  createdAt.UTC(),
			orphanedAt: orphanedAt.UTC(),
		}

		networks = append(networks, network)
	}

	return networks, nil
}

func (c Client) OrphanNetworks() ([]OrphanNetworkResp, error) {
	var networks []OrphanNetworkResp

	err := c.clientRequest.Get("/networks?orphaned=true", &networks)
	if err != nil {
		return networks, bosherr.WrapErrorf(err, "Finding orphaned networks")
	}

	return networks, nil
}

func (c Client) DeleteOrphanNetwork(name string) error {
	if len(name) == 0 {
		return bosherr.Error("Expected non-empty orphaned network name")
	}

	path := fmt.Sprintf("/networks/%s", name)

	_, err := c.taskClientRequest.DeleteResult(path)
	if err != nil {
		return bosherr.WrapErrorf(err, "Deleting orphaned network '%s'", name)
	}

	return nil
}
