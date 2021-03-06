package stress_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"sigs.k8s.io/kind/pkg/cluster"
)

func TestStress(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Route Controller Stress Tests Suite")
}

var (
	kubectl     kubectlRunner
	ytt         yttRunner
	resultsPath string
)

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Minute)

	kubectl = CreateKindCluster()

	// Deploy Route CRD
	session, err := kubectl.Run("apply", "-f", "../../config/crd/networking.cloudfoundry.org_routes.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	// Deploy Istio's Virtual Service CRD
	session, err = kubectl.Run("apply", "-f", "../integration/fixtures/istio-virtual-service.yaml")
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))

	var found bool
	resultsPath, found = os.LookupEnv("RESULTS_PATH")
	if !found {
		resultsPath = "results.json"
	}

})

var _ = AfterSuite(func() {
	kubectl.DestroyCluster()
})

type kubectlRunner struct {
	kubeconfigFilePath string
	clusterName        string
}

func CreateKindCluster() kubectlRunner {
	name := fmt.Sprintf("stress-tests-%d", rand.Uint64())
	provider := cluster.NewProvider()
	err := provider.Create(name)
	// retry once
	if err != nil {
		time.Sleep(5 * time.Second)
		err = provider.Create(name)
	}

	Expect(err).NotTo(HaveOccurred())

	kubeConfig, err := provider.KubeConfig(name, false)
	Expect(err).NotTo(HaveOccurred())

	kubeConfigFile, err := ioutil.TempFile("", fmt.Sprintf("kubeconfig-%s", name))
	Expect(err).NotTo(HaveOccurred())
	defer kubeConfigFile.Close()

	_, err = kubeConfigFile.Write([]byte(kubeConfig))
	Expect(err).NotTo(HaveOccurred())

	return kubectlRunner{clusterName: name, kubeconfigFilePath: kubeConfigFile.Name()}
}

func (k kubectlRunner) DestroyCluster() {
	provider := cluster.NewProvider()
	err := provider.Delete(k.clusterName, k.kubeconfigFilePath)
	Expect(err).NotTo(HaveOccurred())
}

func (k kubectlRunner) Run(kubectlCommandArgs ...string) (*gexec.Session, error) {
	cmd := k.generateCommand(nil, kubectlCommandArgs...)
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	return gexec.Start(cmd, foo, foo)
}

func (k kubectlRunner) RunWithStdin(stdin io.Reader, kubectlCommandArgs ...string) (*gexec.Session, error) {
	cmd := k.generateCommand(stdin, kubectlCommandArgs...)
	var b bytes.Buffer
	foo := bufio.NewWriter(&b)
	return gexec.Start(cmd, foo, foo)
}

func (k kubectlRunner) generateCommand(stdin io.Reader, kubectlCommandArgs ...string) *exec.Cmd {
	// fmt.Fprintf(GinkgoWriter, "+ kubectl %s\n", strings.Join(kubectlCommandArgs, " "))
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
	cmd := exec.Command("ytt", yttCommandArgs...)
	return gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
}

type TestRouteTemplate struct {
	Name            string
	Host            string
	Path            string
	Domain          string
	DestinationGUID string
	AppGUID         string
	Tag             string
}

func buildRoutes(numberOfRoutes int, tag string) io.Reader {
	routeTmpl, err := template.ParseFiles("fixtures/route_template.yml")
	Expect(err).NotTo(HaveOccurred())

	var routesBuilder strings.Builder

	for i := 0; i < numberOfRoutes; i++ {
		route := buildRoute(i, tag)
		// Create a new YAML document for each Route definition
		_, err := routesBuilder.WriteString("---\n")
		Expect(err).NotTo(HaveOccurred())

		// Evaluate the Route template and write the resulting Route definition to routesBuilder
		err = routeTmpl.Execute(&routesBuilder, route)
		Expect(err).NotTo(HaveOccurred())
	}

	return strings.NewReader(routesBuilder.String())
}

func buildSingleRoute(index int, tag string) io.Reader {
	routeTmpl, err := template.ParseFiles("fixtures/route_template.yml")
	Expect(err).NotTo(HaveOccurred())

	var routesBuilder strings.Builder

	route := buildRoute(index, tag)
	err = routeTmpl.Execute(&routesBuilder, route)
	Expect(err).NotTo(HaveOccurred())

	return strings.NewReader(routesBuilder.String())
}

func buildRoute(index int, tag string) TestRouteTemplate {
	return TestRouteTemplate{
		Name:            fmt.Sprintf("route-%s-%d", tag, index),
		Host:            fmt.Sprintf("hostname-%s-%d", tag, index),
		Path:            fmt.Sprintf("/%s-%d", tag, index),
		Domain:          "apps.example.com",
		DestinationGUID: fmt.Sprintf("destination-guid-%s-%d", tag, index),
		AppGUID:         fmt.Sprintf("app-guid-%s-%d", tag, index),
		Tag:             tag,
	}
}
