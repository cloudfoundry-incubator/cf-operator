package bpmconverter

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"code.cloudfoundry.org/quarks-operator/pkg/bosh/bpm"
	bdm "code.cloudfoundry.org/quarks-operator/pkg/bosh/manifest"
	bdv1 "code.cloudfoundry.org/quarks-operator/pkg/kube/apis/boshdeployment/v1alpha1"
	boshnames "code.cloudfoundry.org/quarks-operator/pkg/kube/util/names"
	"code.cloudfoundry.org/quarks-utils/pkg/names"
)

const (
	// VolumeRenderingDataName is the volume name for the rendering data.
	VolumeRenderingDataName = "rendering-data"
	// VolumeRenderingDataMountPath is the mount path for the rendering data.
	VolumeRenderingDataMountPath = "/var/vcap/all-releases"

	// VolumeJobsDirName is the volume name for the jobs directory.
	VolumeJobsDirName = "jobs-dir"
	// VolumeJobsDirMountPath is the mount path for the jobs directory.
	VolumeJobsDirMountPath = "/var/vcap/jobs"

	// VolumeJobsSrcDirName is the volume name for the jobs-src directory.
	VolumeJobsSrcDirName = "jobs-src-dir"
	// VolumeJobsSrcDirMountPath is the mount path for the jobs-src directory.
	VolumeJobsSrcDirMountPath = "/var/vcap/jobs-src"

	// VolumeDataDirMountPath is the mount path for the ephemeral (data) directory.
	VolumeDataDirMountPath = bdm.DataDir

	// VolumeSysDirName is the volume name for the sys directory.
	VolumeSysDirName = "sys-dir"
	// VolumeSysDirMountPath is the mount path for the sys directory.
	VolumeSysDirMountPath = bdm.SysDir

	// VolumeDrainStampsName is the volume name for the drain-stamps directory.
	VolumeDrainStampsName = "drain-stamps"
	// VolumeDrainStampsMountPath is the mount path for the drain-stamps directory.
	VolumeDrainStampsMountPath = "/mnt/drain-stamps"

	// VolumeStoreDirMountPath is the mount path for the store directory.
	VolumeStoreDirMountPath = "/var/vcap/store"

	// AdditionalVolumeBaseName helps in building an additional volume name together with
	// the index under the additional_volumes bpm list inside the bpm process schema.
	AdditionalVolumeBaseName = "bpm-additional-volume"

	// AdditionalVolumesRegex ensures only a valid path is defined
	// under the additional_volumes bpm list inside the bpm process schema.
	AdditionalVolumesRegex = "((/var/vcap/data/.+)|(/var/vcap/store/.+)|(/var/vcap/sys/run/.+))"

	// AdditionalVolumesVcapStoreRegex ensures that the path is of the form
	// /var/vcap/store.
	AdditionalVolumesVcapStoreRegex = "(/var/vcap/store/.+)"

	// UnrestrictedVolumeBaseName is the volume name for the unrestricted ones.
	UnrestrictedVolumeBaseName = "bpm-unrestricted-volume"

	// resolvedPropertiesFormat describes where to mount the BOSH manifest
	resolvedPropertiesFormat = "/var/run/secrets/resolved-properties/%s"

	// defaultEphemeralVolumeSize is the default size for a PVC used for an ephemeral disk (10GB).
	// This value is used if the ephemeral disk size and the persistent disk sizes are not set.
	defaultEphemeralVolumeSize = 10240
)

// VolumeFactoryImpl is a concrete implementation of VolumeFactory
type VolumeFactoryImpl struct {
}

// NewVolumeFactory returns a concrete implementation of VolumeFactory
func NewVolumeFactory() *VolumeFactoryImpl {
	return &VolumeFactoryImpl{}
}

