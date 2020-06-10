package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"

	bm "code.cloudfoundry.org/quarks-operator/testing/boshmanifest"
	"code.cloudfoundry.org/quarks-utils/testing/machine"
)

var _ = Describe("QuarksLink from native to BOSH", func() {
	const (
		manifestRef    = "manifest"
		deploymentName = "test"
	)

	var (
		tearDowns    []machine.TearDownFunc
		boshManifest corev1.Secret
	)

	AfterEach(func() {
		Expect(env.TearDownAll(tearDowns)).To(Succeed())
	})

	JustBeforeEach(func() {
		tearDown, err := env.CreateSecret(env.Namespace, boshManifest)
		Expect(err).NotTo(HaveOccurred())
		tearDowns = append(tearDowns, tearDown)

		_, tearDown, err = env.CreateBOSHDeployment(env.Namespace,
			env.SecretBOSHDeployment(deploymentName, manifestRef))
		Expect(err).NotTo(HaveOccurred())
		tearDowns = append(tearDowns, tearDown)
	})

	Context("when deployment has explicit, external link dependencies", func() {
		BeforeEach(func() {
			natsConfigMap := env.NatsConfigMap(deploymentName)
			tearDown, err := env.CreateConfigMap(env.Namespace, natsConfigMap)
			Expect(err).NotTo(HaveOccurred())
			tearDowns = append(tearDowns, tearDown)

			natsSecret := env.NatsSecret(deploymentName)
			tearDown, err = env.CreateSecret(env.Namespace, natsSecret)
			Expect(err).NotTo(HaveOccurred())
			tearDowns = append(tearDowns, tearDown)

			natsPod := env.NatsPod(deploymentName)
			tearDown, err = env.CreatePod(env.Namespace, natsPod)
			Expect(err).NotTo(HaveOccurred())
			tearDowns = append(tearDowns, tearDown)

			natsService := env.NatsService(deploymentName)
			tearDown, err = env.CreateService(env.Namespace, natsService)
			Expect(err).NotTo(HaveOccurred())
			tearDowns = append(tearDowns, tearDown)

			boshManifest = env.BOSHManifestSecret(manifestRef, bm.NatsSmokeTestWithExternalLinks)
		})

		It("successfully deploys", func() {
			By("waiting for job rendering done", func() {
				err := env.WaitForPods(env.Namespace, "quarks.cloudfoundry.org/instance-group-name=nats-smoke-tests")
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
