package integrations_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"gp_upgrade/hub/cluster"
	"gp_upgrade/hub/configutils"
	"gp_upgrade/hub/services"
	"gp_upgrade/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
	"google.golang.org/grpc"
)

// needs the cli and the hub
var _ = Describe("check config", func() {
	var (
		dir            string
		hub            *services.HubClient
		commandExecer  *testutils.FakeCommandExecer
		cliToHubPort   int
		hubToAgentPort int
	)

	BeforeEach(func() {
		cliToHubPort = 7527
		hubToAgentPort = 6416

		var err error
		dir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())

		conf := &services.HubConfig{
			CliToHubPort:   cliToHubPort,
			HubToAgentPort: hubToAgentPort,
			StateDir:       dir,
		}
		reader := configutils.NewReader()

		commandExecer = &testutils.FakeCommandExecer{}
		commandExecer.SetOutput(&testutils.FakeCommand{})

		hub = services.NewHub(&cluster.Pair{}, &reader, grpc.DialContext, commandExecer.Exec, conf)

		Expect(checkPortIsAvailable(cliToHubPort)).To(BeTrue())
		go hub.Start()
	})

	AfterEach(func() {
		hub.Stop()
		Expect(checkPortIsAvailable(cliToHubPort)).To(BeTrue())
		os.RemoveAll(dir)
	})

	Describe("when a greenplum master db on localhost is up and running", func() {
		It("happy: the database configuration is saved to a specified location", func() {
			session := runCommand("check", "config", "--master-host", "localhost")
			if session.ExitCode() != 0 {
				fmt.Println("make sure greenplum is running")
			}
			Expect(session).To(Exit(0))

			_, err := ioutil.ReadFile(configutils.GetConfigFilePath(dir))
			testutils.Check("cannot read file", err)

			reader := configutils.Reader{}
			reader.OfOldClusterConfig(dir)
			err = reader.Read()
			testutils.Check("cannot read config", err)

			Expect(len(reader.GetSegmentConfiguration())).To(BeNumerically(">", 1))
		})
	})

	It("fails if the --master-host flag is missing", func() {
		checkConfigSession := runCommand("check", "config")
		Expect(checkConfigSession).Should(Exit(1))
		Expect(string(checkConfigSession.Out.Contents())).To(Equal("Required flag(s) \"master-host\" have/has not been set\n"))
	})
})
