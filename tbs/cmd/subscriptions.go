package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/go-sdk/client/billing"
	jww "github.com/spf13/jwalterweatherman"
	"fmt"
	"strings"
)

func init() {
	cmd := subscriptionCmd()
	cmd.AddCommand(subscriptionListCmd(),
	               subscriptionCreateCmd(),
	               subscriptionDescribeCmd(),
	               subscriptionDeleteCmd())
	RootCmd.AddCommand(cmd)
}

func subscriptionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "subscription",
		Short: "Manage your subscriptions",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("subscription_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func subscriptionListCmd() *cobra.Command {
	var lf utils.ListFlags
	cmd := &cobra.Command{
		Use: "ls",
		Short: "List Subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingSubscriptionsListParams()
			params.SetNamespace(cli.Namespace)
			lf.Apply(params)
			resp, err := cli.Billing.BillingSubscriptionsList(params)
			if err != nil {
				return err
			}
			return api.Render("subscription_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	return cmd
}

func subscriptionCreateCmd() *cobra.Command {
	body := billing.BillingSubscriptionsCreateBody{
		Plan: new(string),
	}
	cmd := &cobra.Command{
		Use: "create",
		Short: "Create new subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingSubscriptionsCreateParams()
			params.SetNamespace(cli.Namespace)
			params.SetData(body)
			resp, err := cli.Billing.BillingSubscriptionsCreate(params)

			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Subscription successfully created")
			return api.Render("subscription_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(body.Plan, "plan", "", "Plan ID")
	return cmd
}

func subscriptionDescribeCmd() *cobra.Command {
	var subscriptionID string

	cmd := &cobra.Command{
		Use: "describe",
		Short: "Subscription Details",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingSubscriptionsReadParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(subscriptionID)

			resp, err := cli.Billing.BillingSubscriptionsRead(params)

			if err != nil {
				return err
			}
			return api.Render("subscription_format", resp.Payload)

		},
	}

	cmd.Flags().StringVar(&subscriptionID, "uuid", "", "Subscription ID")
	return cmd
}

func subscriptionDeleteCmd() *cobra.Command {
	var subscriptionID string
	cmd := &cobra.Command{
		Use: "cancel",
		Short: "Cancel a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			confirm, err := readStdin(fmt.Sprint("Are you sure you want to cancel this subscription? (Y/n): "))
			if err != nil {
				return err
			}
			confirm = strings.ToLower(confirm)
			if confirm == "n" || confirm == "no" {
				jww.FEEDBACK.Println("Aborted")
				return nil
			}

			cli := api.Client()
			params := billing.NewBillingSubscriptionsDeleteParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(subscriptionID)

			_, err = cli.Billing.BillingSubscriptionsDelete(params)

			if err != nil {
				return err
			}

			jww.FEEDBACK.Println("Subscription canceled.")
			return nil
		},
	}

	cmd.Flags().StringVar(&subscriptionID, "uuid", "", "Subscription ID")
	return cmd
}
