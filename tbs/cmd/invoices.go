package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/billing"
)

func init() {
	cmd := invoiceCmd()
	cmd.AddCommand(invoiceListCmd(),
		       invoiceDescribeCmd())
	RootCmd.AddCommand(cmd)
}

func invoiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "invoice",
		Short: "View Invoices",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("invoice_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func invoiceListCmd() *cobra.Command {
	var lf utils.ListFlags
	cmd := &cobra.Command{
		Use: "ls",
		Short: "List the 10 most recent invoices",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingInvoicesListParams()
			params.SetNamespace(cli.Namespace)
			lf.Apply(params)

			resp, err := cli.Billing.BillingInvoicesList(params)

			if err != nil {
				return err
			}

			return api.Render("invoice_format", resp.Payload)
		},
	}

	lf.Set(cmd)
	return cmd
}

func invoiceDescribeCmd() *cobra.Command {
	var invoiceID string

	cmd := &cobra.Command{
		Use: "describe",
		Short: "Getails for an individual invoice.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingInvoicesReadParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(invoiceID)

			resp, err := cli.Billing.BillingInvoicesRead(params)

			if err != nil {
				return err
			}

			return api.Render("invoice_format", resp.Payload)
		},
	}

	cmd.Flags().StringVar(&invoiceID, "uuid", "", "Invoice ID")
	return cmd
}
