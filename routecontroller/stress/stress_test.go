package stress_test

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Stress Tests", func() {
	var (
		numberOfRoutes = 100
	)

	BeforeEach(func() {
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

		for i := 0; i < numberOfRoutes; i++ {
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
		Eventually(session).Should(gexec.Exit(0))

		Expect(kubectl.GetNumberOf("routes")).To(Equal(numberOfRoutes))
	})

	AfterEach(func() {
		// // Note: Eventually this should be captured as a measurement
		// session, err := kubectl.Run("delete", "routes", "--all")
		// Expect(err).NotTo(HaveOccurred())
		// Eventually(session).Should(gexec.Exit(0))

		session, err := kubectl.Run("delete", "deployment", "routecontroller")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		session, err = kubectl.Run("delete", "virtualservices", "--all")
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
	})

	Measure("routecontroller stress", func(b Benchmarker) {
		// Make sure we're starting from a blank slate
		Expect(kubectl.GetNumberOf("virtualservices")).To(Equal(0))

		yttSession, err := ytt.Run(
			"-f", filepath.Join("..", "..", "config", "routecontroller"),
			"-f", filepath.Join("..", "..", "config", "values.yaml"),
			"-v", "systemNamespace=default",
		)
		Expect(err).NotTo(HaveOccurred())
		Eventually(yttSession).Should(gexec.Exit(0))
		// TODO: why do we need to get Contents() ?
		yttContents := yttSession.Out.Contents()
		testReader := bytes.NewReader(yttContents)

		b.Time("Deploying the routecontroller", func() {
			session, err := kubectl.RunWithStdin(testReader, "apply", "-f", "-")
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			session, err = kubectl.Run("rollout", "status", "deployment", "routecontroller")
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Should(gbytes.Say("successfully rolled out"))
			Eventually(session).Should(gexec.Exit(0))
		})

		b.Time("Processing 1000 new routes at once", func() {
			// Heisenberg's VirtualServices? Does running 'get' interfere with routecontroller's processing?
			Eventually(func() int { return kubectl.GetNumberOf("virtualservices") }, 30*time.Minute, 500*time.Millisecond).Should(Equal(numberOfRoutes))
		})
	}, 2)
})
