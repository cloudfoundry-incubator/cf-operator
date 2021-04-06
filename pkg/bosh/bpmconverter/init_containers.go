package bpmconverter

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	qjv1a1 "code.cloudfoundry.org/quarks-job/pkg/kube/apis/quarksjob/v1alpha1"
	"code.cloudfoundry.org/quarks-operator/pkg/bosh/bpm"
	bdm "code.cloudfoundry.org/quarks-operator/pkg/bosh/manifest"
	"code.cloudfoundry.org/quarks-operator/pkg/kube/util/operatorimage"
	"code.cloudfoundry.org/quarks-utils/pkg/names"
)

// JobsToInitContainers creates a list of Containers for corev1.PodSpec InitContainers field.
func (c *ContainerFactoryImpl) JobsToInitContainers(
	jobs []bdm.Job,
	defaultVolumeMounts []corev1.VolumeMount,
	bpmDisks bdm.Disks,
	requiredService *string,
) ([]corev1.Container, error) {
	copyingSpecsInitContainers := make([]corev1.Container, 0)
	boshPreStartInitContainers := make([]corev1.Container, 0)
	bpmPreStartInitContainers := make([]corev1.Container, 0)

	copyingSpecsUniq := map[string]struct{}{}
	for _, job := range jobs {
		jobImage, err := c.releaseImageProvider.GetReleaseImage(c.instanceGroupName, job.Name)
		if err != nil {
			return []corev1.Container{}, err
		}

		// One copying specs init container for each release.
		if _, done := copyingSpecsUniq[job.Release]; !done {
			copyingSpecsUniq[job.Release] = struct{}{}
			copyingSpecsInitContainer := JobSpecCopierContainer(job.Release, jobImage, VolumeRenderingDataName)
			copyingSpecsInitContainers = append(copyingSpecsInitContainers, copyingSpecsInitContainer)
		}

		// Setup the BPM pre-start init containers before the BOSH pre-start init container in order to
		// collect all the extra BPM volumes and pass them to the BOSH pre-start init container.
		bpmConfig, ok := c.bpmConfigs[job.Name]
		if !ok {
			return []corev1.Container{}, errors.Errorf("failed to lookup bpm config for bosh job '%s' in bpm configs", job.Name)
		}

		jobDisks := bpmDisks.Filter("job_name", job.Name)
		var ephemeralMount *corev1.VolumeMount
		ephemeralDisks := jobDisks.Filter("ephemeral", "true")
		if len(ephemeralDisks) > 0 {
			ephemeralMount = ephemeralDisks[0].VolumeMount
		}
		var persistentDiskMount *corev1.VolumeMount
		persistentDiskDisks := jobDisks.Filter("persistent", "true")
		if len(persistentDiskDisks) > 0 {
			persistentDiskMount = persistentDiskDisks[0].VolumeMount
		}

		for _, process := range bpmConfig.Processes {
			if process.Hooks.PreStart != "" {
				processDisks := jobDisks.Filter("process_name", process.Name)
				bpmVolumeMounts := make([]corev1.VolumeMount, 0)
				for _, processDisk := range processDisks {
					bpmVolumeMounts = append(bpmVolumeMounts, *processDisk.VolumeMount)
				}
				processVolumeMounts := append(defaultVolumeMounts, bpmVolumeMounts...)
				if ephemeralMount != nil {
					processVolumeMounts = append(processVolumeMounts, *ephemeralMount)
				}
				if persistentDiskMount != nil {
					processVolumeMounts = append(processVolumeMounts, *persistentDiskMount)
				}
				container := bpmPreStartInitContainer(
					process,
					jobImage,
					processVolumeMounts,
					bpmConfig.Debug,
					bpmConfig.Run.SecurityContext.DeepCopy(),
				)

				bpmPreStartInitContainers = append(bpmPreStartInitContainers, *container.DeepCopy())
			}
		}

		// Setup the BOSH pre-start init container for the job.
		boshPreStartInitContainer := boshPreStartInitContainer(
			job.Name,
			jobImage,
			append(defaultVolumeMounts, bpmDisks.VolumeMounts()...),
			bpmConfig.Debug,
			bpmConfig.Run.SecurityContext.DeepCopy(),
		)
		boshPreStartInitContainers = append(boshPreStartInitContainers, *boshPreStartInitContainer.DeepCopy())
	}

	initContainers := flattenContainers(
		containerRunCopier(),
		copyingSpecsInitContainers,
		templateRenderingContainer(c.instanceGroupName, c.version == "1"),
		createDirContainer(jobs, c.instanceGroupName),
		createWaitContainer(requiredService),
		boshPreStartInitContainers,
		bpmPreStartInitContainers,
	)

	return initContainers, nil
}

func createWaitContainer(requiredService *string) []corev1.Container {
	if requiredService == nil {
		return nil
	}
	return []corev1.Container{{
		Name:    fmt.Sprintf("wait-for-%s", *requiredService),
		Image:   operatorimage.GetOperatorDockerImage(),
		Command: []string{"/usr/bin/dumb-init", "--"},
		Args: []string{
			"/bin/sh",
			"-xc",
			fmt.Sprintf("time quarks-operator util wait %s", *requiredService),
		},
	}}

}

