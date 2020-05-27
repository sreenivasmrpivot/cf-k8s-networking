package stress_test

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

func TestStress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Route Controller Stress Tests Suite")
}

var kubectl kubectlRunner

var _ = BeforeSuite(func() {
	kubeconfigFromEnv := os.Getenv("KUBECONFIG")
	if kubeconfigFromEnv == "" {
		// TODO: Default to $HOME/.kube/config ?
		Fail("Required environment variable KUBECONFIG was not set.")
	}

	kubectl = kubectlRunner{kubeconfigFilePath: kubeconfigFromEnv}

	// Deploy Route CRD
	Expect(kubectl.Run("apply", "-f", "../../config/crd/networking.cloudfoundry.org_routes.yaml")).
		To(Say("success")) // TODO figure out real success message

	// Deploy Istio's Virtual Service CRD
	Expect(kubectl.Run("apply", "-f", "../integration/fixtures/istio-virtual-service.yaml")).
		To(Say("success"))
})

type kubectlRunner struct {
	kubeconfigFilePath string
}

func (k kubectlRunner) Run(kubectlCommandArgs ...string) (*gexec.Session, error) {
	cmd := exec.Command("kubectl", kubectlCommandArgs...)
	cmd.Env = append(cmd.Env, "KUBECONFIG="+k.kubeconfigFilePath)

	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}
