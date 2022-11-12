package cmd

import (
	"github.com/jatalocks/opsilon/pkg/run"

	"github.com/spf13/cobra"
)

type runOptions struct {
	file bool
}

func defaultRunOptions() *runOptions {
	return &runOptions{}
}

func newRunCmd() *cobra.Command {
	o := defaultRunOptions()

	cmd := &cobra.Command{
		Use:          "run",
		Short:        "run actions",
		SilenceUsage: true,
		RunE:         o.run,
	}

	cmd.Flags().BoolVarP(&o.file, "file", "f", o.file, "path to an actions file (override configure)")

	return cmd
}

func (o *runOptions) run(cmd *cobra.Command, args []string) error {
	values, err := o.parseArgs(args)
	if err != nil {
		return err
	}

	run.Select(values[0])

	return nil
}

func (o *runOptions) parseArgs(args []string) ([]string, error) {
	values := make([]string, 1) //nolint: gomnd

	for i, a := range args {
		values[i] = a
	}

	return values, nil
}