// GenerateDefaultDisks defines default disks. This looks for:
// - the rendering data volume
// - the the jobs volume
// - the ephemeral (data) volume
// - the sys volume
// - the "not interpolated" manifest volume
// - resolved properties data volume
// - shared empty dir for drain-stamps files
func (f *VolumeFactoryImpl) GenerateDefaultDisks(instanceGroup *bdm.InstanceGroup, igResolvedSecretVersion string, namespace string) bdm.Disks {
	resolvedPropertiesSecretName := boshnames.InstanceGroupSecretName(
		instanceGroup.Name,
		igResolvedSecretVersion,
	)

	var pvc *corev1.PersistentVolumeClaim
	ephemeralAsPVC := instanceGroup.Env.AgentEnvBoshConfig.Agent.Settings.EphemeralAsPVC
	if ephemeralAsPVC {
		pvc = ephemeralPVC(instanceGroup, namespace)
	}

	return bdm.Disks{
		{
			Volume:      renderingVolume(),
			VolumeMount: renderingVolumeMount(),
		},
		{
			Volume:      jobsDirVolume(),
			VolumeMount: jobsDirVolumeMount(),
		},
		{
			// For ephemeral job data.
			// https://bosh.io/docs/vm-config/#jobs-and-packages
			Volume: &corev1.Volume{
				Name:         volumeDataDirName(instanceGroup.Name),
				VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
			},
			VolumeMount: &corev1.VolumeMount{
				Name:      volumeDataDirName(instanceGroup.Name),
				MountPath: VolumeDataDirMountPath,
			},
			PersistentVolumeClaim: pvc,
		},
		{
			Volume:      sysDirVolume(),
			VolumeMount: sysDirVolumeMount(),
		},
		{
			Volume: resolvedPropertiesVolume(resolvedPropertiesSecretName),
		},
		{
			Volume: &corev1.Volume{
				Name: VolumeDrainStampsName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{Medium: corev1.StorageMediumMemory},
				},
			},
			VolumeMount: &corev1.VolumeMount{
				Name:      VolumeDrainStampsName,
				MountPath: VolumeDrainStampsMountPath,
			},
		},
	}
}

// GenerateBPMDisks defines any other volumes required to be mounted,
// based on the bpm process schema definition. This looks for:
// - ephemeral_disk (boolean)
// - persistent_disk (boolean)
// - additional_volumes (list of volumes)
// - unrestricted_volumes (list of volumes)
func (f *VolumeFactoryImpl) GenerateBPMDisks(instanceGroup *bdm.InstanceGroup, bpmConfigs bpm.Configs, namespace string) (bdm.Disks, error) {
	bpmDisks := make(bdm.Disks, 0)

	rAdditionalVolumes := regexp.MustCompile(AdditionalVolumesRegex)

	for _, job := range instanceGroup.Jobs {
		bpmConfig := bpmConfigs[job.Name]
		hasEphemeralDisk := false
		hasPersistentDisk := false
		for _, process := range bpmConfig.Processes {
			processEphemeralDisk := false
			if process.EphemeralDisk != nil {
				processEphemeralDisk = *process.EphemeralDisk
			}

			processPersistentDisk := false
			if process.PersistentDisk != nil {
				processPersistentDisk = *process.PersistentDisk
			}

			if !hasEphemeralDisk && processEphemeralDisk {
				hasEphemeralDisk = true
			}
			if !hasPersistentDisk && processPersistentDisk {
				hasPersistentDisk = true
			}

			// Because we use subpaths for data, store or sys mounts, we want to
			// treat unrestricted volumes as additional volumes.
			filteredUnrestrictedVolumes := []bpm.Volume{}
			for _, unrestrictedVolume := range process.Unsafe.UnrestrictedVolumes {
				// /var/vcap/jobs is already mounted everywhere for quarks, so we ignore anything in there.
				if strings.HasPrefix(unrestrictedVolume.Path, VolumeJobsDirMountPath) {
					continue
				}

				// To consider an unrestricted volume to be an additional volume, the prefix must match, but
				// the volume shouldn't be equal to the known
				// - /var/vcap/data
				// - /var/vcap/store
				// - /var/vcap
				if (strings.HasPrefix(unrestrictedVolume.Path, VolumeDataDirMountPath) && unrestrictedVolume.Path != VolumeDataDirMountPath) ||
					(strings.HasPrefix(unrestrictedVolume.Path, VolumeStoreDirMountPath) && unrestrictedVolume.Path != VolumeStoreDirMountPath) ||
					(strings.HasPrefix(unrestrictedVolume.Path, VolumeSysDirMountPath) && unrestrictedVolume.Path != VolumeSysDirMountPath) {
					process.AdditionalVolumes = append(process.AdditionalVolumes, unrestrictedVolume)
					continue
				}

				filteredUnrestrictedVolumes = append(filteredUnrestrictedVolumes, unrestrictedVolume)
			}
			process.Unsafe.UnrestrictedVolumes = filteredUnrestrictedVolumes

			for _, additionalVolume := range process.AdditionalVolumes {
				match := rAdditionalVolumes.MatchString(additionalVolume.Path)
				if !match {
					return nil, errors.Errorf("the '%s' path, must be a path inside"+
						" '/var/vcap/data', '/var/vcap/store' or '/var/vcap/sys/run', for a path outside these,"+
						" you must use the unrestricted_volumes key", additionalVolume.Path)
				}

				var (
					err        error
					volumeName string
					subPath    string
				)
				if strings.HasPrefix(additionalVolume.Path, VolumeDataDirMountPath) {
					volumeName = volumeDataDirName(
						instanceGroup.Name)
					subPath, err = filepath.Rel(VolumeDataDirMountPath, additionalVolume.Path)
				}
				if strings.HasPrefix(additionalVolume.Path, VolumeStoreDirMountPath) {
					volumeName = generatePersistentVolumeClaimName(instanceGroup.Name)
					subPath, err = filepath.Rel(VolumeStoreDirMountPath, additionalVolume.Path)
				}
				if strings.HasPrefix(additionalVolume.Path, VolumeSysDirMountPath) {
					volumeName = VolumeSysDirName
					subPath, err = filepath.Rel(VolumeSysDirMountPath, additionalVolume.Path)
				}

				if err != nil {
					return nil, errors.Wrapf(err, "failed to calculate subpath for additional volume mount '%s'", additionalVolume.Path)
				}

				additionalDisk := bdm.Disk{
					VolumeMount: &corev1.VolumeMount{
						Name:      volumeName,
						ReadOnly:  !additionalVolume.Writable,
						MountPath: additionalVolume.Path,
						SubPath:   subPath,
					},
					Filters: map[string]string{
						"job_name":     job.Name,
						"process_name": process.Name,
					},
				}
				bpmDisks = append(bpmDisks, additionalDisk)
			}

			for i, unrestrictedVolume := range process.Unsafe.UnrestrictedVolumes {
				volumeName := names.Sanitize(fmt.Sprintf("%s-%s-%s-%b", UnrestrictedVolumeBaseName, job.Name, process.Name, i))
				unrestrictedDisk := bdm.Disk{
					Volume: &corev1.Volume{
						Name:         volumeName,
						VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
					},
					VolumeMount: &corev1.VolumeMount{
						Name:      volumeName,
						ReadOnly:  !unrestrictedVolume.Writable,
						MountPath: unrestrictedVolume.Path,
					},
					Filters: map[string]string{
						"job_name":     job.Name,
						"process_name": process.Name,
					},
				}
				bpmDisks = append(bpmDisks, unrestrictedDisk)
			}
		}

		if hasEphemeralDisk {

			disk := ephemeralDisk(
				namespace,
				instanceGroup,
				job,
			)

			bpmDisks = append(bpmDisks, disk)
		}

		if hasPersistentDisk {
			if instanceGroup.PersistentDisk == nil || *instanceGroup.PersistentDisk <= 0 {
				return bpmDisks, errors.Errorf("job '%s' wants to use persistent disk"+
					" but instance group '%s' doesn't have any persistent disk declaration", job.Name, instanceGroup.Name)
			}

			disk := persistentDisk(
				namespace,
				instanceGroup,
				job,
			)

			bpmDisks = append(bpmDisks, disk)
		}
	}

	return bpmDisks, nil
}