func containerRunCopier() corev1.Container {
	dstDir := fmt.Sprintf("%s/container-run", VolumeRenderingDataMountPath)
	return corev1.Container{
		Name:            "container-run-copier",
		Image:           operatorimage.GetOperatorDockerImage(),
		ImagePullPolicy: operatorimage.GetOperatorImagePullPolicy(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      VolumeRenderingDataName,
				MountPath: VolumeRenderingDataMountPath,
			},
		},
		Command: entrypoint,
		Args: []string{
			"/bin/sh",
			"-c",
			fmt.Sprintf(`
				set -o errexit
				mkdir -p '%[1]s'
				time cp /usr/local/bin/container-run '%[1]s'/container-run
			`, dstDir),
		},
	}
}

func templateRenderingContainer(instanceGroupName string, initialRollout bool) corev1.Container {
	container := corev1.Container{
		Name:            "template-render",
		Image:           operatorimage.GetOperatorDockerImage(),
		ImagePullPolicy: operatorimage.GetOperatorImagePullPolicy(),
		VolumeMounts: []corev1.VolumeMount{
			*renderingVolumeMount(),
			*jobsDirVolumeMount(),
			resolvedPropertiesVolumeMount(instanceGroupName),
		},
		Env: []corev1.EnvVar{
			{
				Name:  EnvInstanceGroupName,
				Value: instanceGroupName,
			},
			{
				Name:  qjv1a1.RemoteIDKey,
				Value: instanceGroupName,
			},
			{
				Name:  EnvBOSHManifestPath,
				Value: fmt.Sprintf(resolvedPropertiesFormat+"/properties.yaml", instanceGroupName),
			},
			{
				Name:  EnvJobsDir,
				Value: VolumeRenderingDataMountPath,
			},
			{
				Name: PodIPEnvVar,
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
			podOrdinalEnv,
			replicasEnv,
			azIndexEnv,
		},
		Command: entrypoint,
		Args: []string{
			"/bin/sh",
			"-xc",
			"time quarks-operator util template-render",
		},
	}

	// Default is true for template-render
	if !initialRollout {
		container.Env = append(container.Env,
			corev1.EnvVar{
				Name:  EnvInitialRollout,
				Value: "false",
			})
	}

	return container
}

func createDirContainer(jobs []bdm.Job, instanceGroupName string) corev1.Container {
	dirs := []string{}
	for _, job := range jobs {
		jobDirs := append(job.DataDirs(), job.SysDirs()...)
		dirs = append(dirs, jobDirs...)
	}

	return corev1.Container{
		Name:            "create-dirs",
		Image:           operatorimage.GetOperatorDockerImage(),
		ImagePullPolicy: operatorimage.GetOperatorImagePullPolicy(),
		VolumeMounts: []corev1.VolumeMount{
			{
				Name: volumeDataDirName(
					instanceGroupName),
				MountPath: VolumeDataDirMountPath,
			},
			*sysDirVolumeMount(),
		},
		Command: entrypoint,
		Args: []string{
			"/bin/sh",
			"-xc",
			fmt.Sprintf("time mkdir -p %s", strings.Join(dirs, " ")),
		},
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &vcapUserID,
		},
	}
}

func boshPreStartInitContainer(
	jobName string,
	jobImage string,
	volumeMounts []corev1.VolumeMount,
	debug bool,
	securityContext *corev1.SecurityContext,
) corev1.Container {
	boshPreStart := filepath.Join(VolumeJobsDirMountPath, jobName, "bin", "pre-start")

	var script string
	if debug {
		script = fmt.Sprintf(`if [ -x "%[1]s" ]; then "%[1]s" || ( echo "Debug window 1hr" ; sleep 3600 ); fi`, boshPreStart)
	} else {
		script = fmt.Sprintf(`if [ -x "%[1]s" ]; then time "%[1]s" ; fi`, boshPreStart)
	}

	if securityContext == nil {
		securityContext = &corev1.SecurityContext{}
	}
	securityContext.RunAsUser = &rootUserID

	return corev1.Container{
		Name:         names.Sanitize(fmt.Sprintf("bosh-pre-start-%s", jobName)),
		Image:        jobImage,
		VolumeMounts: deduplicateVolumeMounts(volumeMounts),
		Command:      entrypoint,
		Args: []string{
			"/bin/sh",
			"-xc",
			script,
		},
		Env: []corev1.EnvVar{
			podOrdinalEnv,
			replicasEnv,
			azIndexEnv,
		},
		SecurityContext: securityContext,
	}
}

func bpmPreStartInitContainer(
	process bpm.Process,
	jobImage string,
	volumeMounts []corev1.VolumeMount,
	debug bool,
	securityContext *corev1.SecurityContext,
) corev1.Container {
	var script string
	if debug {
		script = fmt.Sprintf(`%s || ( echo "Debug window 1hr" ; sleep 3600 )`, process.Hooks.PreStart)
	} else {
		script = "time " + process.Hooks.PreStart
	}

	if securityContext == nil {
		securityContext = &corev1.SecurityContext{}
	}
	if securityContext.Capabilities == nil && len(process.Capabilities) > 0 {
		securityContext.Capabilities = &corev1.Capabilities{
			Add: capability(process.Capabilities),
		}
	}
	if securityContext.Privileged == nil {
		securityContext.Privileged = &process.Unsafe.Privileged
	}
	securityContext.RunAsUser = &rootUserID

	return corev1.Container{
		Name:         names.Sanitize(fmt.Sprintf("bpm-pre-start-%s", process.Name)),
		Image:        jobImage,
		VolumeMounts: deduplicateVolumeMounts(volumeMounts),
		Command:      entrypoint,
		Args: []string{
			"/bin/sh",
			"-xc",
			script,
		},
		Env: []corev1.EnvVar{
			podOrdinalEnv,
			replicasEnv,
			azIndexEnv,
		},
		SecurityContext: securityContext,
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
