package cli_test

import (
	"encoding/json"
	"io/ioutil"
	"os/exec"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/yaml"

	"code.cloudfoundry.org/cf-operator/pkg/bosh/manifest"
)

var _ = Describe("bpm-configs", func() {
	var (
		manifestPath string
	)

	act := func(manifestPath string) (session *gexec.Session, err error) {
		args := []string{"util", "bpm-configs", "-m", manifestPath, "-b", assetPath, "-g", "log-api", "--output-file-path", assetPath + "/output.json"}
		cmd := exec.Command(cliPath, args...)
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		return
	}

	Context("when manifest exists", func() {
		BeforeEach(func() {
			manifestPath = assetPath + "/gatherManifest.yml"
		})

		It("prints the bpm configs to a file", func() {
			session, err := act(manifestPath)
			Expect(err).ToNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))

			var jsonOutput map[string]string
			dataBytes, err := ioutil.ReadFile(filepath.Join(assetPath, "output.json"))
			Expect(err).ToNot(HaveOccurred())

			err = json.Unmarshal(dataBytes, &jsonOutput)
			Expect(err).ToNot(HaveOccurred())

			bpmInfo := manifest.BPMInfo{}
			err = yaml.Unmarshal([]byte(jsonOutput["bpm.yaml"]), &bpmInfo)
			Expect(err).ToNot(HaveOccurred())

			config := bpmInfo.Configs["loggregator_trafficcontroller"]
			Expect(len(config.Processes)).To(Equal(1))
			Expect(config.Processes[0].Executable).To(Equal("/var/vcap/packages/loggregator_trafficcontroller/trafficcontroller"))
		})
	})
})
