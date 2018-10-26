package rotation

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	boshui "github.com/cloudfoundry/bosh-cli/ui"
)

/*
To kick-off you need:

1. credhub set -n /main/cf/credential_rotation_action -t value -v create-and-deploy-transitional-cas
2. bosh deploy ... --progressive-certificate-rotation
*/

const (
	// PhaseCreateAndDeployTransitionals will look for all CAs in the deployment, and generate a new transitional CA,
	// if no transitional CA already exists. Any transitional CAs are then appended into the manifest before deployment.
	PhaseCreateAndDeployTransitionals = "create-and-deploy-transitional-cas"

	// PhaseCreateAndDeployChildCerts will delete all leaf certificates in a deployment, and then look for all CAs,
	// and ensure that the oldest current version is marked as transitional, and that value appended into the manifest
	// before deployment. During deployment bosh will then generate new leaf certificates for all child nodes.
	PhaseCreateAndDeployChildCerts = "create-and-deploy-new-child-certs"

	// PhaseRemoveLegacyCAs will remove all transitional CAs
	PhaseRemoveLegacyCAs = "remove-legacy-cas"

	// PhaseNone will do nothing.
	PhaseNone = "no-action-required"
)

// CredhubRotator will talk directly to Credhub prior to a bosh deployment to rotate certificates as needed.
type CredhubRotator struct {
	Prefix                 string
	CredhubBaseURL         string
	CredhubCACerts         string
	CredhubUAAClient       string
	CredhubUAAClientSecret string

	UI boshui.UI
}

const (
	actionPreDeploy  = 1
	actionPostDeploy = 2
)

// enumerateCertificates find all certificates under prefix, only returning those that are or are not CAs (depending on isCA flag)
// Return value is a map of name to the credhub response object
func enumerateCertificates(chc credhubClient, prefix string, isCA bool) (map[string]*credhubCredResp, error) {
	creds, err := chc.ListCredentials(prefix)
	if err != nil {
		return nil, err
	}

	rv := make(map[string]*credhubCredResp)
	for _, cred := range creds.Credentials {
		cd, err := chc.GetCredential(cred.Name)
		if err != nil {
			return nil, err
		}
		if len(cd.Data) == 0 {
			return nil, errors.New("no creds returned")
		}

		for _, c := range cd.Data {
			if c.Type == "certificate" {
				asMap, ok := c.Value.(map[string]interface{})
				if !ok {
					return nil, errors.New("cannot decode cert type value")
				}
				certificateString, ok := asMap["certificate"].(string)
				if !ok {
					return nil, errors.New("cannot decode cert type value (2)")
				}

				block, _ := pem.Decode([]byte(certificateString))
				if block == nil {
					return nil, errors.New("no cert found")
				}
				if block.Type != "CERTIFICATE" {
					return nil, errors.New("pem not cert")
				}
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return nil, err
				}
				if cert.IsCA == isCA {
					rv[cred.Name] = cd
				}
			}
		}
	}
	return rv, nil
}

func (c *CredhubRotator) prepareForCreateAndDeployTransitionals(chc credhubClient) (map[string]string, error) {
	c.UI.PrintLinef("Beginning certificate rotation process, step 1 of 3 - Creating transitional CAs...")

	certsToRotate, err := enumerateCertificates(chc, c.Prefix, true)
	if err != nil {
		return nil, err
	}

	rv := make(map[string]string)
	for certName, cd := range certsToRotate {
		if len(cd.Data) > 1 {
			c.UI.PrintLinef("More than one active cert found, we won't create new transitional: %s", certName)
			for _, cvd := range cd.Data {
				if !cvd.Transitional { // bosh seems to pick up the transitional one?
					err = c.setTransitionalInRVMap(rv, certName, cvd)
					if err != nil {
						return nil, err
					}
				}
			}
			continue
		}

		// only attempt rotation if we are in a clean state, ie exactly one active
		c.UI.PrintLinef("Creating new transitional CA: %s", certName)

		// get certificate ID - (this is different?)
		certID, err := chc.GetCertificateID(certName)
		if err != nil {
			return nil, err
		}

		// regenerate it
		_, err = chc.MakeTransitionalCertificate(certID)
		if err != nil {
			return nil, err
		}

		// include old one in map, as bosh seems to pick up the new one
		err = c.setTransitionalInRVMap(rv, certName, cd.Data[0])
		if err != nil {
			return nil, err
		}
	}

	return rv, nil
}

func (c *CredhubRotator) setTransitionalInRVMap(rv map[string]string, certName string, cvd *credVersionData) error {
	m, ok := cvd.Value.(map[string]interface{})
	if !ok {
		return errors.New("bad coersion")
	}
	key := fmt.Sprintf("%s.certificate", certName[len(c.Prefix)+1:])
	rv[key] = fmt.Sprintf("%s\n%s", m["certificate"], rv[key])
	return nil
}

