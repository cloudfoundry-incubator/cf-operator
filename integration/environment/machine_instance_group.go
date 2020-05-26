package environment

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"

	bdv1 "code.cloudfoundry.org/quarks-operator/pkg/kube/apis/boshdeployment/v1alpha1"
	"code.cloudfoundry.org/quarks-operator/pkg/kube/apis/quarksstatefulset/v1alpha1"
)

// WaitForInstanceGroup blocks until all selected pods of the instance group are running. It fails after the timeout.
func (m *Machine) WaitForInstanceGroup(namespace string, deployment string, igName string, version string, count int) error {
	return m.WaitForInstanceGroupVersions(namespace, deployment, igName, count, version)
}

// WaitForInstanceGroupVersions blocks until the specified number of pods from
// the instance group are running.  It counts running pods from all given
// versions. It fails after the timeout.
func (m *Machine) WaitForInstanceGroupVersions(namespace string, deployment string, igName string, count int, versions ...string) error {
	labels := labels.Set(map[string]string{
		bdv1.LabelDeploymentName:    deployment,
		bdv1.LabelInstanceGroupName: igName,
	}).String()
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		n, err := m.PodCount(namespace, labels, func(pod corev1.Pod) bool {
			return pod.Status.Phase == corev1.PodRunning && contains(versions, pod.Annotations[v1alpha1.AnnotationVersion])
		})
		if err != nil {
			return false, err
		}
		return n == count, nil
	})
}

// GetInstanceGroupPods returns all pods from a specific instance group version
func (m *Machine) GetInstanceGroupPods(namespace string, deployment string, igName string) (*corev1.PodList, error) {
	labels := labels.Set(map[string]string{
		bdv1.LabelDeploymentName:    deployment,
		bdv1.LabelInstanceGroupName: igName,
	}).String()
	return m.GetPods(namespace, labels)
}

func contains(versions []string, version string) bool {
	for _, a := range versions {
		if a == version {
			return true
		}
	}
	return false
}
