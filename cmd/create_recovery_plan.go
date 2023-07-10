package cmd

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v2"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

type InstanceGroupPlan struct {
	Name                string            `yaml:"name"`
	MaxInFlightOverride string            `yaml:"max_in_flight_override,omitempty"`
	PlannedResolutions  map[string]string `yaml:"planned_resolutions"`
}

func (p InstanceGroupPlan) resolutionName(problem boshdir.Problem) string {
	return p.PlannedResolutions[problem.Type]
}

func (p InstanceGroupPlan) resolutionPlan(problem boshdir.Problem) string {
	for _, r := range problem.Resolutions {
		if *r.Name == p.resolutionName(problem) {
			return r.Plan
		}
	}

	return "No resolution planned"
}

type RecoveryPlan struct {
	InstanceGroupsPlan []InstanceGroupPlan `yaml:"instance_groups_plan"`
}

type CreateRecoveryPlanCmd struct {
	deployment boshdir.Deployment
	ui         boshui.UI
	fs         boshsys.FileSystem
}

func NewCreateRecoveryPlanCmd(deployment boshdir.Deployment, ui boshui.UI, fs boshsys.FileSystem) CreateRecoveryPlanCmd {
	return CreateRecoveryPlanCmd{deployment: deployment, ui: ui, fs: fs}
}

func (c CreateRecoveryPlanCmd) Run(opts CreateRecoveryPlanOpts) error {
	problemsByInstanceGroup, err := getProblemsByInstanceGroup(c.deployment)
	if err != nil {
		return err
	}

	if len(problemsByInstanceGroup) == 0 {
		c.ui.PrintLinef("No problems found\n")
		return nil
	}

	maxInFlightByInstanceGroup, err := c.getMaxInFlightByInstanceGroup()
	if err != nil {
		return err
	}

	var plan RecoveryPlan
	for _, instanceGroup := range sortedMapKeys(problemsByInstanceGroup) {
		c.ui.PrintLinef("Instance Group '%s'\n", instanceGroup)

		instanceGroupResolutions, err := c.processProblemsByType(problemsByInstanceGroup[instanceGroup])
		if err != nil {
			return err
		}

		instanceGroupCurrentMaxInFlight := maxInFlightByInstanceGroup[instanceGroup]
		var instanceGroupMaxInFlightOverride string
		if c.ui.AskForConfirmationWithLabel(
			fmt.Sprintf("Override current max_in_flight value of '%s'?", instanceGroupCurrentMaxInFlight),
		) == nil {
			instanceGroupMaxInFlightOverride, err = c.ui.AskForTextWithDefaultValue(
				fmt.Sprintf("max_in_flight override for '%s'", instanceGroup),
				instanceGroupCurrentMaxInFlight,
			)
			if err != nil {
				return err
			}
		}

		plan.InstanceGroupsPlan = append(plan.InstanceGroupsPlan, InstanceGroupPlan{
			Name:                instanceGroup,
			MaxInFlightOverride: instanceGroupMaxInFlightOverride,
			PlannedResolutions:  instanceGroupResolutions,
		})
	}

	bytes, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}

	return c.fs.WriteFile(opts.Args.RecoveryPlan.ExpandedPath, bytes)
}

type updateInstanceGroup struct {
	Name   string                 `yaml:"name"`
	Update map[string]interface{} `yaml:"update"`
}

type updateManifest struct {
	InstanceGroups []updateInstanceGroup  `yaml:"instance_groups"`
	Update         map[string]interface{} `yaml:"update"`
}

func (c CreateRecoveryPlanCmd) getMaxInFlightByInstanceGroup() (map[string]string, error) {
	rawManifest, err := c.deployment.Manifest()
	if err != nil {
		return nil, err
	}

	var updateManifest updateManifest
	err = yaml.Unmarshal([]byte(rawManifest), &updateManifest)
	if err != nil {
		return nil, err
	}

	globalMaxInFlight := updateManifest.Update["max_in_flight"]
	flightMap := make(map[string]string)
	for _, instanceGroup := range updateManifest.InstanceGroups {
		groupMaxInFlight := instanceGroup.Update["max_in_flight"]
		if groupMaxInFlight == nil {
			groupMaxInFlight = globalMaxInFlight
		}
		flightMap[instanceGroup.Name] = fmt.Sprintf("%v", groupMaxInFlight)
	}

	return flightMap, nil
}

func sortedMapKeys(problemMap map[string][]boshdir.Problem) []string {
	var keys []string
	for k := range problemMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func (c CreateRecoveryPlanCmd) processProblemsByType(problems []boshdir.Problem) (map[string]string, error) {
	problemsByType := mapProblemsByTrait(problems, func(p boshdir.Problem) string { return p.Type })

	resolutions := make(map[string]string)
	for _, problemType := range sortedMapKeys(problemsByType) {
		problemsForType := problemsByType[problemType]
		c.printProblemTable(problemType, problemsForType)

		var opts []string
		for _, res := range problemsForType[0].Resolutions {
			opts = append(opts, res.Plan)
		}

		chosenIndex, err := c.ui.AskForChoice(problemType, opts)
		if err != nil {
			return nil, err
		}

		resolutions[problemType] = *problemsForType[0].Resolutions[chosenIndex].Name
	}

	return resolutions, nil
}

func (c CreateRecoveryPlanCmd) printProblemTable(problemType string, problemsForType []boshdir.Problem) {
	table := boshtbl.Table{
		Title:   fmt.Sprintf("Problem type: %s", problemType),
		Content: fmt.Sprintf("%s problems", problemType),
		Header: []boshtbl.Header{
			boshtbl.NewHeader("#"),
			boshtbl.NewHeader("Description"),
		},
		SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
	}

	for _, p := range problemsForType {
		table.Rows = append(table.Rows, []boshtbl.Value{
			boshtbl.NewValueInt(p.ID),
			boshtbl.NewValueString(p.Description),
		})
	}

	c.ui.PrintTable(table)
}

func getProblemsByInstanceGroup(deployment boshdir.Deployment) (map[string][]boshdir.Problem, error) {
	problems, err := deployment.ScanForProblems()
	if err != nil {
		return nil, err
	}

	if anyProblemsHaveNoInstanceGroups(problems) {
		return nil, bosherr.Error("Director does not support this command.  Try 'bosh cloud-check' instead")
	}

	return mapProblemsByTrait(problems, func(p boshdir.Problem) string { return p.InstanceGroup }), nil
}

func anyProblemsHaveNoInstanceGroups(problems []boshdir.Problem) bool {
	for _, p := range problems {
		if p.InstanceGroup == "" {
			return true
		}
	}

	return false
}

func mapProblemsByTrait(problems []boshdir.Problem, traitFunc func(p boshdir.Problem) string) map[string][]boshdir.Problem {
	probMap := make(map[string][]boshdir.Problem)
	for _, p := range problems {
		probMap[traitFunc(p)] = append(probMap[traitFunc(p)], p)
	}

	return probMap
}