func (c *CredhubRotator) prepareForCreateAndDeployChildCerts(chc credhubClient) (map[string]string, error) {
	c.UI.PrintLinef("Continuing certificate rotation process, step 2 of 3 - regenerating leaf certificates...")

	// First, delete all leaf certificates - regardless of what happens next, these are always safe to regenerate
	certsToDelete, err := enumerateCertificates(chc, c.Prefix, false)
	if err != nil {
		return nil, err
	}
	for certName := range certsToDelete {
		c.UI.PrintLinef("Deleting leaf certificate: %s", certName)
		err := chc.DeleteCredential(certName)
		if err != nil {
			return nil, err
		}
	}

	// Now make sure oldest CAs are the transitional ones
	certsToRotate, err := enumerateCertificates(chc, c.Prefix, true)
	if err != nil {
		return nil, err
	}
	rv := make(map[string]string)
	for certName, cd := range certsToRotate {
		transitionalID := ""

		// Assume first, and that's OK as we surely wouldn't be here if there were 0?
		oldestID := cd.Data[0].ID
		oldestTime := cd.Data[0].VersionCreatedAt
		for _, cvd := range cd.Data {
			if cvd.Transitional {
				transitionalID = cvd.ID
			}
			if cvd.VersionCreatedAt.Before(oldestTime) {
				oldestID = cvd.ID
			}
		}

		switch transitionalID {
		case "":
			c.UI.PrintLinef("No transitional CA found to flip: %s", certName)

		case oldestID:
			c.UI.PrintLinef("Oldest CA is already transitional: %s", certName)

		default:
			c.UI.PrintLinef("Flipping transitional flags for: %s", certName)

			// get certificate ID - (this is different?)
			certID, err := chc.GetCertificateID(certName)
			if err != nil {
				return nil, err
			}

			err = chc.MakeThisOneTransitional(certID, oldestID)
			if err != nil {
				return nil, err
			}

			// make the rest of this code easier
			transitionalID = oldestID
		}

		// Now transitional is the old one, so put that as extra
		for _, cvd := range cd.Data {
			if cvd.ID == transitionalID {
				c.setTransitionalInRVMap(rv, certName, cvd)
			}
		}
	}

	return rv, nil
}

func (c *CredhubRotator) prepareForRemoveLegacyCAs(chc credhubClient) (map[string]string, error) {
	c.UI.PrintLinef("Continuing certificate rotation process, step 3 of 3 - removing legacy CAs...")

	certsToRotate, err := enumerateCertificates(chc, c.Prefix, true)
	if err != nil {
		return nil, err
	}
	for certName, cd := range certsToRotate {
		hasTransitional := false
		for _, cvd := range cd.Data {
			if cvd.Transitional {
				hasTransitional = true
			}
		}
		if !hasTransitional {
			c.UI.PrintLinef("No transitional CA found for: %s", certName)
			continue
		}

		// get certificate ID - (this is different?)
		certID, err := chc.GetCertificateID(certName)
		if err != nil {
			return nil, err
		}

		c.UI.PrintLinef("Removing transitional flags for: %s", certName)
		err = chc.MakeThisOneTransitional(certID, "")
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// PrepareForNewDeploy returns a map of values that should be appended before normal
// variable substition.
func (c *CredhubRotator) PrepareForNewDeploy() (map[string]string, error) {
	chc, action, err := c.getActionAndClient()
	if err != nil {
		return nil, err
	}

	switch action {
	case PhaseCreateAndDeployTransitionals:
		return c.prepareForCreateAndDeployTransitionals(chc)

	case PhaseCreateAndDeployChildCerts:
		return c.prepareForCreateAndDeployChildCerts(chc)

	case PhaseRemoveLegacyCAs:
		return c.prepareForRemoveLegacyCAs(chc)

	case PhaseNone:
		c.UI.PrintLinef("No certification rotation will be performed.")
		return nil, nil

	default:
		return nil, fmt.Errorf("unrecognized certificate rotation action requested: %s", action)
	}
}

// PostSuccessfulDeploy returns whether another deployment run is needed.
func (c *CredhubRotator) PostSuccessfulDeploy() (bool, error) {
	chc, action, err := c.getActionAndClient()
	if err != nil {
		return false, err
	}

	var nextAction string
	var newDeploy bool
	switch action {
	case PhaseCreateAndDeployTransitionals:
		nextAction, newDeploy = PhaseCreateAndDeployChildCerts, true

	case PhaseCreateAndDeployChildCerts:
		nextAction, newDeploy = PhaseRemoveLegacyCAs, true

	case PhaseRemoveLegacyCAs:
		nextAction, newDeploy = PhaseNone, false

	case PhaseNone:
		nextAction, newDeploy = PhaseNone, false

	default:
		return false, fmt.Errorf("unrecognized certificate rotation action requested: %s", action)
	}

	err = chc.SetValueCredential(c.actionKey(), nextAction)
	if err != nil {
		return false, err
	}

	return newDeploy, nil
}

func (c *CredhubRotator) actionKey() string {
	return fmt.Sprintf("%s/credential_rotation_action", c.Prefix)
}

func (c *CredhubRotator) getActionAndClient() (credhubClient, string, error) {
	chc, err := newCredhubClient(c.CredhubCACerts, c.CredhubBaseURL, c.CredhubUAAClient, c.CredhubUAAClientSecret)
	if err != nil {
		return nil, "", err
	}

	cr, err := chc.GetCredentialValue(c.actionKey(), PhaseNone)
	if err != nil {
		return nil, "", err
	}

	return chc, cr, nil
}
