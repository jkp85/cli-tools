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
		createTriggerCmd(),
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

func createTrigger(body *models.TriggerData) (*triggers.TriggersCreateCreated, error) {
	cli := api.Client()
	params := triggers.NewTriggersCreateParams()
	params.SetNamespace(cli.Namespace)
	params.SetData(body)
	return cli.Triggers.TriggersCreate(params, cli.AuthInfo)
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
	body := &models.TriggerData{
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

func createTriggerCmd() *cobra.Command {
	cause := &models.TriggerAction{
		ActionName: new(string),
		Method:     new(string),
	}
	effect := &models.TriggerAction{
		ActionName: new(string),
		Method:     new(string),
	}
	webhook := &models.Webhook{
		URL: new(string),
	}
	causePayload := api.NewJSONVal()
	effectPayload := api.NewJSONVal()
	webhookConfig := api.NewJSONVal()
	schedule := ""
	body := &models.TriggerData{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			if *effect.ActionName != "" && *effect.Method != "" {
				body.Effect = effect
				body.Effect.Payload = effectPayload.Value
			}
			if *cause.ActionName != "" && *cause.Method != "" {
				body.Cause = cause
				body.Cause.Payload = causePayload.Value
			}
			if *webhook.URL != "" {
				body.Webhook = webhook
				body.Webhook.Config = webhookConfig.Value
			}
			if schedule != "" {
				body.Schedule = schedule
			}
			resp, err := createTrigger(body)
			if err != nil {
				return err
			}
			return api.Render("trigger_format", resp)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(cause.ActionName, "cause-action", "", "Cause action")
	flags.StringVar(cause.Method, "cause-method", "", "Cause method")
	flags.StringVar(&cause.Model, "cause-model", "", "Cause type")
	flags.StringVar(&cause.ObjectID, "cause-object", "", "Cause object")
	flags.VarP(causePayload, "cause-payload", "", "Cause payload")
	flags.StringVar(effect.ActionName, "effect-action", "", "Effect action")
	flags.StringVar(effect.Method, "effect-method", "", "Effect method")
	flags.StringVar(&effect.Model, "effect-model", "", "Effect type")
	flags.StringVar(&effect.ObjectID, "effect-object", "", "Effect object")
	flags.VarP(effectPayload, "effect-payload", "", "Effect payload")
	flags.StringVar(webhook.URL, "webhook-url", "", "Webhook url")
	flags.VarP(webhookConfig, "webhook-congig", "", "Webhook config")
	flags.StringVar(&schedule, "schedule", "", "Cron schedule")
	return cmd
}
