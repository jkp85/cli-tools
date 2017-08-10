package cmd

import (
	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/billing"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmd := planCmd()
	cmd.AddCommand(planListCmd(),
		planDescribeCmd(),
	)
	RootCmd.AddCommand(cmd)
}

func planCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "View plans, or manage them if you have the proper permissions.",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("plan_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func planListCmd() *cobra.Command {
	var lf utils.ListFlags
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List available plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingPlansListParams()
			params.SetNamespace(cli.Namespace)
			lf.Apply(params)
			resp, err := cli.Billing.BillingPlansList(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("plan_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	return cmd
}

func planDescribeCmd() *cobra.Command {
	var planID string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Plan Details",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingPlansReadParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(planID)

			resp, err := cli.Billing.BillingPlansRead(params, cli.AuthInfo)

			if err != nil {
				return err
			}

			return api.Render("plan_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&planID, "uuid", "", "Plan ID")

	return cmd
}
