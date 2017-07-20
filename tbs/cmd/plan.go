package cmd

import (
	"fmt"
	"strings"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/billing"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	cmd := planCmd()
	cmd.AddCommand(planListCmd(),
		planCreateCmd(),
		planUpdateCmd(),
		planDescribeCmd(),
		planDeleteCmd())
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

func planCreateCmd() *cobra.Command {
	body := &models.PlanData{
		Amount:        new(int64),
		Interval:      new(string),
		IntervalCount: new(int64),
		Name:          new(string),
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a plan",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingPlansCreateParams()
			params.SetNamespace(cli.Namespace)
			params.SetData(body)
			resp, err := cli.Billing.BillingPlansCreate(params, cli.AuthInfo)

			if err != nil {
				return err
			}

			jww.FEEDBACK.Println("Plan successfully created")
			return api.Render("plan_format", resp.Payload)
		},
	}
	cmd.Flags().Int64Var(body.Amount, "amount", 0, "Amount, in cents, the plan will cost.")
	cmd.Flags().StringVar(body.Interval, "interval", "month", "day|week|month|year")
	cmd.Flags().Int64Var(body.IntervalCount, "interval_count", 0, "The number of intervals between each billing")
	cmd.Flags().StringVar(body.Name, "name", "", "Name of the plan.")
	cmd.Flags().StringVar(&body.Currency, "currency", "usd", "ISO currency code that the plan should be billed in, e.g. usd")
	cmd.Flags().StringVar(&body.StatementDescriptor, "statement_descriptor", "", "Additional info that will show on customer's credit card statement.")
	cmd.Flags().Int64Var(&body.TrialPeriodDays, "trial_period", 0, "Length of the plan's trial period, in days.")

	return cmd
}

func planUpdateCmd() *cobra.Command {
	var planID string
	updateBody := &models.PlanData{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update plan information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()

			params := billing.NewBillingPlansUpdateParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(planID)
			params.SetData(updateBody)

			resp, err := cli.Billing.BillingPlansUpdate(params, cli.AuthInfo)

			if err != nil {
				return err
			}

			jww.FEEDBACK.Println("Plan Updated")
			return api.Render("plan_format", resp.Payload)
		},
	}

	cmd.Flags().StringVar(&planID, "uuid", "", "Plan ID")
	cmd.Flags().StringVar(updateBody.Name, "name", "", "Name of the plan.")
	cmd.Flags().StringVar(&updateBody.StatementDescriptor, "statement_descriptor", "", "Additional info that will show on customer's credit card statement.")
	cmd.Flags().Int64Var(&updateBody.TrialPeriodDays, "trial_period", 0, "Length of the plan's trial period, in days.")

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

func planDeleteCmd() *cobra.Command {
	var planID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a plan. Also deletes all subscriptions to the plan.",
		RunE: func(cmd *cobra.Command, vars []string) error {
			confirm, err := readStdin(fmt.Sprint("Are you sure you want to delete this plan? (Y/n): "))
			if err != nil {
				return err
			}
			confirm = strings.ToLower(confirm)
			if confirm == "n" || confirm == "no" {
				jww.FEEDBACK.Println("Aborted")
				return nil
			}

			cli := api.Client()
			params := billing.NewBillingPlansDeleteParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(planID)

			_, err = cli.Billing.BillingPlansDelete(params, cli.AuthInfo)

			if err != nil {
				return err
			}

			jww.FEEDBACK.Println("Plan deleted")
			return nil
		},
	}

	cmd.Flags().StringVar(&planID, "uuid", "", "Plan ID")
	return cmd
}
