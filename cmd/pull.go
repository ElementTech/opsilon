package cmd

import (
	"fmt"

	"github.com/jatalocks/opsilon/pkg/pull"
	"github.com/spf13/cobra"
)

const (
	pullNumberOfArgs = 1
)

type pullOptions struct {
	file bool
}

func defaultPullOptions() *pullOptions {
	return &pullOptions{}
}

func newPullCmd() *cobra.Command {
	o := defaultPullOptions()

	cmd := &cobra.Command{
		Use:          "pull",
		Short:        "pull a resource from your cloud",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(pullNumberOfArgs),
		RunE:         o.run,
	}

	cmd.Flags().BoolVarP(&o.file, "file", "f", o.file, "path to a resource file")

	return cmd
}

func (o *pullOptions) run(cmd *cobra.Command, args []string) error {
	values, err := o.parseArgs(args)
	if err != nil {
		return err
	}

	if o.file {
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", pull.Compare(values[0]))
	}

	return nil
}

func (o *pullOptions) parseArgs(args []string) ([]string, error) {
	values := make([]string, 1) //nolint: gomnd

	for i, a := range args {
		values[i] = a
	}

	return values, nil
}
