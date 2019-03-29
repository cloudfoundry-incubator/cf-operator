package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	btg "github.com/viovanov/bosh-template-go"
	yaml "gopkg.in/yaml.v2"

	"code.cloudfoundry.org/cf-operator/pkg/bosh/manifest"
)

// templateRenderCmd represents the template-render command
var templateRenderCmd = &cobra.Command{
	Use:   "template-render [flags]",
	Short: "Renders a bosh manifest",
	Long: `Renders a bosh manifest.

This will render a provided manifest instance-group
`,
	RunE: func(cmd *cobra.Command, args []string) error {

		deploymentManifest := viper.GetString("bosh_manifest")
		jobsDir := viper.GetString("jobs_dir")
		instanceGroupName := viper.GetString("instance_group_name")

		index := viper.GetInt("index")
		if index < 0 {
			// calculate index following the formula specified in
			// docs/rendering_templates.md
			azIndex := viper.GetInt("azindex")
			if azIndex < 0 {
				return fmt.Errorf("required parameter 'azindex' not set")
			}
			replicas := viper.GetInt("replicas")
			if replicas < 0 {
				return fmt.Errorf("required parameter 'replicas' not set")
			}
			podOrdinal := viper.GetInt("podordinal")
			if podOrdinal < 0 {
				// Infer ordinal from hostname
				hostname, err := os.Hostname()
				r, _ := regexp.Compile(`-(\d+)(-|\.|\z)`)
				if err != nil {
					return errors.Wrap(err, "getting the hostname")
				}
				match := r.FindStringSubmatch(hostname)
				if len(match) < 2 {
					return fmt.Errorf("can not extract the pod ordinal from hostname '%s'", hostname)
				}
				podOrdinal, _ = strconv.Atoi(match[1])
			}

			index = (azIndex-1)*replicas + podOrdinal
		}

		return RenderData(deploymentManifest, jobsDir, "/var/vcap/jobs/", instanceGroupName, index)
	},
}

func init() {
	rootCmd.AddCommand(templateRenderCmd)

	templateRenderCmd.Flags().StringP("manifest", "m", "", "path to the manifest file")
	templateRenderCmd.Flags().StringP("jobs-dir", "j", "", "path to the jobs dir.")
	templateRenderCmd.Flags().StringP("instance-group-name", "g", "", "the instance-group name to render")
	templateRenderCmd.Flags().IntP("index", "", -1, "index of the instance spec")
	templateRenderCmd.Flags().IntP("azindex", "", -1, "az index")
	templateRenderCmd.Flags().IntP("podordinal", "", -1, "pod ordinal")
	templateRenderCmd.Flags().IntP("replicas", "", -1, "number of replicas")

	viper.BindPFlag("bosh_manifest", templateRenderCmd.Flags().Lookup("manifest"))
	viper.BindPFlag("jobs_dir", templateRenderCmd.Flags().Lookup("jobs-dir"))
	viper.BindPFlag("instance_group_name", templateRenderCmd.Flags().Lookup("instance-group-name"))
	viper.BindPFlag("azindex", templateRenderCmd.Flags().Lookup("azindex"))
	viper.BindPFlag("index", templateRenderCmd.Flags().Lookup("index"))
	viper.BindPFlag("podordinal", templateRenderCmd.Flags().Lookup("podordinal"))
	viper.BindPFlag("replicas", templateRenderCmd.Flags().Lookup("replicas"))

	viper.AutomaticEnv()
	viper.BindEnv("bosh_manifest", "MANIFEST_PATH")
	viper.BindEnv("jobs_dir", "JOBS_DIR")
	viper.BindEnv("instance_group_name", "INSTANCE_GROUP_NAME")
	viper.BindEnv("index", "SPEC_INDEX")
	viper.BindEnv("azindex", "CF_OPERATOR_AZ_INDEX")
	viper.BindEnv("podordinal", "POD_ORDINAL")
	viper.BindEnv("replicas", "REPLICAS")

}

// RenderData will render manifest instance group
func RenderData(manifestPath string, jobsDir, jobsOutputDir string, instanceGroupName string, index int) error {

	// Loading deployment manifest file
	resolvedYML, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return errors.Wrapf(err, "couldn't read manifest file %s", manifestPath)
	}
	deploymentManifest := manifest.Manifest{}
	err = yaml.Unmarshal(resolvedYML, &deploymentManifest)
	if err != nil {
		return errors.Wrapf(err, "failed to unmarshal deployment manifest %s", manifestPath)
	}

	// Loop over instancegroups
	for _, instanceGroup := range deploymentManifest.InstanceGroups {

		// Filter based on the instance group name
		if instanceGroup.Name != instanceGroupName {
			continue
		}

		// Render all files for all jobs included in this instance_group.
		for _, job := range instanceGroup.Jobs {
			jobInstanceLinks := []manifest.Link{}

			// Find job instance that's being rendered
			var currentJobInstance *manifest.JobInstance
			for _, instance := range job.Properties.BOSHContainerization.Instances {
				if instance.Index == index {
					currentJobInstance = &instance
					break
				}
			}
			if currentJobInstance == nil {
				return fmt.Errorf("no instance found for index '%d'", index)
			}

			// Loop over name and link
			for name, jobConsumersLink := range job.Properties.BOSHContainerization.Consumes {
				jobInstances := []manifest.JobInstance{}

				// Loop over instances of link
				for _, jobConsumerLinkInstance := range jobConsumersLink.Instances {
					jobInstances = append(jobInstances, manifest.JobInstance{
						Address: jobConsumerLinkInstance.Address,
						AZ:      jobConsumerLinkInstance.AZ,
						ID:      jobConsumerLinkInstance.ID,
						Index:   jobConsumerLinkInstance.Index,
						Name:    jobConsumerLinkInstance.Name,
					})
				}

				jobInstanceLinks = append(jobInstanceLinks, manifest.Link{
					Name:       name,
					Instances:  jobInstances,
					Properties: jobConsumersLink.Properties,
				})
			}

			jobSrcDir := filepath.Join(jobsDir, "jobs-src", job.Release, job.Name)
			jobMFFile := filepath.Join(jobSrcDir, "job.MF")
			jobMfBytes, err := ioutil.ReadFile(jobMFFile)
			if err != nil {
				return errors.Wrapf(err, "failed to read job spec file %s", jobMFFile)
			}

			jobSpec := manifest.JobSpec{}
			if err := yaml.Unmarshal([]byte(jobMfBytes), &jobSpec); err != nil {
				return errors.Wrapf(err, "failed to unmarshal job spec %s", jobMFFile)
			}

			// Loop over templates for rendering files
			for source, destination := range jobSpec.Templates {
				absDest := filepath.Join(jobsOutputDir, job.Name, destination)
				os.MkdirAll(filepath.Dir(absDest), 0755)

				properties := job.Properties.ToMap()

				renderPointer := btg.NewERBRenderer(
					&btg.EvaluationContext{
						Properties: properties,
					},

					&btg.InstanceInfo{
						Address: currentJobInstance.Address,
						AZ:      currentJobInstance.AZ,
						ID:      currentJobInstance.ID,
						Index:   string(currentJobInstance.Index),
						Name:    currentJobInstance.Name,
					},

					jobMFFile,
				)

				// Create the destination file
				absDestFile, err := os.Create(absDest)
				if err != nil {
					return err
				}
				defer absDestFile.Close()
				if err = renderPointer.Render(filepath.Join(jobSrcDir, "templates", source), absDestFile.Name()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
