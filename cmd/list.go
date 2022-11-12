package cmd

import (
	"github.com/jatalocks/opsilon/pkg/list"

	"github.com/spf13/cobra"
)

type listOptions struct {
	file bool
}

func defaultListOptions() *listOptions {
	return &listOptions{}
}

func newListCmd() *cobra.Command {
	o := defaultListOptions()

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "list actions",
		SilenceUsage: true,
		RunE:         o.list,
	}

	cmd.Flags().BoolVarP(&o.file, "file", "f", o.file, "path to an actions file")

	return cmd
}

func (o *listOptions) list(cmd *cobra.Command, args []string) error {
	values, err := o.parseArgs(args)
	if err != nil {
		return err
	}

	list.List(values[0])

	return nil
}

func (o *listOptions) parseArgs(args []string) ([]string, error) {
	values := make([]string, 1) //nolint: gomnd

	for i, a := range args {
		values[i] = a
	}

	return values, nil
}
