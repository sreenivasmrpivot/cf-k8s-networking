package acceptance_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/cloudfoundry-incubator/cf-test-helpers/generator"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routte Mappings", func() {
	var (
		app1name      string
		app2name      string
		routeHostname string
		// app2guid string
		domain string
	)

	BeforeEach(func() {
		app1name = generator.PrefixedRandomName("ACCEPTANCE", "cats")
		app2name = generator.PrefixedRandomName("ACCEPTANCE", "dogs")

		_ = pushProxy(app1name, "--no-start")
		// app2guid = pushProxy(app2name)
		_ = pushProxy(app2name, "--no-start")

		configFile, err := ioutil.ReadFile(os.Getenv("CONFIG"))
		if err != nil {
			panic(fmt.Errorf("error reading config %v", err))
		}
		config := &Config{}
		err = json.Unmarshal([]byte(configFile), config)

		if err != nil {
			panic(fmt.Errorf("error parsing json %v", err))
		}
		domain = config.AppsDomain

		routeHostname = generator.PrefixedRandomName("ACCEPTANCE", "shared-route")

		mapRoute(app1name, domain, routeHostname)
		mapRoute(app2name, domain, routeHostname)
	})

	AfterEach(func() {
		cf.Cf("delete", app1name)
		cf.Cf("delete", app2name)
	})

	When("there are two apps and one route", func() {
		When("only one app is started", func() {
			BeforeEach(func() {
				cf.Cf("start", app1name)
				time.Sleep(20 * time.Second)
			})

			It("routes to the started app", func() {
				route := fmt.Sprintf("http://%s.%s/", routeHostname, domain)

				appIPs := make(map[string]struct{})

				for i := 0; i < 10; i++ {
					for _, ip := range getProxyListeningAddresses(route) {
						appIPs[ip] = struct{}{}
					}
				}

				fmt.Printf("%v\n", appIPs)
				Expect(len(appIPs)).To(Equal(2))
			})
		})

		When("both app are started", func() {
			BeforeEach(func() {
				cf.Cf("start", app1name)
				cf.Cf("start", app2name)
				time.Sleep(20 * time.Second)
			})

			It("routes to both", func() {
				route := fmt.Sprintf("http://%s.%s/", routeHostname, domain)

				appIPs := make(map[string]struct{})

				for i := 0; i < 10; i++ {
					for _, ip := range getProxyListeningAddresses(route) {
						appIPs[ip] = struct{}{}
					}
				}

				fmt.Printf("%v\n", appIPs)
				Expect(len(appIPs)).To(Equal(3))
			})
		})

	})
})

func getProxyListeningAddresses(route string) []string {
	fmt.Printf("Attempting to reach %s", route)
	resp, err := http.Get(route)
	Expect(err).NotTo(HaveOccurred())
	Expect(resp.StatusCode).To(Equal(200))
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	Expect(err).NotTo(HaveOccurred())

	var resJSON struct {
		ListenAddresses []string `json:"ListenAddresses"`
	}

	err = json.Unmarshal(body, &resJSON)
	Expect(err).NotTo(HaveOccurred())

	return resJSON.ListenAddresses
}
