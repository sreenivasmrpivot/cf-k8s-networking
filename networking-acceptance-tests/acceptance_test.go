package acceptance_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"code.cloudfoundry.org/cf-k8s-networking/acceptance/cfg"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"
	"github.com/cloudfoundry-incubator/cf-test-helpers/workflowhelpers"
	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	kubectl   *KubeCtl
	TestSetup *workflowhelpers.ReproducibleTestSuiteSetup
	globals   *Globals
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)

	config, err := cfg.NewConfig(
		os.Getenv("CONFIG"),
		os.Getenv("KUBECONFIG"),
		os.Getenv("CONFIG_KEEP_CLUSTER") != "",
		os.Getenv("CONFIG_KEEP_CF") != "",
	)

	if err != nil {
		defer GinkgoRecover()
		fmt.Println("Failed to load config.")
		t.Fail()
	}

	kubectl = &KubeCtl{kubeConfigPath: config.KubeConfigPath}

	var _ = SynchronizedBeforeSuite(func() []byte {
		_, err := kubectl.Run("cluster-info")
		if err != nil {
			panic(err)
		}

		TestSetup = workflowhelpers.NewTestSuiteSetup(config)
		TestSetup.Setup()
		workflowhelpers.AsUser(TestSetup.AdminUserContext(), time.Minute, func() {
			Eventually(cf.Cf("enable-feature-flag", "diego_docker")).Should(gexec.Exit(0))
		})

		g := &Globals{}
		g.SysComponentSelector = createSystemComponent()
		g.AppGuid = pushApp(generator.PrefixedRandomName("ACCEPTANCE", "app"))

		data, err := g.Serialize()
		if err != nil {
			panic(err)
		}

		return data
	}, func(data []byte) {
		globals = &Globals{}
		err := globals.Deserialize(data)
		if err != nil {
			panic(err)
		}

		SetDefaultEventuallyTimeout(1 * time.Minute)
		SetDefaultEventuallyPollingInterval(1 * time.Second)
	})

	SynchronizedAfterSuite(func() {}, func() {
		if TestSetup != nil && !config.KeepCFChanges {
			TestSetup.Teardown()
		}

		if !config.KeepClusterChanges {
			destroySystemComponent()
		}
	})

	RunSpecs(t, "Acceptance Suite")
}

type Globals struct {
	AppGuid              string `json:"app_guid"`
	SysComponentSelector string `json:"sys_component_selector"`
}

func (g *Globals) Serialize() ([]byte, error) {
	return json.Marshal(g)
}

func (g *Globals) Deserialize(data []byte) error {
	return json.Unmarshal(data, g)
}

type Config struct {
	AppsDomain string `json:"apps_domain"`
}

type KubeCtl struct {
	kubeConfigPath string
}

func (kc *KubeCtl) Run(args ...string) ([]byte, error) {
	cmd := exec.Command("kubectl", args...)
	cmd.Env = []string{
		fmt.Sprintf("KUBECONFIG=%s", kc.kubeConfigPath),
		fmt.Sprintf("PATH=%s", os.Getenv("PATH")),
		fmt.Sprintf("HOME=%s", filepath.Dir(kc.kubeConfigPath)), // fixme because kubectl will create another .kube folder inside our provided kubeConfgPath
	}
	fmt.Fprintf(GinkgoWriter, "\n+ kubectl '%s'\n", strings.Join(args, "' '"))
	output, err := cmd.CombinedOutput()
	GinkgoWriter.Write(output)
	return output, err
}

func createSystemComponent() string {
	cmd := exec.Command("kapp", "deploy", "-n", systemNamespace, "-a", "system-component", "-y", "-f", "./assets/system-component.yml")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	session.Wait(2 * time.Minute)

	return "app=test-system-component"
}

func destroySystemComponent() {
	cmd := exec.Command("kapp", "delete", "-n", systemNamespace, "-a", "system-component", "-y")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	session.Wait(1 * time.Minute)
}

func pushDockerApp(name string, container string, flags ...string) string {
	args := append(
		[]string{"push",
			name,
			"-o", container,
			"-u", "http",
		},
		flags...,
	)

	session := cf.Cf(args...)

	// cf push does not exit 0 on cf-for-k8s yet because logcache is unreliable (stats server error)
	Expect(session.Wait(120 * time.Second)).To(gexec.Exit())

	guid := strings.TrimSpace(string(cf.Cf("app", name, "--guid").Wait().Out.Contents()))

	return guid
}

func pushProxy(name string, flags ...string) string {
	return pushDockerApp(name, "cfrouting/proxy", flags...)
}

func pushApp(name string) string {
	return pushDockerApp(name, "cfrouting/httpbin8080")
}

func mapRoute(appName, domain, hostname string) {
	session := cf.Cf("map-route",
		appName,
		domain,
		"--hostname", hostname,
	)
	Expect(session.Wait(120 * time.Second)).To(gexec.Exit())
}