func persistentDisk(namespace string, instanceGroup *bdm.InstanceGroup, job bdm.Job) bdm.Disk {
	persistentVolumeClaim := generatePersistentVolumeClaim(instanceGroup, namespace)

	// Specify the job sub-path inside of the instance group PV
	bpmPersistentDisk := bdm.Disk{
		PersistentVolumeClaim: &persistentVolumeClaim,
		Volume: &corev1.Volume{
			Name: persistentVolumeClaim.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: persistentVolumeClaim.Name,
				},
			},
		},
		VolumeMount: &corev1.VolumeMount{
			Name:      persistentVolumeClaim.Name,
			MountPath: path.Join(VolumeStoreDirMountPath, job.Name),
			SubPath:   job.Name,
		},
		Filters: map[string]string{
			"job_name":   job.Name,
			"persistent": "true",
		},
	}

	return bpmPersistentDisk
}

func ephemeralDisk(namespace string, instanceGroup *bdm.InstanceGroup, job bdm.Job) bdm.Disk {
	persistent := instanceGroup.Env.AgentEnvBoshConfig.Agent.Settings.EphemeralAsPVC

	ephemeralDisk := bdm.Disk{
		VolumeMount: &corev1.VolumeMount{
			Name:      volumeDataDirName(instanceGroup.Name),
			MountPath: path.Join(VolumeDataDirMountPath, job.Name),
			SubPath:   job.Name,
		},
		Filters: map[string]string{
			"job_name":  job.Name,
			"ephemeral": "true",
		},
	}

	if persistent {
		persistentVolumeClaim := ephemeralPVC(instanceGroup, namespace)

		ephemeralDisk = bdm.Disk{
			PersistentVolumeClaim: persistentVolumeClaim,
			Volume: &corev1.Volume{
				Name: persistentVolumeClaim.Name,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: persistentVolumeClaim.Name,
					},
				},
			},
			VolumeMount: &corev1.VolumeMount{
				Name:      persistentVolumeClaim.Name,
				MountPath: path.Join(VolumeDataDirMountPath, job.Name),
				SubPath:   job.Name,
			},
			Filters: map[string]string{
				"job_name":  job.Name,
				"ephemeral": "true",
			},
		}
	}

	return ephemeralDisk
}

