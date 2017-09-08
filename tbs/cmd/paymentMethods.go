package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/billing"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/token"
)

func init() {
	cmd := billingCmd()
	cmd.AddCommand(billingListCardCmd(),
		billingDescribeCardCmd(),
		billingUpdateCardCmd(),
		billingDeleteCardCmd(),
		billingCreateCardInteractiveCmd())
	RootCmd.AddCommand(cmd)
}

func billingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Handle Credit Cards",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("billing_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func billingListCardCmd() *cobra.Command {
	var lf utils.ListFlags
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List payment methods",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingCardsListParams()
			params.SetNamespace(cli.Namespace)
			lf.Apply(params)
			resp, err := cli.Billing.BillingCardsList(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("billing_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	return cmd
}

func billingDescribeCardCmd() *cobra.Command {
	var cardID string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Credit Card Details",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := billing.NewBillingCardsReadParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(cardID)
			resp, err := cli.Billing.BillingCardsRead(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("billing_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&cardID, "uuid", "", "Card id")
	return cmd
}

func billingUpdateCardCmd() *cobra.Command {
	var cardID string
	updateBody := &models.CardDataPutandPatch{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update credit card information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()

			params := billing.NewBillingCardsUpdateParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(cardID)
			params.SetCardData(updateBody)

			resp, err := cli.Billing.BillingCardsUpdate(params, cli.AuthInfo)
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
		Use:   "rm",
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

			_, err = cli.Billing.BillingCardsDelete(params, cli.AuthInfo)
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

func billingCreateCardInteractiveCmd() *cobra.Command {
	body := BillingCardsCreateBody{}
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a credit card.",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := readStdin("Cardholder Name:")
			if err != nil {
				return err
			}
			body.Name = name

			number, err := readStdin("Card Number:")
			if err != nil {
				return err
			}
			body.Number = number

			var expMonth int64

			fmt.Println("Expiry Month:")
			fmt.Scan(&expMonth)
			body.ExpMonth = expMonth

			var expYear int64

			fmt.Println("Expiry Year:")
			fmt.Scan(&expYear)
			body.ExpYear = expYear

			Cvc, err := readStdin("CVC: ")
			if err != nil {
				return err
			}
			body.Cvc = Cvc

			address1, err := readStdin("Address Line 1: ")
			if err != nil {
				return err
			}
			body.AddressLine1 = address1

			address2, err := readStdin("Address Line 2: ")
			if err != nil {
				return err
			}
			body.AddressLine2 = address2

			city, err := readStdin("City: ")
			if err != nil {
				return err
			}
			body.AddressCity = city

			state, err := readStdin("State: ")
			if err != nil {
				return err
			}
			body.AddressState = state

			zip, err := readStdin("Zip Code: ")
			if err != nil {
				return err
			}
			body.AddressZip = zip

			country, err := readStdin("Country: ")
			if err != nil {
				return err
			}
			body.AddressCountry = country

			stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

			tkn, err := token.New(&stripe.TokenParams{
				Card: &stripe.CardParams{
					Number:   body.Number,
					Month:    strconv.FormatInt(body.ExpMonth, 10),
					Year:     strconv.FormatInt(body.ExpYear, 10),
					CVC:      body.Cvc,
					City:     body.AddressCity,
					Country:  body.AddressCountry,
					Address1: body.AddressLine1,
					Address2: body.AddressLine2,
					State:    body.AddressState,
					Zip:      body.AddressZip,
					Name:     body.Name,
				},
			})

			cli := api.Client()
			params := billing.NewBillingCardsCreateParams()
			params.SetNamespace(cli.Namespace)
			body := &models.CardDataPost{
				Token: tkn.ID,
			}
			params.SetCardData(body)

			resp, err := cli.Billing.BillingCardsCreate(params, cli.AuthInfo)

			jww.FEEDBACK.Println("Card added")
			return api.Render("billing_format", resp.Payload)
		},
	}

	return cmd
}

type BillingCardsCreateBody struct {

	// address city
	AddressCity string `json:"address_city,omitempty"`

	// address country
	AddressCountry string `json:"address_country,omitempty"`

	// address line1
	AddressLine1 string `json:"address_line1,omitempty"`

	// address line2
	AddressLine2 string `json:"address_line2,omitempty"`

	// address state
	AddressState string `json:"address_state,omitempty"`

	// address zip
	AddressZip string `json:"address_zip,omitempty"`

	Cvc string `json:"cvc,omitempty"`

	// exp month
	ExpMonth int64 `json:"exp_month,omitempty"`

	// exp year
	ExpYear int64 `json:"exp_year,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	Number string `json:"number,omitempty"`

	// token
	Token string `json:"token,omitempty"`
}
