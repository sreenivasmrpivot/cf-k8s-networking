package stress_test

import (
	"fmt"
	"strings"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Stress Tests", func() {
	BeforeEach(func() {
		By("Creating Many Routes")

		routeTmpl, err := template.ParseFiles("fixtures/route_template.yml")
		Expect(err).NotTo(HaveOccurred())

		type Route struct {
			Name            string
			Host            string
			Path            string
			Domain          string
			DestinationGUID string
			AppGUID         string
		}

		var routesBuilder strings.Builder

		for i := 0; i < 1000; i++ {
			route := Route{
				Name:            fmt.Sprintf("route-%d", i),
				Host:            fmt.Sprintf("hostname-%d", i),
				Path:            fmt.Sprintf("/%d", i),
				Domain:          "apps.example.com",
				DestinationGUID: fmt.Sprintf("destination-guid-%d", i),
				AppGUID:         fmt.Sprintf("app-guid-%d", i),
			}

			// Create a new YAML document for each Route definition
			_, err := routesBuilder.WriteString("---\n")
			Expect(err).NotTo(HaveOccurred())

			// Evaluate the Route template and write the resulting Route definition to routesBuilder
			err = routeTmpl.Execute(&routesBuilder, route)
			Expect(err).NotTo(HaveOccurred())
		}

		routesReader := strings.NewReader(routesBuilder.String())

		session, err := kubectl.RunWithStdin(routesReader, "apply", "-f", "-")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session, 5*time.Minute).Should(gexec.Exit(0))

		// TODO check that 1000 routes are now on the cluster
	})

	Measure("routecontroller stress", func(b Benchmarker) {
	}, 1)
})