func ephemeralPVCSize(instanceGroup *bdm.InstanceGroup) int {
	diskSize := defaultEphemeralVolumeSize

	if instanceGroup.PersistentDisk != nil {
		diskSize = *instanceGroup.PersistentDisk
	}

	if instanceGroup.VMResources != nil && instanceGroup.VMResources.EphemeralDiskSize > 0 {
		diskSize = instanceGroup.VMResources.EphemeralDiskSize
	}

	return diskSize
}

// ephemeralPVC returns a PVC to be used for ephemeral disks
func ephemeralPVC(instanceGroup *bdm.InstanceGroup, namespace string) *corev1.PersistentVolumeClaim {
	diskSize := ephemeralPVCSize(instanceGroup)

	if instanceGroup.PersistentDisk != nil {
		diskSize = *instanceGroup.PersistentDisk
	}

	if instanceGroup.VMResources != nil && instanceGroup.VMResources.EphemeralDiskSize > 0 {
		diskSize = instanceGroup.VMResources.EphemeralDiskSize
	}

	persistentVolumeClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ephemeralPVCName(instanceGroup.Name),
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(fmt.Sprintf("%d%s", diskSize, "Mi")),
				},
			},
		},
	}

	// add storage class if specified
	if instanceGroup.PersistentDiskType != "" {
		persistentVolumeClaim.Spec.StorageClassName = &instanceGroup.PersistentDiskType
	}

	return &persistentVolumeClaim
}

func generatePersistentVolumeClaim(instanceGroup *bdm.InstanceGroup, namespace string) corev1.PersistentVolumeClaim {
	// Spec of a persistentVolumeClaim
	persistentVolumeClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generatePersistentVolumeClaimName(instanceGroup.Name),
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse(fmt.Sprintf("%d%s", *instanceGroup.PersistentDisk, "Mi")),
				},
			},
		},
	}

	// add storage class if specified
	if instanceGroup.PersistentDiskType != "" {
		persistentVolumeClaim.Spec.StorageClassName = &instanceGroup.PersistentDiskType
	}

	return persistentVolumeClaim
}

func ephemeralPVCName(instanceGroupName string) string {
	return names.Sanitize(fmt.Sprintf("%s-%s", instanceGroupName, "ephemeral"))
}

func generatePersistentVolumeClaimName(instanceGroupName string) string {
	return names.Sanitize(fmt.Sprintf("%s-%s", instanceGroupName, "pvc"))
}

func renderingVolume() *corev1.Volume {
	return &corev1.Volume{
		Name:         VolumeRenderingDataName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
}

func renderingVolumeMount() *corev1.VolumeMount {
	return &corev1.VolumeMount{
		Name:      VolumeRenderingDataName,
		MountPath: VolumeRenderingDataMountPath,
	}
}

func jobsDirVolume() *corev1.Volume {
	return &corev1.Volume{
		Name:         VolumeJobsDirName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
}

func jobsDirVolumeMount() *corev1.VolumeMount {
	return &corev1.VolumeMount{
		Name:      VolumeJobsDirName,
		MountPath: VolumeJobsDirMountPath,
	}
}

func resolvedPropertiesVolume(name string) *corev1.Volume {
	return &corev1.Volume{
		Name: bdv1.DeploymentSecretTypeInstanceGroupResolvedProperties.String(),
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: name,
			},
		},
	}
}

func resolvedPropertiesVolumeMount(instanceGroupName string) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "ig-resolved",
		MountPath: fmt.Sprintf(resolvedPropertiesFormat, instanceGroupName),
		ReadOnly:  true,
	}
}

func sysDirVolume() *corev1.Volume {
	return &corev1.Volume{
		Name:         VolumeSysDirName,
		VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
	}
}

// volumeDataDirName is the volume name for the data directory.
func volumeDataDirName(instanceGroupName string) string {
	return ephemeralPVCName(instanceGroupName)
}

func sysDirVolumeMount() *corev1.VolumeMount {
	return &corev1.VolumeMount{
		Name:      VolumeSysDirName,
		MountPath: VolumeSysDirMountPath,
	}
}

func deduplicateVolumeMounts(volumeMounts []corev1.VolumeMount) []corev1.VolumeMount {
	result := []corev1.VolumeMount{}
	uniqueMounts := map[string]struct{}{}

	for _, volumeMount := range volumeMounts {
		if _, ok := uniqueMounts[volumeMount.MountPath]; ok {
			continue
		}

		uniqueMounts[volumeMount.MountPath] = struct{}{}
		result = append(result, volumeMount)
	}

	return result
}
