package stress_test

import (
	"io"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	session, err := kubectl.Run("apply", "-f", "../../config/crd/networking.cloudfoundry.org_routes.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 1*time.Minute).Should(gexec.Exit(0))

	// Deploy Istio's Virtual Service CRD
	session, err = kubectl.Run("apply", "-f", "../integration/fixtures/istio-virtual-service.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, 1*time.Minute).Should(gexec.Exit(0))
})

type kubectlRunner struct {
	kubeconfigFilePath string
}

func (k kubectlRunner) Run(kubectlCommandArgs ...string) (*gexec.Session, error) {
	cmd := k.generateCommand(nil, kubectlCommandArgs...)
	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}

func (k kubectlRunner) RunWithStdin(stdin io.Reader, kubectlCommandArgs ...string) (*gexec.Session, error) {
	cmd := k.generateCommand(stdin, kubectlCommandArgs...)
	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}

func (k kubectlRunner) generateCommand(stdin io.Reader, kubectlCommandArgs ...string) *exec.Cmd {
	cmd := exec.Command("kubectl", kubectlCommandArgs...)
	cmd.Env = append(cmd.Env, "KUBECONFIG="+k.kubeconfigFilePath)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	return cmd
}
