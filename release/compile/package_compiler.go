package compile

import (
	boshsys "github.com/cloudfoundry/bosh-agent/system"
	bmrel "github.com/cloudfoundry/bosh-micro-cli/release"
)

type PackageCompiler interface {
	Compile(*bmrel.Package) error
}

type packageCompiler struct {
	runner boshsys.CmdRunner
}

func (pc *packageCompiler) Compile(pkg *bmrel.Package) error {
	packageSrcDir := pkg.ExtractedPath
	enablePath := "/fake-dir/packages/pkg_name"
	cmd := boshsys.Command{
		Name: "bash",
		Args: []string{"-x", "packaging"},
		Env: map[string]string{
			"BOSH_COMPILE_TARGET":  packageSrcDir,
			"BOSH_INSTALL_TARGET":  enablePath,
			"BOSH_PACKAGE_NAME":    pkg.Name,
			"BOSH_PACKAGE_VERSION": pkg.Version,
			"BOSH_PACKAGES_DIR":    "/fake-packages-dir/",
		},
		WorkingDir: packageSrcDir,
	}
	pc.runner.RunComplexCommand(cmd)

	return nil
}

func NewPackageCompiler(runner boshsys.CmdRunner) PackageCompiler {
	return &packageCompiler{
		runner: runner,
	}
}
