package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	RootCmd.AddCommand(envCmd())

}

const (
	projectTmpl   = `export THREEBLADES_PROJECT=%s`
	namespaceTmlp = `export THREEBLADES_NAMESPACE=%s`
	infoTmpl      = "\n# Run this command to configure your shell:\n# eval $(%s)"
)

func envCmd() *cobra.Command {
	var projectName, namespace string
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Prints env variables for later use",
		RunE: func(cmd *cobra.Command, args []string) error {
			var out string
			cmdTmpl := "tbs env"
			if namespace != "" {
				out += fmt.Sprintf(namespaceTmlp, namespace)
				cmdTmpl += fmt.Sprintf(" --namespace=%s", namespace)
			}
			if projectName != "" {
				out += fmt.Sprintf(projectTmpl, projectName)
				cmdTmpl += fmt.Sprintf(" --project=%s", projectName)
			}
			out += fmt.Sprintf(infoTmpl, cmdTmpl)
			jww.FEEDBACK.Println(out)
			return nil
		},
	}
	cmd.Flags().StringVar(&projectName, "project", "", "Project name")
	cmd.Flags().StringVar(&namespace, "namespace", "", "Namespace")
	return cmd
}
