package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	runtimedebug "runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/bmc-toolbox/bmclib-tester/internal"
	"github.com/go-yaml/yaml"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	// file containing tests to run
	testsFile string

	// file containing hardware to run tests on
	hardwareFile string

	// duration to allow tests to run
	timeout time.Duration

	logLevel string
)

// runCmd represents the tester command to test bmclib features
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run tests defined in the configuration files (--tests) on the hardware defined (--hardware)",
	Run: func(cmd *cobra.Command, args []string) {
		run(cmd.Context())
	},
}

func testsConfig() *internal.ConfigTests {
	b, err := os.ReadFile(testsFile)
	if err != nil {
		log.Fatal(err)
	}

	// load tests configuration
	cfgTests := &internal.ConfigTests{}
	if err := yaml.Unmarshal(b, cfgTests); err != nil {
		log.Fatal(err)
	}

	if cfgTests.Provider == "" {
		log.Fatal("no bmclib Provider defined in configuration")
	}

	if cfgTests.Protocol == "" {
		log.Fatal("no bmclib Protocol defined in configuration")
	}

	if !slices.Contains(internal.SupportedProviders(), cfgTests.Provider) {
		log.Fatalf("unsupported bmclib provider '%s' defined in test", cfgTests.Provider)
	}

	if len(cfgTests.Features) == 0 {
		log.Fatal("no bmclib features to test defined in configuration")
	}

	return cfgTests
}

func hardwareConfig() *internal.ConfigHardware {
	h, err := os.ReadFile(hardwareFile)
	if err != nil {
		log.Fatal(err)
	}

	// load hardware configuration
	cfgHardware := &internal.ConfigHardware{}
	if err := yaml.Unmarshal(h, cfgHardware); err != nil {
		log.Fatal(err)
	}

	if cfgHardware == nil || len(cfgHardware.Devices) == 0 {
		log.Fatal("no servers defined in configuration")
	}

	return cfgHardware
}

func run(ctx context.Context) {
	tests := testsConfig()
	hardware := hardwareConfig()

	testers := []*internal.Tester{}
	resultStore := internal.NewTestResultStore()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	wg := &sync.WaitGroup{}

	for _, server := range hardware.Devices {
		server := server
		tester := internal.NewTester(
			server.BmcHost,
			server.BmcUser,
			server.BmcPass,
			server.IpmiPort,
			logLevel,
		)

		testers = append(testers, tester)

		wg.Add(1)
		go func() {
			defer wg.Done()
			tester.Run(ctx, tests)

			result := internal.DeviceResult{
				Vendor:  server.Vendor,
				Model:   server.Model,
				Name:    server.Name,
				BMCIP:   server.BmcHost,
				Results: tester.Results(),
			}

			resultStore.Save(result)
		}()
	}

	log.Println("waiting for tests to complete...")

	wg.Wait()

	results := resultStore.Read()

	type pretty struct {
		BmclibVersion string                  `json:"bmclib_version"`
		Results       []internal.DeviceResult `json:"results"`
	}

	out := &pretty{bmclibVersion(), results}

	b, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

func bmclibVersion() string {
	buildInfo, ok := runtimedebug.ReadBuildInfo()
	if !ok {
		return ""
	}

	for _, d := range buildInfo.Deps {
		if strings.Contains(d.Path, "bmclib") {
			return d.Version
		}
	}

	return ""
}

func init() {
	runCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level")

	runCmd.PersistentFlags().StringVar(&testsFile, "tests", "", "YAML file with test configuration")
	runCmd.PersistentFlags().StringVar(&hardwareFile, "hardware", "", "YAML file with test configuration")
	runCmd.PersistentFlags().DurationVar(&timeout, "timeout", time.Minute*1, "Abort tests after timeout value")
	runCmd.MarkPersistentFlagRequired("tests")
	runCmd.MarkPersistentFlagRequired("hardware")

	rootCmd.AddCommand(runCmd)
}
