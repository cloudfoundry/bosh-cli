package fakes

import (
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type FakeDependencyAnalysis struct {
	DeterminePackageCompilationOrderSource []*bmrel.Package
	DeterminePackageCompilationOrderResult []*bmrel.Package
}

func NewFakeDependencyAnalysis() *FakeDependencyAnalysis {
	return &FakeDependencyAnalysis{}
}

func (fda *FakeDependencyAnalysis) DeterminePackageCompilationOrder(source []*bmrel.Package) ([]*bmrel.Package, error) {
	fda.DeterminePackageCompilationOrderSource = source
	return fda.DeterminePackageCompilationOrderResult, nil
}
