// Copyright DataStax, Inc.
// Please see the included license file for details.

package tolerations

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	ginkgo_util "github.com/datastax/cass-operator/mage/ginkgo"
	"github.com/datastax/cass-operator/mage/kubectl"
)

var (
	testName    = "Tolerations"
	opNamespace = "test-tolerations"
	dc1Name     = "dc1"
	dc1Yaml     = "../testdata/tolerations-dc.yaml"
	dc1Resource = fmt.Sprintf("CassandraDatacenter/%s", dc1Name)
	ns          = ginkgo_util.NewWrapper(testName, opNamespace)
)

func TestLifecycle(t *testing.T) {
	AfterSuite(func() {
		logPath := fmt.Sprintf("%s/aftersuite", ns.LogDir)
		err := kubectl.DumpAllLogs(logPath).ExecV()
		if err != nil {
			fmt.Printf("\n\tError during dumping logs: %s\n\n", err.Error())
		}
		fmt.Printf("\n\tPost-run logs dumped at: %s\n\n", logPath)
		ns.Terminate()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, testName)
}

var _ = Describe(testName, func() {
	Context("when in a new cluster", func() {
		Specify("the operator can build pods with tolerations", func() {

			By("creating a namespace for the cass-operator")
			err := kubectl.CreateNamespace(opNamespace).ExecV()
			Expect(err).ToNot(HaveOccurred())

			// For now, let's taint 5 of the 6 nodes, and put the operator on the first
			for i := 2; i <= 6; i++ {
				node := fmt.Sprintf("kind-worker%d", i)
				step := fmt.Sprintf("tainting %s", node)
				k := kubectl.Taint(
					node,
					"test",
					"testvalue",
					"NoSchedule")
				ns.ExecAndLog(step, k)
			}

			step := "setting up cass-operator resources via helm chart"
			ns.HelmInstall("../../charts/cass-operator-chart")

			ns.WaitForOperatorReady()

			step = "creating first datacenter resource"
			k := kubectl.ApplyFiles(dc1Yaml)
			ns.ExecAndLog(step, k)

			ns.WaitForDatacenterReady(dc1Name)
		})
	})
})
