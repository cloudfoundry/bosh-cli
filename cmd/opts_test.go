package cmd_test

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/bosh-cli/cmd"
)

func getStructTagForName(field string, opts interface{}) string {
	st, _ := reflect.TypeOf(opts).Elem().FieldByName(field)

	return string(st.Tag)
}

var _ = Describe("Opts", func() {
	Describe("BoshOpts", func() {

		var opts *BoshOpts

		BeforeEach(func() {
			opts = &BoshOpts{}
		})

		It("long or short command options do not shadow global opts", func() {
			globalLongOptNames := []string{}
			globalShortOptNames := []string{}
			cmdOptss := []reflect.Value{}

			optsStruct := reflect.TypeOf(opts).Elem()

			for i := 0; i < optsStruct.NumField(); i++ {
				field := optsStruct.Field(i)
				if field.Tag.Get("long") != "" {
					globalLongOptNames = append(globalLongOptNames, field.Tag.Get("long"))
				}
				if field.Tag.Get("short") != "" {
					globalShortOptNames = append(globalShortOptNames, field.Tag.Get("short"))
				}
				if field.Tag.Get("command") != "" {
					m := reflect.ValueOf(opts).Elem().Field(i).Addr()
					cmdOptss = append(cmdOptss, m)
				}
			}

			var errs []string

			for _, optName := range globalLongOptNames {
				for _, cmdOpts := range cmdOptss {
					cmdOptsStruct := reflect.TypeOf(cmdOpts.Interface()).Elem()

					for i := 0; i < cmdOptsStruct.NumField(); i++ {
						field := cmdOptsStruct.Field(i)

						if field.Tag.Get("long") == optName {
							errs = append(errs, fmt.Sprintf("Command '%s' shadows global long option '%s'", cmdOptsStruct.Name(), optName))
						}
					}
				}
			}

			for _, optName := range globalShortOptNames {
				for _, cmdOpts := range cmdOptss {
					cmdOptsStruct := reflect.TypeOf(cmdOpts.Interface()).Elem()

					for i := 0; i < cmdOptsStruct.NumField(); i++ {
						field := cmdOptsStruct.Field(i)

						if field.Tag.Get("short") == optName {
							errs = append(errs, fmt.Sprintf("Command '%s' shadows global short option '%s'", cmdOptsStruct.Name(), optName))
						}
					}
				}
			}

			// --version flag is a bit awkward so let's ignore conflicts
			Expect(errs).To(Equal([]string{
				"Command 'UploadStemcellOpts' shadows global long option 'version'",
				"Command 'UploadReleaseOpts' shadows global long option 'version'",
				"Command 'CreateReleaseOpts' shadows global long option 'version'",
				"Command 'FinalizeReleaseOpts' shadows global long option 'version'",
			}))
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("VersionOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("VersionOpt", opts)).To(Equal(
						`long:"version" short:"v" description:"Show CLI version"`,
					))
				})
			})

			Describe("ConfigPathOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("ConfigPathOpt", opts)).To(Equal(
						`long:"config" description:"Config file path" env:"BOSH_CONFIG" default:"~/.bosh/config"`,
					))
				})
			})

			Describe("EnvironmentOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("EnvironmentOpt", opts)).To(Equal(
						`long:"environment" short:"e" description:"Director environment name or URL" env:"BOSH_ENVIRONMENT"`,
					))
				})
			})

			Describe("CACertOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CACertOpt", opts)).To(Equal(
						`long:"ca-cert"               description:"Director CA certificate path or value" env:"BOSH_CA_CERT"`,
					))
				})
			})

			Describe("UsernameOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UsernameOpt", opts)).To(Equal(
						`long:"user"     description:"Override username" env:"BOSH_USER"`,
					))
				})
			})

			Describe("PasswordOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("PasswordOpt", opts)).To(Equal(
						`long:"password" description:"Override password" env:"BOSH_PASSWORD"`,
					))
				})
			})

			Describe("UAAClientOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UAAClientOpt", opts)).To(Equal(
						`long:"uaa-client"        description:"Override UAA client"        env:"BOSH_CLIENT"`,
					))
				})
			})

			Describe("UAAClientSecretOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UAAClientSecretOpt", opts)).To(Equal(
						`long:"uaa-client-secret" description:"Override UAA client secret" env:"BOSH_CLIENT_SECRET"`,
					))
				})
			})

			Describe("DeploymentOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeploymentOpt", opts)).To(Equal(
						`long:"deployment" short:"d" description:"Deployment name" env:"BOSH_DEPLOYMENT"`,
					))
				})
			})

			Describe("JSONOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("JSONOpt", opts)).To(Equal(
						`long:"json"                      description:"Output as JSON"`,
					))
				})
			})

			Describe("TTYOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("TTYOpt", opts)).To(Equal(
						`long:"tty"                       description:"Force TTY-like output"`,
					))
				})
			})

			Describe("NoColorOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("NoColorOpt", opts)).To(Equal(
						`long:"no-color"                  description:"Toggle colorized output"`,
					))
				})
			})

			Describe("NonInteractiveOpt", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("NonInteractiveOpt", opts)).To(Equal(
						`long:"non-interactive" short:"n" description:"Don't ask for user input"`,
					))
				})
			})

			Describe("CreateEnv", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CreateEnv", opts)).To(Equal(
						`command:"create-env" description:"Create or update BOSH environment"`,
					))
				})
			})

			Describe("DeleteEnv", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeleteEnv", opts)).To(Equal(
						`command:"delete-env" description:"Delete BOSH environment"`,
					))
				})
			})

			Describe("Environment", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Environment", opts)).To(Equal(
						`command:"environment" alias:"env" description:"Set or show current environment"`,
					))
				})
			})

			Describe("Environments", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Environments", opts)).To(Equal(
						`command:"environments" alias:"envs" description:"List environments"`,
					))
				})
			})

			Describe("LogIn", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("LogIn", opts)).To(Equal(
						`command:"log-in"  alias:"l" description:"Log in"`,
					))
				})
			})

			Describe("LogOut", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("LogOut", opts)).To(Equal(
						`command:"log-out"           description:"Forget saved credentials for Director in the current environment"`,
					))
				})
			})

			Describe("Task", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Task", opts)).To(Equal(
						`command:"task"        alias:"t"  description:"Show task status and start tracking its output"`,
					))
				})
			})

			Describe("Tasks", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Tasks", opts)).To(Equal(
						`command:"tasks"       alias:"ts" description:"List running or recent tasks"`,
					))
				})
			})

			Describe("CancelTask", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CancelTask", opts)).To(Equal(
						`command:"cancel-task" alias:"ct" description:"Cancel task at its next checkpoint"`,
					))
				})
			})

			Describe("Locks", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Locks", opts)).To(Equal(
						`command:"locks"    alias:"ls" description:"List current locks"`,
					))
				})
			})

			Describe("CleanUp", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CleanUp", opts)).To(Equal(
						`command:"clean-up" alias:"cl" description:"Clean up releases, stemcells, disks, etc."`,
					))
				})
			})

			Describe("BackUp", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("BackUp", opts)).To(Equal(
						`command:"back-up"  alias:"bu" description:"Backup the Director to a tarball"`,
					))
				})
			})

			Describe("BuildManifest", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("BuildManifest", opts)).To(Equal(
						`command:"build-manifest"  alias:"bm" hidden:"yes" description:"Interpolates variables into a manifest template."`,
					))
				})
			})

			Describe("CloudConfig", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CloudConfig", opts)).To(Equal(
						`command:"cloud-config"        alias:"cc"  description:"Show current cloud config"`,
					))
				})
			})

			Describe("UpdateCloudConfig", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UpdateCloudConfig", opts)).To(Equal(
						`command:"update-cloud-config" alias:"ucc" description:"Update current cloud config"`,
					))
				})
			})

			Describe("RuntimeConfig", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("RuntimeConfig", opts)).To(Equal(
						`command:"runtime-config"        alias:"rc"  description:"Show current runtime config"`,
					))
				})
			})

			Describe("UpdateRuntimeConfig", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UpdateRuntimeConfig", opts)).To(Equal(
						`command:"update-runtime-config" alias:"urc" description:"Update current runtime config"`,
					))
				})
			})

			Describe("Deployment", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Deployment", opts)).To(Equal(
						`command:"deployment"        alias:"dep"             description:"Set or show current deployment"`,
					))
				})
			})

			Describe("Deployments", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Deployments", opts)).To(Equal(
						`command:"deployments"       alias:"ds" alias:"deps" description:"List deployments"`,
					))
				})
			})

			Describe("DeleteDeployment", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeleteDeployment", opts)).To(Equal(
						`command:"delete-deployment" alias:"deld"            description:"Delete deployment"`,
					))
				})
			})

			Describe("Deploy", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Deploy", opts)).To(Equal(
						`command:"deploy"   alias:"d"                                       description:"Deploy according to the currently selected deployment manifest"`,
					))
				})
			})

			Describe("Manifest", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Manifest", opts)).To(Equal(
						`command:"manifest" alias:"m" alias:"man" alias:"download-manifest" description:"Download deployment manifest locally"`,
					))
				})
			})

			Describe("Stemcells", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Stemcells", opts)).To(Equal(
						`command:"stemcells"       alias:"ss" alias:"stems" description:"List stemcells"`,
					))
				})
			})

			Describe("UploadStemcell", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UploadStemcell", opts)).To(Equal(
						`command:"upload-stemcell" alias:"us"               description:"Upload stemcell"`,
					))
				})
			})

			Describe("DeleteStemcell", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeleteStemcell", opts)).To(Equal(
						`command:"delete-stemcell" alias:"dels"             description:"Delete stemcell"`,
					))
				})
			})

			Describe("Releases", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Releases", opts)).To(Equal(
						`command:"releases"        alias:"rs" alias:"rels" description:"List releases"`,
					))
				})
			})

			Describe("UploadRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UploadRelease", opts)).To(Equal(
						`command:"upload-release"  alias:"ur"              description:"Upload release"`,
					))
				})
			})

			Describe("ExportRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("ExportRelease", opts)).To(Equal(
						`command:"export-release"  alias:"expr"            description:"Export the compiled release to a tarball"`,
					))
				})
			})

			Describe("InspectRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("InspectRelease", opts)).To(Equal(
						`command:"inspect-release" alias:"insr"            description:"List all jobs, packages, and compiled packages associated with a release"`,
					))
				})
			})

			Describe("DeleteRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeleteRelease", opts)).To(Equal(
						`command:"delete-release"  alias:"delr"            description:"Delete release"`,
					))
				})
			})

			Describe("Errands", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Errands", opts)).To(Equal(
						`command:"errands"    alias:"es" alias:"errs" description:"List errands"`,
					))
				})
			})

			Describe("RunErrand", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("RunErrand", opts)).To(Equal(
						`command:"run-errand" alias:"re"              description:"Run errand"`,
					))
				})
			})

			Describe("Disks", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Disks", opts)).To(Equal(
						`command:"disks"       description:"List disks"`,
					))
				})
			})

			Describe("DeleteDisk", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeleteDisk", opts)).To(Equal(
						`command:"delete-disk" description:"Delete disk"`,
					))
				})
			})

			Describe("Snapshots", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Snapshots", opts)).To(Equal(
						`command:"snapshots"        alias:"snaps"    description:"List snapshots"`,
					))
				})
			})

			Describe("TakeSnapshot", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("TakeSnapshot", opts)).To(Equal(
						`command:"take-snapshot"    alias:"tsnap"    description:"Take snapshot"`,
					))
				})
			})

			Describe("DeleteSnapshot", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeleteSnapshot", opts)).To(Equal(
						`command:"delete-snapshot"  alias:"delsnap"  description:"Delete snapshot"`,
					))
				})
			})

			Describe("DeleteSnapshots", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DeleteSnapshots", opts)).To(Equal(
						`command:"delete-snapshots" alias:"delsnaps" description:"Delete all snapshots in a deployment"`,
					))
				})
			})

			Describe("Instances", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Instances", opts)).To(Equal(
						`command:"instances"       alias:"is" alias:"ins"         description:"List all instances in a deployment"`,
					))
				})
			})

			Describe("VMs", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("VMs", opts)).To(Equal(
						`command:"vms"                                            description:"List all VMs in all deployments"`,
					))
				})
			})

			Describe("UpdateResurrection", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UpdateResurrection", opts)).To(Equal(
						`command:"update-resurrection"                            description:"Enable/disable resurrection"`,
					))
				})
			})

			Describe("CloudCheck", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CloudCheck", opts)).To(Equal(
						`command:"cloud-check"     alias:"cck" alias:"cloudcheck" description:"Cloud consistency check and interactive repair"`,
					))
				})
			})

			Describe("Logs", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Logs", opts)).To(Equal(
						`command:"logs"     description:"Fetch logs from instance(s)"`,
					))
				})
			})

			Describe("Start", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Start", opts)).To(Equal(
						`command:"start"    description:"Start instance(s)"`,
					))
				})
			})

			Describe("Stop", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Stop", opts)).To(Equal(
						`command:"stop"     description:"Stop instance(s)"`,
					))
				})
			})

			Describe("Restart", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Restart", opts)).To(Equal(
						`command:"restart"  description:"Restart instance(s)"`,
					))
				})
			})

			Describe("Recreate", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Recreate", opts)).To(Equal(
						`command:"recreate" description:"Recreate instance(s)"`,
					))
				})
			})

			Describe("SSH", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SSH", opts)).To(Equal(
						`command:"ssh" description:"SSH into instance(s)"`,
					))
				})
			})

			Describe("SCP", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SCP", opts)).To(Equal(
						`command:"scp" description:"SCP to/from instance(s)"`,
					))
				})
			})

			Describe("InitRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("InitRelease", opts)).To(Equal(
						`command:"init-release"                  description:"Initialize release"`,
					))
				})
			})

			Describe("ResetRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("ResetRelease", opts)).To(Equal(
						`command:"reset-release"                 description:"Reset release"`,
					))
				})
			})

			Describe("GenerateJob", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("GenerateJob", opts)).To(Equal(
						`command:"generate-job"     alias:"genj" description:"Generate job"`,
					))
				})
			})

			Describe("GeneratePackage", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("GeneratePackage", opts)).To(Equal(
						`command:"generate-package" alias:"genp" description:"Generate package"`,
					))
				})
			})

			Describe("CreateRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CreateRelease", opts)).To(Equal(
						`command:"create-release"   alias:"cr"   description:"Create release"`,
					))
				})
			})

			Describe("FinalizeRelease", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("FinalizeRelease", opts)).To(Equal(
						`command:"finalize-release" alias:"finr" description:"Create final release from dev release tarball"`,
					))
				})
			})

			Describe("Blobs", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Blobs", opts)).To(Equal(
						`command:"blobs"        alias:"bls"  description:"List blobs"`,
					))
				})
			})

			Describe("AddBlob", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("AddBlob", opts)).To(Equal(
						`command:"add-blob"     alias:"abl"  description:"Add blob"`,
					))
				})
			})

			Describe("RemoveBlob", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("RemoveBlob", opts)).To(Equal(
						`command:"remove-blob"  alias:"rmbl" description:"Remove blob"`,
					))
				})
			})

			Describe("SyncBlobs", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SyncBlobs", opts)).To(Equal(
						`command:"sync-blobs"   alias:"sbls" description:"Sync blobs"`,
					))
				})
			})

			Describe("UploadBlobs", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("UploadBlobs", opts)).To(Equal(
						`command:"upload-blobs" alias:"ubls" description:"Upload blobs"`,
					))
				})
			})
		})
	})

	Describe("CreateEnvOpts", func() {
		var opts *CreateEnvOpts

		BeforeEach(func() {
			opts = &CreateEnvOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("CreateEnvArgs", func() {
		var args *CreateEnvArgs

		BeforeEach(func() {
			args = &CreateEnvArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Manifest", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Manifest", args)).To(Equal(
						`positional-arg-name:"PATH" description:"Path to a manifest file"`,
					))
				})
			})

		})
	})

	Describe("DeleteEnvOpts", func() {
		var opts *DeleteEnvOpts

		BeforeEach(func() {
			opts = &DeleteEnvOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("DeleteEnvArgs", func() {
		var args *DeleteEnvArgs

		BeforeEach(func() {
			args = &DeleteEnvArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Manifest", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Manifest", args)).To(Equal(
						`positional-arg-name:"PATH" description:"Path to a manifest file"`,
					))
				})
			})

		})
	})

	Describe("EnvironmentOpts", func() {
		var opts *EnvironmentOpts

		BeforeEach(func() {
			opts = &EnvironmentOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

		})
	})

	Describe("EnvironmentArgs", func() {
		var args *EnvironmentArgs

		BeforeEach(func() {
			args = &EnvironmentArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("URL", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("URL", args)).To(Equal(
						`positional-arg-name:"URL"   description:"Director URL (e.g.: https://192.168.50.4:25555 or 192.168.50.4)"`,
					))
				})
			})

			Describe("Alias", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Alias", args)).To(Equal(
						`positional-arg-name:"ALIAS" description:"Environment alias"`,
					))
				})
			})

		})
	})

	Describe("TaskOpts", func() {
		var opts *TaskOpts

		BeforeEach(func() {
			opts = &TaskOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("Event", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Event", opts)).To(Equal(
						`long:"event"  description:"Track event log"`,
					))
				})
			})

			Describe("CPI", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CPI", opts)).To(Equal(
						`long:"cpi"    description:"Track CPI log"`,
					))
				})
			})

			Describe("Debug", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Debug", opts)).To(Equal(
						`long:"debug"  description:"Track debug log"`,
					))
				})
			})

			Describe("Result", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Result", opts)).To(Equal(
						`long:"result" description:"Track result log"`,
					))
				})
			})

			Describe("All", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("All", opts)).To(Equal(
						`long:"all" short:"a" description:"Include all task types (ssh, logs, vms, etc)"`,
					))
				})
			})
		})
	})

	Describe("TaskArgs", func() {
		var opts *TaskArgs

		BeforeEach(func() {
			opts = &TaskArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("ID", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("ID", opts)).To(Equal(
						`positional-arg-name:"ID"`,
					))
				})
			})
		})
	})

	Describe("TasksOpts", func() {
		var opts *TasksOpts

		BeforeEach(func() {
			opts = &TasksOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Recent", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Recent", opts)).To(Equal(
						`long:"recent" short:"r" description:"Number of tasks to show" optional:"true" optional-value:"30"`,
					))
				})
			})

			Describe("All", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("All", opts)).To(Equal(
						`long:"all" short:"a" description:"Include all task types (ssh, logs, vms, etc)"`,
					))
				})
			})

		})
	})

	Describe("CancelTaskOpts", func() {
		var opts *CancelTaskOpts

		BeforeEach(func() {
			opts = &CancelTaskOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("CleanUpOpts", func() {
		var opts *CleanUpOpts

		BeforeEach(func() {
			opts = &CleanUpOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("All", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("All", opts)).To(Equal(
						`long:"all" description:"Remove all unused releases, stemcells, etc.; otherwise most recent resources will be kept"`,
					))
				})
			})

		})
	})

	Describe("BackUpOpts", func() {
		var opts *BackUpOpts

		BeforeEach(func() {
			opts = &BackUpOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force" description:"Overwrite if the backup file already exists"`,
					))
				})
			})

		})
	})

	Describe("BackUpArgs", func() {
		var opts *BackUpArgs

		BeforeEach(func() {
			opts = &BackUpArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Path", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Path", opts)).To(Equal(
						`positional-arg-name:"PATH"`,
					))
				})
			})

		})
	})

	Describe("BuildManifestOpts", func() {
		var opts *BuildManifestOpts

		BeforeEach(func() {
			opts = &BuildManifestOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("BuildManifestArgs", func() {
		var opts *BuildManifestArgs

		BeforeEach(func() {
			opts = &BuildManifestArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Manifest", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Manifest", opts)).To(Equal(
						`positional-arg-name:"PATH" description:"Path to a template that will be interpolated"`,
					))
				})
			})

		})
	})

	Describe("UpdateCloudConfigOpts", func() {
		var opts *UpdateCloudConfigOpts

		BeforeEach(func() {
			opts = &UpdateCloudConfigOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("UpdateCloudConfigArgs", func() {
		var opts *UpdateCloudConfigArgs

		BeforeEach(func() {
			opts = &UpdateCloudConfigArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("CloudConfig", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CloudConfig", opts)).To(Equal(
						`positional-arg-name:"PATH" description:"Path to a cloud config file"`,
					))
				})
			})

		})
	})

	Describe("UpdateRuntimeConfigOpts", func() {
		var opts *UpdateRuntimeConfigOpts

		BeforeEach(func() {
			opts = &UpdateRuntimeConfigOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("UpdateRuntimeConfigArgs", func() {
		var opts *UpdateRuntimeConfigArgs

		BeforeEach(func() {
			opts = &UpdateRuntimeConfigArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("RuntimeConfig", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("RuntimeConfig", opts)).To(Equal(
						`positional-arg-name:"PATH" description:"Path to a runtime config file"`,
					))
				})
			})

		})
	})

	Describe("DeploymentOpts", func() {
		var opts *DeploymentOpts

		BeforeEach(func() {
			opts = &DeploymentOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

		})
	})

	Describe("DeployOpts", func() {
		var opts *DeployOpts

		BeforeEach(func() {
			opts = &DeployOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Recreate", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Recreate", opts)).To(Equal(
						`long:"recreate"   description:"Recreate all VMs in deployment"`,
					))
				})
			})

			Describe("NoRedact", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("NoRedact", opts)).To(Equal(
						`long:"no-redact" description:"Show non-redacted manifest diff"`,
					))
				})
			})

			Describe("SkipDrain", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SkipDrain", opts)).To(Equal(
						`long:"skip-drain" description:"Skip running drain scripts"`,
					))
				})
			})

		})
	})

	Describe("DeployArgs", func() {
		var opts *DeployArgs

		BeforeEach(func() {
			opts = &DeployArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Manifest", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Manifest", opts)).To(Equal(
						`positional-arg-name:"PATH" description:"Path to a manifest file"`,
					))
				})
			})

		})
	})

	Describe("DeleteDeploymentOpts", func() {
		var opts *DeleteDeploymentOpts

		BeforeEach(func() {
			opts = &DeleteDeploymentOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force" description:"Ignore errors"`,
					))
				})
			})

		})
	})

	Describe("DeploymentArgs", func() {
		var opts *DeploymentArgs

		BeforeEach(func() {
			opts = &DeploymentArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("NameOrPath", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("NameOrPath", opts)).To(Equal(
						`positional-arg-name:"NAME"`,
					))
				})
			})

		})
	})

	Describe("DeploymentNameArgs", func() {
		var opts *DeploymentNameArgs

		BeforeEach(func() {
			opts = &DeploymentNameArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`positional-arg-name:"NAME"`,
					))
				})
			})

		})
	})

	Describe("UploadStemcellOpts", func() {
		var opts *UploadStemcellOpts

		BeforeEach(func() {
			opts = &UploadStemcellOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Fix", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Fix", opts)).To(Equal(
						`long:"fix" description:"Replaces the stemcell if already exists"`,
					))
				})
			})

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`long:"name"     description:"Name used in existence check (is not used with local stemcell file)"`,
					))
				})
			})

			Describe("Version", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Version", opts)).To(Equal(
						`long:"version"  description:"Version used in existence check (is not used with local stemcell file)"`,
					))
				})
			})

			Describe("SHA1", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SHA1", opts)).To(Equal(
						`long:"sha1" description:"SHA1 of the remote stemcell (is not used with local files)"`,
					))
				})
			})

		})
	})

	Describe("UploadStemcellArgs", func() {
		var opts *UploadStemcellArgs

		BeforeEach(func() {
			opts = &UploadStemcellArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("URL", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("URL", opts)).To(Equal(
						`positional-arg-name:"URL" description:"Path to a local file or URL"`,
					))
				})
			})

		})
	})

	Describe("DeleteStemcellOpts", func() {
		var opts *DeleteStemcellOpts

		BeforeEach(func() {
			opts = &DeleteStemcellOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force" description:"Ignore errors"`,
					))
				})
			})

		})
	})

	Describe("DeleteStemcellArgs", func() {
		var opts *DeleteStemcellArgs

		BeforeEach(func() {
			opts = &DeleteStemcellArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Slug", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Slug", opts)).To(Equal(
						`positional-arg-name:"NAME/VERSION"`,
					))
				})
			})

		})
	})

	Describe("UploadReleaseOpts", func() {
		var opts *UploadReleaseOpts

		BeforeEach(func() {
			opts = &UploadReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"zzzRelease directory path if not current working directory" default:"."`,
					))
				})
			})

			Describe("Rebase", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Rebase", opts)).To(Equal(
						`long:"rebase" description:"Rebases this release onto the latest version known by the Director"`,
					))
				})
			})

			Describe("Fix", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Fix", opts)).To(Equal(
						`long:"fix" description:"Replaces corrupt and missing jobs and packages"`,
					))
				})
			})

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`long:"name"     description:"Name used in existence check (is not used with local release file)"`,
					))
				})
			})

			Describe("Version", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Version", opts)).To(Equal(
						`long:"version"  description:"Version used in existence check (is not used with local release file)"`,
					))
				})
			})

			Describe("SHA1", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SHA1", opts)).To(Equal(
						`long:"sha1" description:"SHA1 of the remote release (is not used with local files)"`,
					))
				})
			})

		})
	})

	Describe("UploadReleaseArgs", func() {
		var opts *UploadReleaseArgs

		BeforeEach(func() {
			opts = &UploadReleaseArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("URL", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("URL", opts)).To(Equal(
						`positional-arg-name:"URL" description:"Path to a local file or URL"`,
					))
				})
			})

		})
	})

	Describe("DeleteReleaseOpts", func() {
		var opts *DeleteReleaseOpts

		BeforeEach(func() {
			opts = &DeleteReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force" description:"Ignore errors"`,
					))
				})
			})

		})
	})

	Describe("DeleteReleaseArgs", func() {
		var opts *DeleteReleaseArgs

		BeforeEach(func() {
			opts = &DeleteReleaseArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Slug", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Slug", opts)).To(Equal(
						`positional-arg-name:"NAME[/VERSION]"`,
					))
				})
			})

		})
	})

	Describe("ExportReleaseOpts", func() {
		var opts *ExportReleaseOpts

		BeforeEach(func() {
			opts = &ExportReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Destination directory" default:"."`,
					))
				})
			})

		})
	})

	Describe("ExportReleaseArgs", func() {
		var opts *ExportReleaseArgs

		BeforeEach(func() {
			opts = &ExportReleaseArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("ReleaseSlug", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("ReleaseSlug", opts)).To(Equal(
						`positional-arg-name:"NAME/VERSION"`,
					))
				})
			})

			Describe("OSVersionSlug", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("OSVersionSlug", opts)).To(Equal(
						`positional-arg-name:"OS/VERSION"`,
					))
				})
			})

		})
	})

	Describe("InspectReleaseOpts", func() {
		var opts *InspectReleaseOpts

		BeforeEach(func() {
			opts = &InspectReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("InspectReleaseArgs", func() {
		var opts *InspectReleaseArgs

		BeforeEach(func() {
			opts = &InspectReleaseArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Slug", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Slug", opts)).To(Equal(
						`positional-arg-name:"NAME/VERSION"`,
					))
				})
			})

		})
	})

	Describe("RunErrandOpts", func() {
		var opts *RunErrandOpts

		BeforeEach(func() {
			opts = &RunErrandOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("KeepAlive", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("KeepAlive", opts)).To(Equal(
						`long:"keep-alive" description:"Use existing VM to run an errand and keep it after completion"`,
					))
				})
			})

			Describe("DownloadLogs", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DownloadLogs", opts)).To(Equal(
						`long:"download-logs" description:"Download logs"`,
					))
				})
			})

			Describe("LogsDirectory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("LogsDirectory", opts)).To(Equal(
						`long:"logs-dir" description:"Destination directory for logs" default:"."`,
					))
				})
			})

		})
	})

	Describe("RunErrandArgs", func() {
		var opts *RunErrandArgs

		BeforeEach(func() {
			opts = &RunErrandArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`positional-arg-name:"NAME"`,
					))
				})
			})

		})
	})

	Describe("DisksOpts", func() {
		var opts *DisksOpts

		BeforeEach(func() {
			opts = &DisksOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Orphaned", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Orphaned", opts)).To(Equal(
						`long:"orphaned" short:"o" description:"List orphaned disks"`,
					))
				})
			})

		})
	})

	Describe("DeleteDiskOpts", func() {
		var opts *DeleteDiskOpts

		BeforeEach(func() {
			opts = &DeleteDiskOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("DeleteDiskArgs", func() {
		var opts *DeleteDiskArgs

		BeforeEach(func() {
			opts = &DeleteDiskArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("CID", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CID", opts)).To(Equal(
						`positional-arg-name:"CID"`,
					))
				})
			})

		})
	})

	Describe("SnapshotsOpts", func() {
		var opts *SnapshotsOpts

		BeforeEach(func() {
			opts = &SnapshotsOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

		})
	})

	Describe("TakeSnapshotOpts", func() {
		var opts *TakeSnapshotOpts

		BeforeEach(func() {
			opts = &TakeSnapshotOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

		})
	})

	Describe("DeleteSnapshotOpts", func() {
		var opts *DeleteSnapshotOpts

		BeforeEach(func() {
			opts = &DeleteSnapshotOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("DeleteSnapshotArgs", func() {
		var opts *DeleteSnapshotArgs

		BeforeEach(func() {
			opts = &DeleteSnapshotArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("CID", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("CID", opts)).To(Equal(
						`positional-arg-name:"CID"`,
					))
				})
			})

		})
	})

	Describe("InstanceSlugArgs", func() {
		var opts *InstanceSlugArgs

		BeforeEach(func() {
			opts = &InstanceSlugArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Slug", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Slug", opts)).To(Equal(
						`positional-arg-name:"POOL/ID"`,
					))
				})
			})

		})
	})

	Describe("InstancesOpts", func() {
		var opts *InstancesOpts

		BeforeEach(func() {
			opts = &InstancesOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Details", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Details", opts)).To(Equal(
						`long:"details" short:"i" description:"Show details including VM CID, persistent disk CID, etc."`,
					))
				})
			})

			Describe("DNS", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DNS", opts)).To(Equal(
						`long:"dns"               description:"Show DNS A records"`,
					))
				})
			})

			Describe("Vitals", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Vitals", opts)).To(Equal(
						`long:"vitals"            description:"Show vitals"`,
					))
				})
			})

			Describe("Processes", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Processes", opts)).To(Equal(
						`long:"ps"      short:"p" description:"Show processes"`,
					))
				})
			})

			Describe("Failing", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Failing", opts)).To(Equal(
						`long:"failing" short:"f" description:"Only show failing instances"`,
					))
				})
			})

		})
	})

	Describe("VMsOpts", func() {
		var opts *VMsOpts

		BeforeEach(func() {
			opts = &VMsOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Details", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Details", opts)).To(Equal(
						`long:"details" short:"i" description:"Show details including VM CID, persistent disk CID, etc."`,
					))
				})
			})

			Describe("DNS", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("DNS", opts)).To(Equal(
						`long:"dns"               description:"Show DNS A records"`,
					))
				})
			})

			Describe("Vitals", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Vitals", opts)).To(Equal(
						`long:"vitals"            description:"Show vitals"`,
					))
				})
			})

		})
	})

	Describe("CloudCheckOpts", func() {
		var opts *CloudCheckOpts

		BeforeEach(func() {
			opts = &CloudCheckOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Auto", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Auto", opts)).To(Equal(
						`long:"auto"   short:"a" description:"Resolve problems automatically"`,
					))
				})
			})

			Describe("Report", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Report", opts)).To(Equal(
						`long:"report" short:"r" description:"Only generate report; don't attempt to resolve problems"`,
					))
				})
			})

		})
	})

	Describe("UpdateResurrectionOpts", func() {
		var opts *UpdateResurrectionOpts

		BeforeEach(func() {
			opts = &UpdateResurrectionOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

		})
	})

	Describe("UpdateResurrectionArgs", func() {
		var opts *UpdateResurrectionArgs

		BeforeEach(func() {
			opts = &UpdateResurrectionArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Enabled", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Enabled", opts)).To(Equal(
						`positional-arg-name:"on|off"`,
					))
				})
			})

		})
	})

	Describe("LogsOpts", func() {
		var opts *LogsOpts

		BeforeEach(func() {
			opts = &LogsOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Destination directory" default:"."`,
					))
				})
			})

			Describe("Follow", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Follow", opts)).To(Equal(
						`long:"follow" short:"f" description:"Follow logs via SSH"`,
					))
				})
			})

			Describe("Num", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Num", opts)).To(Equal(
						`long:"num"              description:"Last number of lines"`,
					))
				})
			})

			Describe("Quiet", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Quiet", opts)).To(Equal(
						`long:"quiet"  short:"q" description:"Suppresses printing of headers when multiple files are being examined."`,
					))
				})
			})

			Describe("Jobs", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Jobs", opts)).To(Equal(
						`long:"job"   description:"Limit to only specific jobs"`,
					))
				})
			})

			Describe("Filters", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Filters", opts)).To(Equal(
						`long:"only"  description:"Filter logs (comma-separated)"`,
					))
				})
			})

			Describe("Agent", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Agent", opts)).To(Equal(
						`long:"agent" description:"Include only agent logs"`,
					))
				})
			})

		})
	})

	Describe("StartOpts", func() {
		var opts *StartOpts

		BeforeEach(func() {
			opts = &StartOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force" description:"No-op for backwards compatibility"`,
					))
				})
			})

		})
	})

	Describe("StopOpts", func() {
		var opts *StopOpts

		BeforeEach(func() {
			opts = &StopOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("Soft", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Soft", opts)).To(Equal(
						`long:"soft" description:"Stop process only (default)"`,
					))
				})
			})

			Describe("Hard", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Hard", opts)).To(Equal(
						`long:"hard" description:"Delete VM (but keep persistent disk)"`,
					))
				})
			})

			Describe("SkipDrain", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SkipDrain", opts)).To(Equal(
						`long:"skip-drain" description:"Skip running drain scripts"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force"      description:"No-op for backwards compatibility"`,
					))
				})
			})

		})
	})

	Describe("RestartOpts", func() {
		var opts *RestartOpts

		BeforeEach(func() {
			opts = &RestartOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("SkipDrain", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SkipDrain", opts)).To(Equal(
						`long:"skip-drain" description:"Skip running drain scripts"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force"      description:"No-op for backwards compatibility"`,
					))
				})
			})

		})
	})

	Describe("RecreateOpts", func() {
		var opts *RecreateOpts

		BeforeEach(func() {
			opts = &RecreateOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("SkipDrain", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("SkipDrain", opts)).To(Equal(
						`long:"skip-drain" description:"Skip running drain scripts"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force"      description:"No-op for backwards compatibility"`,
					))
				})
			})

		})
	})

	Describe("SSHOpts", func() {
		var opts *SSHOpts

		BeforeEach(func() {
			opts = &SSHOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("Command", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Command", opts)).To(Equal(
						`long:"command" short:"c" description:"Command"`,
					))
				})
			})

			Describe("RawOpts", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("RawOpts", opts)).To(Equal(
						`long:"opts"              description:"Options to pass through to SSH"`,
					))
				})
			})

			Describe("Results", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Results", opts)).To(Equal(
						`long:"results" short:"r" description:"Collect results into a table instead of streaming"`,
					))
				})
			})

		})
	})

	Describe("SCPOpts", func() {
		var opts *SCPOpts

		BeforeEach(func() {
			opts = &SCPOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Recursive", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Recursive", opts)).To(Equal(
						`long:"recursive" short:"r" description:"Recursively copy entire directories. Note that symbolic links encountered are followed in the tree traversal"`,
					))
				})
			})

		})
	})

	Describe("SCPArgs", func() {
		var opts *SCPArgs

		BeforeEach(func() {
			opts = &SCPArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Paths", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Paths", opts)).To(Equal(
						`positional-arg-name:"PATH"`,
					))
				})
			})

		})
	})

	Describe("GatewayFlags", func() {
		var opts *GatewayFlags

		BeforeEach(func() {
			opts = &GatewayFlags{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Disable", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Disable", opts)).To(Equal(
						`long:"gw-disable" description:"Disable usage of gateway connection"`,
					))
				})
			})

			Describe("Username", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Username", opts)).To(Equal(
						`long:"gw-user"        description:"Username for gateway connection"`,
					))
				})
			})

			Describe("Host", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Host", opts)).To(Equal(
						`long:"gw-host"        description:"Host for gateway connection"`,
					))
				})
			})

			Describe("PrivateKeyPath", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("PrivateKeyPath", opts)).To(Equal(
						`long:"gw-private-key" description:"Private key path for gateway connection"`,
					))
				})
			})

		})
	})

	Describe("InitReleaseOpts", func() {
		var opts *InitReleaseOpts

		BeforeEach(func() {
			opts = &InitReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

			Describe("Git", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Git", opts)).To(Equal(
						`long:"git" description:"Initialize git repository"`,
					))
				})
			})

		})
	})

	Describe("ResetReleaseOpts", func() {
		var opts *ResetReleaseOpts

		BeforeEach(func() {
			opts = &ResetReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

		})
	})

	Describe("GenerateJobOpts", func() {
		var opts *GenerateJobOpts

		BeforeEach(func() {
			opts = &GenerateJobOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

		})
	})

	Describe("GenerateJobArgs", func() {
		var opts *GenerateJobArgs

		BeforeEach(func() {
			opts = &GenerateJobArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`positional-arg-name:"NAME"`,
					))
				})
			})

		})
	})

	Describe("GeneratePackageOpts", func() {
		var opts *GeneratePackageOpts

		BeforeEach(func() {
			opts = &GeneratePackageOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

		})
	})

	Describe("GeneratePackageArgs", func() {
		var opts *GeneratePackageArgs

		BeforeEach(func() {
			opts = &GeneratePackageArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`positional-arg-name:"NAME"`,
					))
				})
			})

		})
	})

	Describe("CreateReleaseOpts", func() {
		var opts *CreateReleaseOpts

		BeforeEach(func() {
			opts = &CreateReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`long:"name"               description:"Custom release name"`,
					))
				})
			})

			Describe("Version", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Version", opts)).To(Equal(
						`long:"version"            description:"Custom release version (e.g.: 1.0.0, 1.0-beta.2+dev.10)"`,
					))
				})
			})

			Describe("TimestampVersion", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("TimestampVersion", opts)).To(Equal(
						`long:"timestamp-version"  description:"Create release with the timestamp as the dev version (e.g.: 1+dev.TIMESTAMP)"`,
					))
				})
			})

			Describe("Final", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Final", opts)).To(Equal(
						`long:"final" description:"Make it a final release"`,
					))
				})
			})

			Describe("Tarball", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Tarball", opts)).To(Equal(
						`long:"tarball" description:"Create release tarball"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force" description:"Ignore Git dirty state check"`,
					))
				})
			})

		})
	})

	Describe("CreateReleaseArgs", func() {
		var opts *CreateReleaseArgs

		BeforeEach(func() {
			opts = &CreateReleaseArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Manifest", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Manifest", opts)).To(Equal(
						`positional-arg-name:"PATH"`,
					))
				})
			})

		})
	})

	Describe("FinalizeReleaseOpts", func() {
		var opts *FinalizeReleaseOpts

		BeforeEach(func() {
			opts = &FinalizeReleaseOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

			Describe("Name", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Name", opts)).To(Equal(
						`long:"name"    description:"Custom release name"`,
					))
				})
			})

			Describe("Version", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Version", opts)).To(Equal(
						`long:"version" description:"Custom release version (e.g.: 1.0.0, 1.0-beta.2+dev.10)"`,
					))
				})
			})

			Describe("Force", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Force", opts)).To(Equal(
						`long:"force" description:"Ignore Git dirty state check"`,
					))
				})
			})

		})
	})

	Describe("FinalizeReleaseArgs", func() {
		var opts *FinalizeReleaseArgs

		BeforeEach(func() {
			opts = &FinalizeReleaseArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Path", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Path", opts)).To(Equal(
						`positional-arg-name:"PATH"`,
					))
				})
			})

		})
	})

	Describe("BlobsOpts", func() {
		var opts *BlobsOpts

		BeforeEach(func() {
			opts = &BlobsOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

		})
	})

	Describe("AddBlobArgs", func() {
		var opts *AddBlobArgs

		BeforeEach(func() {
			opts = &AddBlobArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Path", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Path", opts)).To(Equal(
						`positional-arg-name:"PATH"`,
					))
				})
			})

			Describe("BlobsPath", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("BlobsPath", opts)).To(Equal(
						`positional-arg-name:"BLOBS-PATH"`,
					))
				})
			})

		})
	})

	Describe("RemoveBlobOpts", func() {
		var opts *RemoveBlobOpts

		BeforeEach(func() {
			opts = &RemoveBlobOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Args", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Args", opts)).To(Equal(
						`positional-args:"true" required:"true"`,
					))
				})
			})

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

		})
	})

	Describe("RemoveBlobArgs", func() {
		var opts *RemoveBlobArgs

		BeforeEach(func() {
			opts = &RemoveBlobArgs{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("BlobsPath", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("BlobsPath", opts)).To(Equal(
						`positional-arg-name:"BLOBS-PATH"`,
					))
				})
			})

		})
	})

	Describe("SyncBlobsOpts", func() {
		var opts *SyncBlobsOpts

		BeforeEach(func() {
			opts = &SyncBlobsOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

		})
	})

	Describe("UploadBlobsOpts", func() {
		var opts *UploadBlobsOpts

		BeforeEach(func() {
			opts = &UploadBlobsOpts{}
		})

		Describe("Tags (these are used by go flags)", func() {

			Describe("Directory", func() {
				It("contains desired values", func() {
					Expect(getStructTagForName("Directory", opts)).To(Equal(
						`long:"dir" description:"Release directory path if not current working directory" default:"."`,
					))
				})
			})

		})
	})

})
