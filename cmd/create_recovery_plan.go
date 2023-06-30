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
	Name               string                               `yaml:"name"`
	MaxInFlight        string                               `yaml:"max_in_flight,omitempty"`
	PlannedResolutions map[string]boshdir.ProblemResolution `yaml:"planned_resolutions"`
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
	problemsByInstanceGroup, err := c.getProblemsByInstanceGroup()
	if err != nil {
		return err
	}

	if len(problemsByInstanceGroup) == 0 {
		c.ui.PrintLinef("No problems found\n")
		return nil
	}

	var plan RecoveryPlan
	for _, instanceGroup := range sortedInstanceGroups(problemsByInstanceGroup) {
		c.ui.PrintLinef("Instance Group '%s'\n", instanceGroup)

		instanceGroupResolutions, err := c.processProblemsByType(problemsByInstanceGroup[instanceGroup])
		if err != nil {
			return err
		}

		plan.InstanceGroupsPlan = append(plan.InstanceGroupsPlan, InstanceGroupPlan{
			Name:               instanceGroup,
			PlannedResolutions: instanceGroupResolutions,
		})
	}

	bytes, err := yaml.Marshal(plan)
	if err != nil {
		return err
	}

	return c.fs.WriteFile(opts.Args.RecoveryPlan.ExpandedPath, bytes)
}

func sortedInstanceGroups(problemsByInstanceGroup map[string][]boshdir.Problem) []string {
	var instanceGroups []string
	for k := range problemsByInstanceGroup {
		instanceGroups = append(instanceGroups, k)
	}
	sort.Strings(instanceGroups)

	return instanceGroups
}

func (c CreateRecoveryPlanCmd) processProblemsByType(problems []boshdir.Problem) (map[string]boshdir.ProblemResolution, error) {
	problemsByType := mapProblemsByTrait(problems, func(p boshdir.Problem) string { return p.Type })

	resolutions := make(map[string]boshdir.ProblemResolution)
	for problemType, problemsForType := range problemsByType {
		c.printProblemTable(problemType, problemsForType)

		var opts []string
		for _, res := range problemsForType[0].Resolutions {
			opts = append(opts, res.Plan)
		}

		chosenIndex, err := c.ui.AskForChoice(problemType, opts)
		if err != nil {
			return nil, err
		}

		resolutions[problemType] = problemsForType[0].Resolutions[chosenIndex]
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

func (c CreateRecoveryPlanCmd) getProblemsByInstanceGroup() (map[string][]boshdir.Problem, error) {
	problems, err := c.deployment.ScanForProblems()
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
