package cmd

import (
	"fmt"

	"github.com/cloudfoundry/bosh-agent/v2/agentpassword"
	"github.com/cloudfoundry/bosh-utils/property"
)

func hashCPIAgentMbus(props property.Map) error {
	agent, _ := props["agent"].(property.Map)
	mbus, ok := agent["mbus"].(string)
	if !ok {
		return nil
	}
	hashed, err := agentpassword.HashURL(mbus)
	if err != nil {
		return err
	}
	agent["mbus"] = hashed
	return nil
}

func hashResourcePoolMbusURLs(env property.Map) error {
	boshRaw, ok := env["bosh"]
	if !ok {
		return nil
	}
	bosh, ok := boshRaw.(property.Map)
	if !ok {
		return fmt.Errorf("env.bosh must be a map")
	}
	mbusRaw, ok := bosh["mbus"]
	if !ok {
		return nil
	}
	mbus, ok := mbusRaw.(property.Map)
	if !ok {
		return fmt.Errorf("env.bosh.mbus must be a map")
	}
	urlsRaw, ok := mbus["urls"]
	if !ok {
		return nil
	}
	urls, ok := urlsRaw.(property.List)
	if !ok {
		return fmt.Errorf("env.bosh.mbus.urls must be a list")
	}
	for i, entryRaw := range urls {
		entry, ok := entryRaw.(string)
		if !ok {
			return fmt.Errorf("env.bosh.mbus.urls entry must be a string")
		}
		hashed, err := agentpassword.HashURL(entry)
		if err != nil {
			return err
		}
		urls[i] = hashed
	}
	return nil
}
