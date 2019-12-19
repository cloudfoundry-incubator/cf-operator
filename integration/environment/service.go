package environment

import (
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	bdm "code.cloudfoundry.org/cf-operator/pkg/bosh/manifest"
)

// WaitForServiceVersion blocks until the service of the instance group is created/updated. It fails after the timeout.
func (m *Machine) WaitForServiceVersion(namespace string, serviceName string, version string) error {
	return wait.PollImmediate(m.PollInterval, m.PollTimeout, func() (bool, error) {
		svc, err := m.Clientset.CoreV1().Services(namespace).Get(serviceName, metav1.GetOptions{})
		if err != nil {
			return false, errors.Wrapf(err, "failed to get service '%s'", svc)
		}

		deploymentVersion, ok := svc.Labels[bdm.LabelDeploymentVersion]
		if ok && deploymentVersion == version {
			return true, nil
		}

		return false, nil
	})
}
