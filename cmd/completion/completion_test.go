package completion_test

import (
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-utils/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-cli/v7/cmd/completion"
)

var _ = Describe("Completion Integration Tests", func() {
	var (
		boshComplete *completion.BoshComplete
	)

	BeforeEach(func() {
		testLogger := logger.NewWriterLogger(logger.LevelInfo, os.Stderr)
		fakeCmdCtx := &completion.CmdContext{}
		fakeDq := completion.NewDirectorQueryFake(fakeCmdCtx)
		fakeCompletionFunctionMap := completion.NewCompleteFunctionsMap(testLogger, fakeDq)
		boshComplete = completion.NewBoshCompleteWithFunctions(testLogger, fakeCmdCtx, fakeCompletionFunctionMap)
	})

	It("is this bosh completion command", func() {
		Expect(completion.IsItCompletionCommand([]string{"completion"})).To(BeTrue())
		Expect(completion.IsItCompletionCommand([]string{"completion", "something-else"})).To(BeTrue())
		Expect(completion.IsItCompletionCommand([]string{"__complete"})).To(BeTrue())
		Expect(completion.IsItCompletionCommand([]string{"__complete", "something-else"})).To(BeTrue())
		Expect(completion.IsItCompletionCommand([]string{})).To(BeFalse())
		Expect(completion.IsItCompletionCommand([]string{"deployments"})).To(BeFalse())
		Expect(completion.IsItCompletionCommand([]string{"something-else"})).To(BeFalse())
	})

	It("completion", func() {
		testCase{args: "completion", wantRes: []string{"Generate the autocompletion script for bosh" +
			" for the specified shell."}, wantStartsWith: true}.check(boshComplete)
	})
	It("completion help", func() {
		testCase{args: "completion", wantRes: []string{"Generate the autocompletion script for bosh" +
			" for the specified shell."}, wantStartsWith: true}.check(boshComplete)
	})
	It("completion bash script", func() {
		testCase{args: "completion bash", wantRes: []string{"# bash completion V2 for bosh" +
			"                                 -*- shell-script -*-"}, wantStartsWith: true}.check(boshComplete)
	})
	It("completion -", func() {
		testCase{args: "__complete -", wantRes: filterCompletion(globalFlags, "-")}.check(boshComplete)
	})
	It("completion --", func() {
		testCase{args: "__complete --", wantRes: filterCompletion(globalFlags, "--")}.check(boshComplete)
	})
	It("completion --d", func() {
		testCase{args: "__complete --d", wantRes: filterCompletion(globalFlags, "--d")}.check(boshComplete)
	})
	It("completion d", func() {
		// format.MaxLength = 0
		testCase{args: "__complete -- d", wantRes: filterCompletion(globalCommands, "d")}.check(boshComplete)
	})
	It("completion with given deployment", func() {
		testCase{args: "__complete -d deployment-name d", wantRes: filterCompletion(globalCommands, "d")}.check(boshComplete)
	})

	It("completion depl", func() {
		testCase{args: "__complete -- depl", wantRes: filterCompletion(globalCommands, "depl")}.check(boshComplete)
	})
	It("completion deployments", func() {
		testCase{args: "__complete -- deployments", wantRes: filterCompletion(globalCommands, "deployments")}.check(boshComplete)
	})
	It("completion deployments ", func() {
		testCase{args: "__complete -- deployments ", wantRes: []string{":0", ""}}.check(boshComplete)
	})

	It("completion in", func() {
		testCase{args: "__complete in", wantRes: filterCompletion(globalCommands, "in")}.check(boshComplete)
	})
	It("completion instances flag --failing", func() {
		testCase{args: "__complete instances --fai", wantRes: []string{"--failing\tOnly show failing instances", ":4", ""}}.check(boshComplete)
	})

	It("completion curl /info", func() {
		testCase{args: "__complete curl /info", wantRes: []string{"/info", ":4", ""}}.check(boshComplete)
	})
	It("completion curl /task", func() {
		testCase{args: "__complete curl /task", wantRes: []string{"/tasks/", "/tasks/{}", ":4", ""}}.check(boshComplete)
		testCase{args: "__complete curl /tasks/", wantRes: []string{"/tasks/", "/tasks/{}", ":4", ""}}.check(boshComplete)
		testCase{args: "__complete curl /tasks/id", wantRes: []string{"/tasks/id/output", "/tasks/id/cancel", ":4", ""}}.check(boshComplete)
		testCase{args: "__complete curl /tasks/id/", wantRes: []string{"/tasks/id/output", "/tasks/id/cancel", ":4", ""}}.check(boshComplete)
		testCase{args: "__complete curl /tasks/id/o", wantRes: []string{"/tasks/id/output", ":4", ""}}.check(boshComplete)
	})

	It("completion deployments", func() {
		testCase{args: "__complete -d fake", wantRes: []string{"fake-d1", "fake-d2", ":4", ""}}.check(boshComplete)
		testCase{args: "__complete -d fake", wantRes: []string{"fake-d1", "fake-d2", ":4", ""}}.check(boshComplete)
	})
	It("completion instances groups", func() {
		testCase{args: "__complete -d d1 ssh fake",
			wantRes: []string{
				"fake-d1-g1", "fake-d1-g1/fake-i1", "fake-d1-g1/fake-i2",
				"fake-d1-g2", "fake-d1-g2/fake-i1", "fake-d1-g2/fake-i2",
				":4", ""}}.check(boshComplete)
		testCase{args: "__complete -d d1 ssh fake-d1-g1",
			wantRes: []string{"fake-d1-g1", "fake-d1-g1/fake-i1", "fake-d1-g1/fake-i2", ":4", ""}}.check(boshComplete)
		testCase{args: "__complete -d d1 ssh fake-d1-g1/",
			wantRes: []string{"fake-d1-g1/fake-i1", "fake-d1-g1/fake-i2", ":4", ""}}.check(boshComplete)
	})
	It("completion snapshots group/ids", func() {
		testCase{args: "__complete -d d2 snapshots fake-d2-g1",
			wantRes: []string{"fake-d2-g1/fake-i1", "fake-d2-g1/fake-i2", ":4", ""}}.check(boshComplete)
		testCase{args: "__complete -d d2 snapshots fake-d2-g1/fake-i2",
			wantRes: []string{"fake-d2-g1/fake-i2", ":4", ""}}.check(boshComplete)
	})

})

