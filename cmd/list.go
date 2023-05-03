package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/bmc-toolbox/bmclib-tester/internal"
	"github.com/spf13/cobra"
)

// listCmd lists the bmclib tests configured.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list tests configured",
	Run: func(cmd *cobra.Command, args []string) {
		list()
	},
}

func list() {
	tests := testsConfig()
	hardware := hardwareConfig()

	type pretty struct {
		Hardware *internal.ConfigHardware `json:"Hardware"`
		Tests    *internal.ConfigTests    `json:"Tests"`
	}

	p := pretty{hardware, tests}

	b, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

func init() {

	listCmd.PersistentFlags().StringVar(&testsFile, "tests", "", "YAML file with test configuration")
	listCmd.PersistentFlags().StringVar(&hardwareFile, "hardware", "", "YAML file with test configuration")
	listCmd.MarkPersistentFlagRequired("tests")
	listCmd.MarkPersistentFlagRequired("hardware")

	rootCmd.AddCommand(listCmd)
}
