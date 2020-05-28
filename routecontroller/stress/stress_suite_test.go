package stress_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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

var (
	kubectl kubectlRunner
	ytt     yttRunner
)

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Minute)

	kubeconfigFromEnv := os.Getenv("KUBECONFIG")
	if kubeconfigFromEnv == "" {
		// TODO: Default to $HOME/.kube/config ?
		Fail("Required environment variable KUBECONFIG was not set.")
	}

	kubectl = kubectlRunner{kubeconfigFilePath: kubeconfigFromEnv}

	// Deploy Route CRD
	session, err := kubectl.Run("apply", "-f", "../../config/crd/networking.cloudfoundry.org_routes.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	// Deploy Istio's Virtual Service CRD
	session, err = kubectl.Run("apply", "-f", "../integration/fixtures/istio-virtual-service.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
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
	fmt.Fprintf(GinkgoWriter, "+ kubectl %s\n", strings.Join(kubectlCommandArgs, " "))
	cmd := exec.Command("kubectl", kubectlCommandArgs...)
	cmd.Env = append(cmd.Env, "KUBECONFIG="+k.kubeconfigFilePath)
	if stdin != nil {
		cmd.Stdin = stdin
	}

	return cmd
}

func (k kubectlRunner) GetNumberOf(resourceName string) int {
	session, err := k.Run("get", resourceName, "--no-headers")
	if err != nil {
		return 0
	}

	session.Wait(5 * time.Minute)

	if session.ExitCode() != 0 {
		return 0
	}

	return strings.Count(string(session.Out.Contents()), "\n")
}

type yttRunner struct {
}

func (y yttRunner) Run(yttCommandArgs ...string) (*gexec.Session, error) {
	fmt.Fprintf(GinkgoWriter, "+ ytt %s\n", strings.Join(yttCommandArgs, " "))
	cmd := exec.Command("ytt", yttCommandArgs...)
	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}
