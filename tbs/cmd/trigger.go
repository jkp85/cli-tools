package cmd

import (
	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/go-sdk/client/triggers"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cmd := triggerCmd()
	cmd.AddCommand(
		sendSlackMessage(),
	)
	RootCmd.AddCommand(cmd)
}

func triggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Handle triggers",
	}
	cmd.Flags().String("format", "json", "Output format")
	viper.BindPFlag("trigger_format", cmd.Flags().Lookup("format"))
	return cmd
}

func createTrigger(body triggers.TriggersCreateBody) (*triggers.TriggersCreateCreated, error) {
	cli := api.Client()
	params := triggers.NewTriggersCreateParams()
	params.SetNamespace(cli.Namespace)
	params.SetData(body)
	return cli.Triggers.TriggersCreate(params)
}

type SlackWebhookConfig struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	IconURL  string `json:"icon_url"`
	Channel  string `json:"channel"`
}

func sendSlackMessage() *cobra.Command {
	whConf := &SlackWebhookConfig{
		Username: "3blades-bot",
	}
	body := triggers.TriggersCreateBody{
		Cause: &models.TriggerAction{
			ActionName: new(string),
			Method:     new(string),
		},
		Webhook: &models.Webhook{
			URL: new(string),
		},
	}
	cmd := &cobra.Command{
		Use:   "slack",
		Short: "Send slack message after an event",
		RunE: func(cmd *cobra.Command, args []string) error {
			body.Webhook.Config = whConf
			resp, err := createTrigger(body)
			if err != nil {
				return err
			}
			return api.Render("trigger_format", resp)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(body.Webhook.URL, "webhook", "", "Slack webhook url")
	flags.StringVar(&whConf.Text, "text", "", "Text to send")
	flags.StringVar(&whConf.Channel, "channel", "", "Channel to send to")
	flags.StringVar(body.Cause.ActionName, "action", "", "Cause action")
	flags.StringVar(body.Cause.Method, "method", "", "Cause method")
	flags.StringVar(&body.Cause.Model, "model", "", "Cause type")
	flags.StringVar(&body.Cause.ObjectID, "object", "", "Cause object")
	return cmd
}
