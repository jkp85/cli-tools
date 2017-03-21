package cmd

import (
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	RootCmd.AddCommand(envCmd())

}

const tmpl = `export THREEBLADES_PROJECT=%s
export THREEBLADES_NAMESPACE=%s
# Run this command to configure your shell:
# eval $(tbs env --project=<project_name> --namespace=<namespace>)
`

func envCmd() *cobra.Command {
	var projectName, namespace string
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Prints env variables for later use",
		RunE: func(cmd *cobra.Command, args []string) error {
			jww.FEEDBACK.Printf(tmpl, projectName, namespace)
			return nil
		},
	}
	cmd.Flags().StringVar(&projectName, "project", "", "Project name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace")
	return cmd
}
