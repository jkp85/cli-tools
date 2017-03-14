package cmd

import (
	"errors"
	"os"

	"github.com/jkp85/cli-tools/api"
	"github.com/jkp85/go-sdk/client/projects"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func projectsCmd() *cobra.Command {
	var limit, offset, format string
	filters := api.NewFilterVal()
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Handle projects",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.APIClient()
			params := projects.NewProjectsListParams()
			ns := viper.GetString("namespace")
			if ns == "" {
				return errors.New("You must provide a namespace")
			}
			params.SetNamespace(ns)
			params.SetLimit(&limit)
			params.SetOffset(&offset)
			params.SetPrivate(filters.Get("private"))
			resp, err := cli.Projects.ProjectsList(params)
			if err != nil {
				return err
			}
			if format == "" {
				format = viper.GetString("projects_template")
			}
			r := api.NewRenderer(format, &resp.Payload)
			return r.Render(os.Stdout)
		},
	}
	cmd.Flags().StringVar(&limit, "limit", "10", "Limit list results")
	cmd.Flags().StringVar(&offset, "offset", "0", "Offset list results")
	cmd.Flags().StringVar(&format, "format", "json", "Output format")
	cmd.Flags().Var(filters, "filter", "Filter results")
	return cmd
}

func init() {
	RootCmd.AddCommand(projectsCmd())
}
