package cmd

import ("strconv"
	"github.com/spf13/cobra"
	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/go-sdk/client/billing"
	"github.com/spf13/viper"
	"github.com/3Blades/cli-tools/tbs/utils"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/stripe/stripe-go"
	"fmt"
	"github.com/stripe/stripe-go/token"
	"strings"
	"os"
	"github.com/3Blades/go-sdk/models"
)


func init(){
	cmd := billingCmd()
	cmd.AddCommand(billingListCardCmd(),
	               billingCreateCardCmd(),
	               billingDescribeCardCmd(),
		       billingUpdateCardCmd(),
		       billingDeleteCardCmd(),
	               billingSetDefaultCardCmd())
	RootCmd.AddCommand(cmd)
}

func billingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "billing",
		Short: "Handle Credit Cards",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("billing_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func billingListCardCmd() *cobra.Command {
	var lf utils.ListFlags
	cmd := &cobra.Command{
		Use: "ls",
		Short: "List payment methods",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingCardsListParams()
			params.SetNamespace(cli.Namespace)
			lf.Apply(params)
			resp, err := cli.Billing.BillingCardsList(params)
			if err != nil {
				return err
			}
			return api.Render("billing_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	return cmd
}

func billingCreateCardCmd() *cobra.Command {
	body := billing.BillingCardsCreateBody{}
	cmd := &cobra.Command{
		Use: "create",
		Short: "Create a new Credit Card",
		RunE: func(cmd *cobra.Command, args []string) error {
			stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

			tkn, err := token.New(&stripe.TokenParams{
					Card: &stripe.CardParams{
						Number: body.Number,
						Month: strconv.FormatInt(body.ExpMonth, 10),
						Year: strconv.FormatInt(body.ExpYear, 10),
						CVC: body.Cvc,
						City: body.AddressCity,
						Country: body.AddressCountry,
						Address1: body.AddressLine1,
						Address2: body.AddressLine2,
						State: body.AddressState,
						Zip: body.AddressZip,
						Name: body.Name,
						},
			})

			if err != nil {
				return err
			}

			cli := api.Client()
			params := billing.NewBillingCardsCreateParams()
			params.SetNamespace(cli.Namespace)
			body := billing.BillingCardsCreateBody{
				Token: tkn.ID,
			}
			params.SetData(body)

			resp, err := cli.Billing.BillingCardsCreate(params)

			if err != nil {
				return err
			}

			jww.FEEDBACK.Println("Credit card successfully created")

			return api.Render("billing_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&body.Name, "name", "", "Cardholder name")
	cmd.Flags().StringVar(&body.Number, "number", "", "Card Number")
	cmd.Flags().Int64Var(&body.ExpMonth, "month", -1, "Expiration Month")
	cmd.Flags().Int64Var(&body.ExpYear, "year", -1, "Expiration Year")
	cmd.Flags().StringVar(&body.Cvc, "cvc", "", "CVC Code")
	cmd.Flags().StringVar(&body.AddressLine1, "address_line1", "", "Address Line One")
	cmd.Flags().StringVar(&body.AddressLine2, "address_line2", "", "Address Line Two")
	cmd.Flags().StringVar(&body.AddressCity, "city", "", "City")
	cmd.Flags().StringVar(&body.AddressState, "state", "", "State")
	cmd.Flags().StringVar(&body.AddressCountry, "country", "", "Country")
	cmd.Flags().StringVar(&body.AddressZip, "zip_code", "", "ZIP Code")

	return cmd
}

func billingDescribeCardCmd() *cobra.Command {
	var cardID string
	cmd := &cobra.Command{
		Use: "describe",
		Short: "Credit Card Details",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingCardsReadParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(cardID)
			resp, err := cli.Billing.BillingCardsRead(params)
			if err != nil {
				return err
			}
			return api.Render("billing_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&cardID, "uuid", "", "Card id")
	return cmd
}

func billingUpdateCardCmd() *cobra.Command{
	var cardID string
	updateBody := billing.BillingCardsPartialUpdateBody{}
	cmd := &cobra.Command{
		Use: "update",
		Short: "Update credit card information.",
		RunE:  func(cmd *cobra.Command, args []string) error {
			cli := api.Client()

			params := billing.NewBillingCardsPartialUpdateParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(cardID)
			params.SetData(updateBody)

			resp, err := cli.Billing.BillingCardsPartialUpdate(params)
			if err != nil {
				return err
			}

			jww.FEEDBACK.Println("Card Updated.")
			return api.Render("billing_format", resp.Payload)

		},
	}
	cmd.Flags().StringVar(&cardID, "uuid", "", "Card ID")
	cmd.Flags().StringVar(&updateBody.Name, "name", "", "Cardholder name")
	cmd.Flags().StringVar(&updateBody.AddressLine1, "address_line1", "", "Address Line One")
	cmd.Flags().StringVar(&updateBody.AddressLine2, "address_line2", "", "Address Line Two")
	cmd.Flags().StringVar(&updateBody.AddressCity, "city", "", "City")
	cmd.Flags().StringVar(&updateBody.AddressState, "state", "", "State")
	cmd.Flags().StringVar(&updateBody.AddressCountry, "country", "", "Country")
	cmd.Flags().StringVar(&updateBody.AddressZip, "zip_code", "", "ZIP Code")
	return cmd
}

func billingDeleteCardCmd() *cobra.Command {
	var cardID string
	cmd := &cobra.Command{
		Use: "delete",
		Short: "Delete a credit card.",
		RunE: func(cmd *cobra.Command, args []string) error {
			confirm, err := readStdin(fmt.Sprint("Are you sure you want to delete this card? (Y/n): "))
			if err != nil {
				return err
			}
			confirm = strings.ToLower(confirm)
			if confirm == "n" || confirm == "no" {
				jww.FEEDBACK.Println("Aborted")
				return nil
			}

			cli := api.Client()
			params := billing.NewBillingCardsDeleteParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(cardID)

			_, err = cli.Billing.BillingCardsDelete(params)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Card deleted")
			return nil
		},
	}

	cmd.Flags().StringVar(&cardID, "uuid", "", "Card UUID")
	return cmd
}

func getCustomerByUser(username string) (*models.Customer, error) {
	// user, err := getUserByName(username)
	// if err != nil {
	// 	return err
	// }

	cli := api.Client()
	namespace := viper.GetString("namespace")
	params := billing.NewBillingCustomersListParams()
	params.SetNamespace(namespace)
	resp, err := cli.Billing.BillingCustomersList(params)

	if err != nil {
		return &models.Customer{}, err
	}

	return resp.Payload[0], nil

}


func billingSetDefaultCardCmd() *cobra.Command {
	updateBody := billing.BillingCustomersPartialUpdateBody{}
	cmd := &cobra.Command{
		Use: "set-default",
		Short: "Set your default payment method.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingCustomersPartialUpdateParams()
			params.SetNamespace(cli.Namespace)

			customer, err := getCustomerByUser(cli.Namespace)

			if err != nil {
				return err
			}

			params.SetID(customer.ID)
			params.SetData(updateBody)

			resp, err := cli.Billing.BillingCustomersPartialUpdate(params)

			if err != nil {
				return err
			}

			jww.FEEDBACK.Println("Default Payment Source Set.")
			return api.Render("billing_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&updateBody.DefaultSource, "uuid", "", "Card ID")
	return cmd
}
