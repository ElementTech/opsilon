package cmd

import (
	"github.com/jatalocks/opsilon/pkg/configure"

	"github.com/spf13/cobra"
)

const (
	configureNumberOfArgs = 1
)

type configureOptions struct {
	file bool
}

func defaultConfigureOptions() *configureOptions {
	return &configureOptions{}
}

func newConfigureCmd() *cobra.Command {
	o := defaultConfigureOptions()

	cmd := &cobra.Command{
		Use:          "configure",
		Short:        "configure cli with an opsilon.yaml file",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(configureNumberOfArgs),
		RunE:         o.configure,
	}

	cmd.Flags().BoolVarP(&o.file, "file", "f", o.file, "path to an actions file")

	return cmd
}

func (o *configureOptions) configure(cmd *cobra.Command, args []string) error {
	values, err := o.parseArgs(args)
	if err != nil {
		return err
	}

	if o.file {
		configure.Configure(values[0])
	}

	return nil
}

func (o *configureOptions) parseArgs(args []string) ([]string, error) {
	values := make([]string, 1) //nolint: gomnd

	for i, a := range args {
		values[i] = a
	}

	return values, nil
}
