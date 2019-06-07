package manifest

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"code.cloudfoundry.org/cf-operator/pkg/bosh/bpm"
	"code.cloudfoundry.org/cf-operator/pkg/kube/util/names"
)

// ContainerFactory builds Kubernetes containers from BOSH jobs
type ContainerFactory struct {
	manifestName         string
	igName               string
	releaseImageProvider ReleaseImageProvider
	bpmConfigs           bpm.Configs
}

// NewContainerFactory returns a new ContainerFactory for a BOSH instant group
func NewContainerFactory(manifestName string, igName string, releaseImageProvider ReleaseImageProvider, bpmConfigs bpm.Configs) *ContainerFactory {
	return &ContainerFactory{
		manifestName:         manifestName,
		igName:               igName,
		releaseImageProvider: releaseImageProvider,
		bpmConfigs:           bpmConfigs,
	}
}

// JobsToInitContainers creates a list of Containers for corev1.PodSpec InitContainers field
func (c *ContainerFactory) JobsToInitContainers(jobs []Job, hasPersistentDisk bool) ([]corev1.Container, error) {
	copyingSpecsInitContainers := make([]corev1.Container, 0)
	boshPreStartInitContainers := make([]corev1.Container, 0)
	bpmPreStartInitContainers := make([]corev1.Container, 0)

	copyingSpecsUniq := map[string]struct{}{}
	for _, job := range jobs {
		jobImage, err := c.releaseImageProvider.GetReleaseImage(c.igName, job.Name)
		if err != nil {
			return []corev1.Container{}, err
		}

		// One copying specs init container for each release.
		if _, done := copyingSpecsUniq[job.Release]; !done {
			copyingSpecsUniq[job.Release] = struct{}{}
			copyingSpecsInitContainer := jobSpecCopierContainer(job.Release, jobImage, VolumeRenderingDataName)
			copyingSpecsInitContainers = append(copyingSpecsInitContainers, copyingSpecsInitContainer)
		}

		// Setup the BOSH pre-start init containers.
		boshPreStart := filepath.Join(VolumeJobsDirMountPath, job.Name, "bin", "pre-start")
		boshPreStartInitContainer := corev1.Container{
			Name:         fmt.Sprintf("bosh-pre-start-%s", job.Name),
			Image:        jobImage,
			VolumeMounts: instanceGroupVolumeMounts(),
			Command:      []string{"/bin/sh", "-c"},
			Args:         []string{fmt.Sprintf(`if [ -x "%[1]s" ]; then "%[1]s"; fi`, boshPreStart)},
		}
		if hasPersistentDisk {
			persistentDisk := corev1.VolumeMount{
				Name:      VolumeStoreDirName,
				MountPath: VolumeStoreDirMountPath,
			}
			boshPreStartInitContainer.VolumeMounts = append(boshPreStartInitContainer.VolumeMounts, persistentDisk)
		}
		boshPreStartInitContainers = append(boshPreStartInitContainers, boshPreStartInitContainer)

		// Setup the BPM pre-start init containers.
		bpmConfig, ok := c.bpmConfigs[job.Name]
		if !ok {
			return []corev1.Container{}, errors.Errorf("failed to lookup bpm config for bosh job '%s' in bpm configs", job.Name)
		}
		for _, process := range bpmConfig.Processes {
			if process.Hooks.PreStart != "" {
				bpmPrestartInitContainer := corev1.Container{
					Name:         fmt.Sprintf("bpm-pre-start-%s", process.Name),
					Image:        jobImage,
					VolumeMounts: instanceGroupVolumeMounts(),
					Command:      []string{process.Hooks.PreStart},
				}
				bpmPreStartInitContainers = append(bpmPreStartInitContainers, bpmPrestartInitContainer)
			}
		}
	}

	_, resolvedPropertiesSecretName := names.CalculateEJobOutputSecretPrefixAndName(
		names.DeploymentSecretTypeInstanceGroupResolvedProperties,
		c.manifestName,
		c.igName,
		true,
	)

	initContainers := flattenContainers(
		copyingSpecsInitContainers,
		templateRenderingContainer(c.igName, resolvedPropertiesSecretName),
		createDirContainer(c.igName, jobs),
		boshPreStartInitContainers,
		bpmPreStartInitContainers,
	)

	return initContainers, nil
}

// JobsToContainers creates a list of Containers for corev1.PodSpec Containers field
func (c *ContainerFactory) JobsToContainers(jobs []Job) ([]corev1.Container, error) {
	var containers []corev1.Container

	if len(jobs) == 0 {
		return nil, fmt.Errorf("instance group %s has no jobs defined", c.igName)
	}

	for _, job := range jobs {
		jobImage, err := c.releaseImageProvider.GetReleaseImage(c.igName, job.Name)
		if err != nil {
			return []corev1.Container{}, err
		}

		processes, err := c.generateJobContainers(job, jobImage)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to apply bpm information on bosh job '%s', instance group '%s'", job.Name, c.igName)
		}

		containers = append(containers, processes...)
	}
	return containers, nil
}