type testCase struct {
	args           string
	wantRes        []string
	wantStartsWith bool
	wantErr        bool
}

func (tc testCase) check(boshComplete *completion.BoshComplete) {
	args := strings.Split(tc.args, " ")
	result, err := boshComplete.ExecuteCaptured(args)
	if tc.wantErr {
		Expect(err).To(HaveOccurred())
	} else {
		Expect(err).ToNot(HaveOccurred())
	}
	if result == nil {
		Expect(result).ToNot(BeNil())
	}
	if tc.wantStartsWith {
		for i, wantLine := range tc.wantRes {
			Expect(wantLine).To(Equal(result.Lines[i]))
		}
	} else {
		Expect(tc.wantRes).To(Equal(result.Lines))
	}
}

var globalFlags = []string{
	"--ca-cert\tDirector CA certificate path or value, env: BOSH_CA_CERT",
	"--client\tOverride username or UAA client, env: BOSH_CLIENT",
	"--client-secret\tOverride password or UAA client secret, env: BOSH_CLIENT_SECRET",
	"--column\tFilter to show only given column(s)",
	"--config\tConfig file path, env: BOSH_CONFIG",
	"--deployment\tDeployment name, env: BOSH_DEPLOYMENT",
	"-d\tDeployment name, env: BOSH_DEPLOYMENT",
	"--environment\tDirector environment name or URL, env: BOSH_ENVIRONMENT",
	"-e\tDirector environment name or URL, env: BOSH_ENVIRONMENT",
	"--help\thelp for bosh",
	"-h\thelp for bosh",
	"--json\tOutput as JSON",
	"--no-color\tToggle colorized output",
	"--non-interactive\tDon't ask for user input, env: BOSH_NON_INTERACTIVE",
	"-n\tDon't ask for user input, env: BOSH_NON_INTERACTIVE",
	"--parallel\tThe max number of parallel operations",
	"--sha2\tUse SHA256 checksums, env: BOSH_SHA2",
	"--tty\tForce TTY-like output",
	"--version\tShow CLI version",
	"-v\tShow CLI version",
}
var globalCommands = []string{
	"add-blob\tAdd blob",
	"alias-env\tAlias environment to save URL and CA certificate",
	"attach-disk\tAttach disk to an instance",
	"blobs\tList blobs",
	"cancel-task\tCancel task at its next checkpoint",
	"cancel-tasks\tCancel tasks at their next checkpoints",
	"clean-up\tClean up old unused resources except orphaned disks",
	"cloud-check\tCloud consistency check and interactive repair",
	"cloud-config\tShow current cloud config",
	"completion\tGenerate the autocompletion script for the specified shell",
	"config\tShow current config for either ID or both type and name",
	"configs\tList configs",
	"cpi-config\tShow current CPI config",
	"create-env\tCreate or update BOSH environment",
	"create-recovery-plan\tInteractively generate a recovery plan for disaster repair",
	"create-release\tCreate release",
	"curl\tMake an HTTP request to the Director",
	"delete-config\tDelete config",
	"delete-deployment\tDelete deployment",
	"delete-disk\tDelete disk",
	"delete-env\tDelete BOSH environment",
	"delete-network\tDelete network",
	"delete-release\tDelete release",
	"delete-snapshot\tDelete snapshot",
	"delete-snapshots\tDelete all snapshots in a deployment",
	"delete-stemcell\tDelete stemcell",
	"delete-vm\tDelete VM",
	"deploy\tUpdate deployment",
	"deployment\tShow deployment information",
	"deployments\tList deployments",
	"diff-config\tDiff two configs by ID or content",
	"disks\tList disks",
	"environment\tShow environment",
	"environments\tList environments",
	"errands\tList errands",
	"event\tShow event details",
	"events\tList events",
	"export-release\tExport the compiled release to a tarball",
	"finalize-release\tCreate final release from dev release tarball",
	"generate-job\tGenerate job",
	"generate-package\tGenerate package",
	"help\tShow this help message",
	"help\tHelp about any command",
	"ignore\tIgnore an instance",
	"init-release\tInitialize release",
	"inspect-local-release\tDisplay information from release metadata",
	"inspect-local-stemcell\tDisplay information from stemcell metadata",
	"inspect-release\tList release contents such as jobs",
	"instances\tList all instances in a deployment",
	"interpolate\tInterpolates variables into a manifest",
	"locks\tList current locks",
	"log-in\tLog in",
	"log-out\tLog out",
	"logs\tFetch logs from instance(s)",
	"manifest\tShow deployment manifest",
	"networks\tList networks",
	"orphan-disk\tOrphan disk",
	"orphaned-vms\tList all the orphaned VMs in all deployments",
	"recover\tApply a recovery plan for disaster repair",
	"recreate\tRecreate instance(s)",
	"releases\tList releases",
	"remove-blob\tRemove blob",
	"repack-stemcell\tRepack stemcell",
	"reset-release\tReset release",
	"restart\tRestart instance(s)",
	"run-errand\tRun errand",
	"runtime-config\tShow current runtime config",
	"scp\tSCP to/from instance(s)",
	"sha1ify-release\tConvert release tarball to use SHA1",
	"sha2ify-release\tConvert release tarball to use SHA256",
	"snapshots\tList snapshots",
	"ssh\tSSH into instance(s)",
	"start\tStart instance(s)",
	"start-env\tStart BOSH environment",
	"stemcells\tList stemcells",
	"stop\tStop instance(s)",
	"stop-env\tStop BOSH environment",
	"sync-blobs\tSync blobs",
	"take-snapshot\tTake snapshot",
	"task\tShow task status and start tracking its output",
	"tasks\tList running or recent tasks",
	"unalias-env\tRemove an aliased environment",
	"unignore\tUnignore an instance",
	"update-cloud-config\tUpdate current cloud config",
	"update-config\tUpdate config",
	"update-cpi-config\tUpdate current CPI config",
	"update-resurrection\tEnable/disable resurrection",
	"update-runtime-config\tUpdate current runtime config",
	"upload-blobs\tUpload blobs",
	"upload-release\tUpload release",
	"upload-stemcell\tUpload stemcell",
	"variables\tList variables",
	"vendor-package\tVendor package",
	"vms\tList all VMs in all deployments",
}

func filterCompletion(src []string, prefix string) []string {
	var result []string
	for _, s := range src {
		if strings.HasPrefix(s, prefix) {
			result = append(result, s)
		}
	}
	result = append(result, ":4")
	result = append(result, "")
	return result
}
