package cmd

import (
	"fmt"

	"github.com/jatalocks/gitform/pkg/push"
	"github.com/spf13/cobra"
)

const (
	pushNumberOfArgs = 1
)

type pushOptions struct {
	file bool
}

func defaultPushOptions() *pushOptions {
	return &pushOptions{}
}

func newPushCmd() *cobra.Command {
	o := defaultPushOptions()

	cmd := &cobra.Command{
		Use:          "push",
		Short:        "push a resource to your cloud",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(pushNumberOfArgs),
		RunE:         o.run,
	}

	cmd.Flags().BoolVarP(&o.file, "file", "f", o.file, "path to a resource file")

	return cmd
}

func (o *pushOptions) run(cmd *cobra.Command, args []string) error {
	values, err := o.parseArgs(args)
	if err != nil {
		return err
	}

	if o.file {
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", push.Create(values[0]))
	}

	return nil
}

func (o *pushOptions) parseArgs(args []string) ([]string, error) {
	values := make([]string, 1) //nolint: gomnd

	for i, a := range args {
		values[i] = a
	}

	return values, nil
}
