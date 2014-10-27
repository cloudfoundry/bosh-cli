package deployment

import (
	"fmt"
	"net/url"

	bosherr "github.com/cloudfoundry/bosh-agent/errors"
	bmkeystr "github.com/cloudfoundry/bosh-micro-cli/keystringifier"
)

type ReleaseJobRef struct {
	Name    string
	Release string
}

type Job struct {
	Name          string
	Instances     int
	Templates     []ReleaseJobRef
	Networks      []JobNetwork
	RawProperties map[interface{}]interface{} `yaml:"properties"`
}

func (j *Job) Properties() (map[string]interface{}, error) {
	return bmkeystr.NewKeyStringifier().ConvertMap(j.RawProperties)
}

type JobNetwork struct {
	Name      string
	StaticIPs []string `yaml:"static_ips"`
}

type Registry struct {
	Username string
	Password string
	Host     string
	Port     int
}

type SSHTunnel struct {
	User       string
	Host       string
	Port       int
	Password   string
	PrivateKey string `yaml:"private_key"`
}

type Deployment struct {
	Name            string
	Properties      map[string]interface{}
	Mbus            string
	Registry        Registry
	AgentEnvService string
	SSHTunnel       SSHTunnel
	Jobs            []Job
	Networks        []Network
	ResourcePools   []ResourcePool
}

func (d Deployment) NetworksSpec(jobName string) (map[string]interface{}, error) {
	job, found := d.findJobByName(jobName)
	if !found {
		return map[string]interface{}{}, bosherr.New("Could not find job with name: %s", jobName)
	}

	networksMap := d.networksToMap()

	result := map[string]interface{}{}
	var err error
	for _, jobNetwork := range job.Networks {
		network := networksMap[jobNetwork.Name]
		staticIPs := jobNetwork.StaticIPs
		if len(staticIPs) > 0 {
			network.IP = staticIPs[0]
		}
		result[jobNetwork.Name], err = network.Spec()
		if err != nil {
			return map[string]interface{}{}, bosherr.WrapError(err, "Building network spec")
		}
	}

	return result, nil
}

func (d Deployment) MbusConfig() (string, string, string, error) {
	parsedURL, err := url.Parse(d.Mbus)
	if err != nil {
		return "", "", "", bosherr.WrapError(err, "Parsing Mbus URL")
	}

	var username, password string
	userInfo := parsedURL.User
	if userInfo != nil {
		username = userInfo.Username()
		password, _ = userInfo.Password()
	}

	return fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host), username, password, nil
}

func (d Deployment) networksToMap() map[string]Network {
	result := map[string]Network{}
	for _, network := range d.Networks {
		result[network.Name] = network
	}
	return result
}

func (d Deployment) findJobByName(jobName string) (Job, bool) {
	for _, job := range d.Jobs {
		if job.Name == jobName {
			return job, true
		}
	}

	return Job{}, false
}
