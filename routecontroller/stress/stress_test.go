package stress_test

import (
	"bytes"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Stress Tests", func() {
	var (
		numberOfRoutes = 400
	)

	BeforeEach(func() {
		routes := buildRoutes(numberOfRoutes)
		session, err := kubectl.RunWithStdin(routes, "apply", "-f", "-")
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
	}, 1)
})
