package storage_test

import (
	"fmt"
	"testing"

	"k8s.io/client-go/rest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pkg/errors"

	"code.cloudfoundry.org/quarks-operator/integration/environment"
	cmdHelper "code.cloudfoundry.org/quarks-utils/testing"
	utils "code.cloudfoundry.org/quarks-utils/testing/integration"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Storage Suite")
}

var (
	env              *environment.Environment
	namespacesToNuke []string
	kubeConfig       *rest.Config
	qjobCmd          environment.QuarksJobCmd
)

var _ = SynchronizedBeforeSuite(func() []byte {
	var err error
	kubeConfig, err = utils.KubeConfig()
	if err != nil {
		fmt.Printf("WARNING: failed to get kube config: %v\n", err)
	}

	// Ginkgo node 1 gets to setup the CRDs
	err = environment.ApplyCRDs(kubeConfig)
	if err != nil {
		fmt.Printf("WARNING: failed to apply CRDs: %v", err)
	}

	// Ginkgo node 1 needs to build the quarks job binary
	qjobCmd = environment.NewQuarksJobCmd()
	err = qjobCmd.Build()
	if err != nil {
		fmt.Printf("WARNING: failed to build quarks-job: %v\n", err)
	}

	return []byte(qjobCmd.Path)
}, func(data []byte) {
	var err error
	kubeConfig, err = utils.KubeConfig()
	if err != nil {
		fmt.Printf("WARNING: failed to get kube config: %v", err)
	}
	qjobCmd = environment.NewQuarksJobCmd()
	qjobCmd.Path = string(data)
})

var _ = BeforeEach(func() {
	env = environment.NewEnvironment(kubeConfig)
	err := env.SetupClientsets()
	if err != nil {
		errors.Wrapf(err, "Integration setup failed. Creating clientsets in %s", env.Namespace)
	}
	err = env.SetupNamespace()
	if err != nil {
		fmt.Printf("WARNING: failed to setup namespace %s: %v\n", env.Namespace, err)
	}
	namespacesToNuke = append(namespacesToNuke, env.Namespace)

	err = env.SetupQjobAccount()
	if err != nil {
		fmt.Printf("WARNING: failed to setup quarks-job operator service account: %s\n", err)
	}

	err = qjobCmd.Start(env.Namespace)
	if err != nil {
		fmt.Printf("WARNING: failed to start operator dependencies: %v\n", err)
	}

	err = env.StartOperator()
	if err != nil {
		fmt.Printf("WARNING: failed to start operator: %v\n", err)
	}
})

var _ = AfterEach(func() {
	env.Teardown(CurrentGinkgoTestDescription().Failed)
	gexec.KillAndWait()
})

var _ = AfterSuite(func() {
	// Nuking all namespaces at the end of the run
	for _, namespace := range namespacesToNuke {
		err := cmdHelper.DeleteNamespace(namespace)
		if err != nil && !env.NamespaceDeletionInProgress(err) {
			fmt.Printf("WARNING: failed to delete namespace %s: %v\n", namespace, err)
		}
		err = cmdHelper.DeleteWebhooks(namespace)
		if err != nil {
			fmt.Printf("WARNING: failed to delete mutatingwebhookconfiguration in %s: %v\n", namespace, err)
		}
	}
})
