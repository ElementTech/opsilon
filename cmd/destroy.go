package cmd

import (
	"fmt"

	"github.com/jatalocks/opsilon/pkg/destroy"
	"github.com/spf13/cobra"
)

const (
	destroyNumberOfArgs = 1
)

type destroyOptions struct {
	file bool
}

func defaultDestroyOptions() *destroyOptions {
	return &destroyOptions{}
}

func newDestroyCmd() *cobra.Command {
	o := defaultDestroyOptions()

	cmd := &cobra.Command{
		Use:          "destroy",
		Short:        "destroy a resource from your cloud",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(destroyNumberOfArgs),
		RunE:         o.run,
	}

	cmd.Flags().BoolVarP(&o.file, "file", "f", o.file, "path to a resource file")

	return cmd
}

func (o *destroyOptions) run(cmd *cobra.Command, args []string) error {
	values, err := o.parseArgs(args)
	if err != nil {
		return err
	}

	if o.file {
		fmt.Fprintf(cmd.OutOrStdout(), "%d\n", destroy.Destroy(values[0]))
	}

	return nil
}

func (o *destroyOptions) parseArgs(args []string) ([]string, error) {
	values := make([]string, 1) //nolint: gomnd

	for i, a := range args {
		values[i] = a
	}

	return values, nil
}
