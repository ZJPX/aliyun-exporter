package cmd

import (
	"fmt"
	"net/http"
	"os"
	"text/tabwriter"

	"aliyun-exporter/pkg/client"
	"aliyun-exporter/pkg/config"
	"aliyun-exporter/pkg/cron"
	"aliyun-exporter/pkg/handler"
	"aliyun-exporter/version"

	"github.com/spf13/cobra"
)

const AppName = "cloudmonitor"

// NewRootCommand create root command
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           AppName,
		Short:         "Exporter for aliyun cloudmonitor",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	cmd.AddCommand(newServeMetricsCommand())
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newListMetricNamespacesCommand())
	return cmd
}

func newServeMetricsCommand() *cobra.Command {
	opt := &options{}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve HTTP metrics handler",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return opt.Complete()
		},
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := config.Parse(opt.configFile)
			if err != nil {
				return err
			}

			err = cron.New(logger, cfg)
			if err != nil {
				return err
			}

			handler.RegisterHandler(cfg.Metrics)
			return http.ListenAndServe(opt.listenAddress, nil)
		},
	}
	opt.AddFlags(cmd)
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(version.Version())
		},
	}
}

func newListMetricNamespacesCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-metrics",
		Short: "List avaliable namespaces of metrics",
		Run: func(_ *cobra.Command, _ []string) {
			w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
			fmt.Fprintln(w, "NAMESPACE\tDESCRIPTION")
			for name, desc := range client.AllNamespaces() {
				fmt.Fprintf(w, "%s\t%s\n", name, desc)
			}
			w.Flush()
		},
	}
}
