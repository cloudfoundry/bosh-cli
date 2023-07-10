package cmd

import (
	"fmt"

	"gopkg.in/yaml.v2"

	boshsys "github.com/cloudfoundry/bosh-utils/system"

	boshdir "github.com/cloudfoundry/bosh-cli/v7/director"
	boshui "github.com/cloudfoundry/bosh-cli/v7/ui"
	boshtbl "github.com/cloudfoundry/bosh-cli/v7/ui/table"

	. "github.com/cloudfoundry/bosh-cli/v7/cmd/opts"
)

type RecoverCmd struct {
	deployment boshdir.Deployment
	ui         boshui.UI
	fs         boshsys.FileSystem
}

func NewRecoverCmd(deployment boshdir.Deployment, ui boshui.UI, fs boshsys.FileSystem) RecoverCmd {
	return RecoverCmd{deployment: deployment, ui: ui, fs: fs}
}

func (c RecoverCmd) Run(opts RecoverOpts) error {
	problemsByInstanceGroup, err := getProblemsByInstanceGroup(c.deployment)
	if err != nil {
		return err
	}

	if len(problemsByInstanceGroup) == 0 {
		c.ui.PrintLinef("No problems found\n")
		return nil
	}

	plan, err := c.readPlan(opts)
	if err != nil {
		return err
	}

	c.printPlanSummary(problemsByInstanceGroup, plan)
	if err := c.ui.AskForConfirmation(); err != nil {
		return err
	}

	var answers []boshdir.ProblemAnswer
	maxInFlightOverrides := make(map[string]string)
	for _, instanceGroupPlan := range plan.InstanceGroupsPlan {
		if instanceGroupPlan.MaxInFlightOverride != "" {
			maxInFlightOverrides[instanceGroupPlan.Name] = instanceGroupPlan.MaxInFlightOverride
		}

		instanceGroupAnswers := getAnswersFromPlan(problemsByInstanceGroup[instanceGroupPlan.Name], instanceGroupPlan)
		answers = append(answers, instanceGroupAnswers...)
	}

	err = c.deployment.ResolveProblems(answers, maxInFlightOverrides)
	if err != nil {
		return err
	}

	return nil
}

func getAnswersFromPlan(problems []boshdir.Problem, instanceGroupPlan InstanceGroupPlan) []boshdir.ProblemAnswer {
	var answers []boshdir.ProblemAnswer
	for _, p := range problems {
		resolutionName := instanceGroupPlan.resolutionName(p)
		answers = append(answers, boshdir.ProblemAnswer{
			ProblemID: p.ID,
			Resolution: boshdir.ProblemResolution{
				Name: &resolutionName,
				Plan: instanceGroupPlan.resolutionPlan(p),
			},
		})
	}

	return answers
}

func (c RecoverCmd) readPlan(opts RecoverOpts) (*RecoveryPlan, error) {
	planContents, err := c.fs.ReadFile(opts.Args.RecoveryPlan.ExpandedPath)
	if err != nil {
		return nil, err
	}

	var plan RecoveryPlan
	err = yaml.Unmarshal(planContents, &plan)
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (c RecoverCmd) printPlanSummary(problemsByInstanceGroup map[string][]boshdir.Problem, plan *RecoveryPlan) {
	for instanceGroup, instanceGroupProblems := range problemsByInstanceGroup {
		instanceGroupPlan := getPlanForInstanceGroup(instanceGroup, plan)

		title := fmt.Sprintf("Instance Group '%s' plan summary", instanceGroup)
		if instanceGroupPlan.MaxInFlightOverride != "" {
			title = fmt.Sprintf("%s (max_in_flight override: %s)", title, instanceGroupPlan.MaxInFlightOverride)
		}

		table := boshtbl.Table{
			Title: title,
			Header: []boshtbl.Header{
				boshtbl.NewHeader("#"),
				boshtbl.NewHeader("Planned resolution"),
				boshtbl.NewHeader("Description"),
			},
			SortBy: []boshtbl.ColumnSort{{Column: 0, Asc: true}},
		}

		for _, p := range instanceGroupProblems {
			table.Rows = append(table.Rows, []boshtbl.Value{
				boshtbl.NewValueInt(p.ID),
				boshtbl.NewValueString(instanceGroupPlan.resolutionPlan(p)),
				boshtbl.NewValueString(p.Description),
			})
		}

		c.ui.PrintTable(table)
	}
}

func getPlanForInstanceGroup(instanceGroup string, plan *RecoveryPlan) InstanceGroupPlan {
	for _, p := range plan.InstanceGroupsPlan {
		if p.Name == instanceGroup {
			return p
		}
	}

	return InstanceGroupPlan{}
}
