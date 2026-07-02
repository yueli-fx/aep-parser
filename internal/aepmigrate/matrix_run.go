package aepmigrate

import "fmt"

type matrixRunPlan struct {
	recipePaths  []string
	sourceLabels []string
	targetLabels []string
	aeOpenLabels []string
	hosts        map[string]string
}

func prepareMatrixRun(opts MatrixOptions) (matrixRunPlan, error) {
	if opts.OutRoot == "" {
		return matrixRunPlan{}, fmt.Errorf("matrix out root is required")
	}
	recipePaths, err := matrixRecipePaths(opts)
	if err != nil {
		return matrixRunPlan{}, err
	}
	if len(recipePaths) == 0 {
		return matrixRunPlan{}, fmt.Errorf("matrix requires at least one recipe")
	}
	plan := matrixRunPlan{
		recipePaths:  recipePaths,
		sourceLabels: expandMatrixSourceLabels(opts.SourceLabels),
		targetLabels: expandMatrixTargetLabels(opts.TargetLabels),
		aeOpenLabels: matrixAEOpenLabels(opts.AEOpenLabels),
		hosts:        mergeHostMaps(BuildAEHostMap(opts.AEInstallRoot), opts.AEHosts),
	}
	if len(plan.sourceLabels) == 0 {
		plan.sourceLabels = []string{"recipe"}
	}
	if len(plan.targetLabels) == 0 {
		plan.targetLabels = []string{string(VersionAE2025)}
	}
	if opts.AEOpen && opts.MaxAEOpenCases > 0 {
		caseCount := matrixRunCaseCount(opts, plan)
		if caseCount > opts.MaxAEOpenCases {
			return matrixRunPlan{}, fmt.Errorf("AE open matrix would run %d cases, above limit %d; narrow recipes/targets or set a higher -max-ae-open-cases", caseCount, opts.MaxAEOpenCases)
		}
	}
	return plan, nil
}

func matrixRunCaseCount(opts MatrixOptions, plan matrixRunPlan) int {
	caseCount := len(plan.recipePaths) * len(plan.sourceLabels) * len(plan.targetLabels)
	if opts.AEOpen && len(plan.aeOpenLabels) > 0 {
		caseCount *= len(plan.aeOpenLabels)
	}
	return caseCount
}

func runMatrixCases(opts MatrixOptions, plan matrixRunPlan) []MatrixCase {
	var cases []MatrixCase
	for _, recipePath := range plan.recipePaths {
		for _, sourceLabel := range plan.sourceLabels {
			for _, targetLabel := range plan.targetLabels {
				if opts.AEOpen && len(plan.aeOpenLabels) > 0 {
					for _, aeOpenLabel := range plan.aeOpenLabels {
						c := runMatrixCase(opts, plan.hosts, recipePath, sourceLabel, targetLabel, aeOpenLabel, true)
						cases = append(cases, c)
					}
					continue
				}
				c := runMatrixCase(opts, plan.hosts, recipePath, sourceLabel, targetLabel, "", false)
				cases = append(cases, c)
			}
		}
	}
	return cases
}