func (c *ContainerFactory) generateJobContainers(job Job, jobImage string) ([]corev1.Container, error) {
	containers := []corev1.Container{}
	template := corev1.Container{
		Name:         job.Name,
		Image:        jobImage,
		VolumeMounts: instanceGroupVolumeMounts(),
	}

	bpmConfig, ok := c.bpmConfigs[job.Name]
	if !ok {
		return containers, errors.Errorf("failed to lookup bpm config for bosh job '%s' in bpm configs", job.Name)
	}

	if len(bpmConfig.Processes) < 1 {
		return containers, errors.New("bpm info has no processes")
	}

	for _, process := range bpmConfig.Processes {
		container := template.DeepCopy()

		container.Name = fmt.Sprintf("%s-%s", job.Name, process.Name)
		container.Command = []string{process.Executable}
		container.Args = process.Args
		for name, value := range process.Env {
			container.Env = append(template.Env, corev1.EnvVar{Name: name, Value: value})
		}
		container.WorkingDir = process.Workdir
		container.SecurityContext = &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: ToCapability(process.Capabilities),
			},
		}

		if len(job.Properties.BOSHContainerization.Run.HealthChecks) > 0 {
			for name, hc := range job.Properties.BOSHContainerization.Run.HealthChecks {
				if name == process.Name {
					if hc.ReadinessProbe != nil {
						container.ReadinessProbe = hc.ReadinessProbe
					}
					if hc.LivenessProbe != nil {
						container.LivenessProbe = hc.LivenessProbe
					}
				}
			}
		}

		containers = append(containers, *container)
	}

	return containers, nil
}

// templateJobContainer creates the template for a job container.
func templateJobContainer(name, image string) corev1.Container {
	return corev1.Container{
		Name:         name,
		Image:        image,
		VolumeMounts: instanceGroupVolumeMounts(),
	}
}

// jobSpecCopierContainer will return a corev1.Container{} with the populated field
func jobSpecCopierContainer(releaseName string, jobImage string, volumeMountName string) corev1.Container {
	inContainerReleasePath := filepath.Join(VolumeRenderingDataMountPath, "jobs-src", releaseName)
	return corev1.Container{
		Name:  fmt.Sprintf("spec-copier-%s", releaseName),
		Image: jobImage,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      volumeMountName,
				MountPath: VolumeRenderingDataMountPath,
			},
		},
		Command: []string{
			"bash",
			"-c",
			fmt.Sprintf(`mkdir -p "%s" && cp -ar %s/* "%s"`, inContainerReleasePath, VolumeJobsSrcDirMountPath, inContainerReleasePath),
		},
	}
}

func templateRenderingContainer(name string, secretName string) corev1.Container {
	return corev1.Container{
		Name:  fmt.Sprintf("renderer-%s", name),
		Image: GetOperatorDockerImage(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      VolumeRenderingDataName,
				MountPath: VolumeRenderingDataMountPath,
			},
			{
				Name:      VolumeJobsDirName,
				MountPath: VolumeJobsDirMountPath,
			},
			{
				Name:      generateVolumeName(secretName),
				MountPath: fmt.Sprintf("/var/run/secrets/resolved-properties/%s", name),
				ReadOnly:  true,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "INSTANCE_GROUP_NAME",
				Value: name,
			},
			{
				Name:  "BOSH_MANIFEST_PATH",
				Value: fmt.Sprintf("/var/run/secrets/resolved-properties/%s/properties.yaml", name),
			},
			{
				Name:  "JOBS_DIR",
				Value: VolumeRenderingDataMountPath,
			},
		},
		Args: []string{"util", "template-render"},
	}
}

func createDirContainer(name string, jobs []Job) corev1.Container {
	dirs := []string{}
	for _, job := range jobs {
		jobDirs := append(job.dataDirs(job.Name), job.sysDirs(job.Name)...)
		dirs = append(dirs, jobDirs...)
	}

	return corev1.Container{
		Name:  fmt.Sprintf("create-dirs-%s", name),
		Image: GetOperatorDockerImage(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      VolumeDataDirName,
				MountPath: VolumeDataDirMountPath,
			},
			{
				Name:      VolumeSysDirName,
				MountPath: VolumeSysDirMountPath,
			},
		},
		Env:     []corev1.EnvVar{},
		Command: []string{"/bin/sh"},
		Args:    []string{"-c", "mkdir -p " + strings.Join(dirs, " ")},
	}
}

// ToCapability converts string slice into Capability slice of kubernetes
func ToCapability(s []string) []corev1.Capability {
	capabilities := make([]corev1.Capability, len(s))
	for capabilityIndex, capability := range s {
		capabilities[capabilityIndex] = corev1.Capability(capability)
	}
	return capabilities
}

func instanceGroupVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      VolumeRenderingDataName,
			MountPath: VolumeRenderingDataMountPath,
		},
		{
			Name:      VolumeJobsDirName,
			MountPath: VolumeJobsDirMountPath,
		},
		{
			Name:      VolumeDataDirName,
			MountPath: VolumeDataDirMountPath,
		},
		{
			Name:      VolumeSysDirName,
			MountPath: VolumeSysDirMountPath,
		},
	}
}

// flattenContainers will flatten the containers parameter. Each argument passed to
// flattenContainers should be a corev1.Container or []corev1.Container. The final
// []corev1.Container creation is optimized to prevent slice re-allocation.
func flattenContainers(containers ...interface{}) []corev1.Container {
	var totalLen int
	for _, instance := range containers {
		switch v := instance.(type) {
		case []corev1.Container:
			totalLen += len(v)
		case corev1.Container:
			totalLen++
		}
	}
	result := make([]corev1.Container, 0, totalLen)
	for _, instance := range containers {
		switch v := instance.(type) {
		case []corev1.Container:
			result = append(result, v...)
		case corev1.Container:
			result = append(result, v)
		}
	}
	return result
}
