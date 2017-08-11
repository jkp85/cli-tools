package cmd

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	cmd := serverCmd()
	triggerCmd := serverTriggerCmd()
	triggerCmd.AddCommand(
		serverTriggerListCmd(),
		serverTriggerDescribeCmd(),
		serverTriggerCreateCmd(),
		serverTriggerUpdateCmd(),
		serverTriggerDeleteCmd(),
	)
	cmd.AddCommand(
		serverLsCmd(),
		serverCreateCmd(),
		serverUpdateCmd(),
		serverDescribeCmd(),
		serverStartCmd(),
		serverStopCmd(),
		serverTerminateCmd(),
		serverLogsCmd(),
		triggerCmd,
	)
	RootCmd.AddCommand(cmd)
}

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "User server management",
	}
	cmd.Flags().String("format", "json", "Output format")
	viper.BindPFlag("server_format", cmd.Flags().Lookup("format"))
	return cmd
}

func serverLsCmd() *cobra.Command {
	ls := &utils.ListFlags{}
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			resp, err := cli.ListServers(ls)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp)
		},
	}
	ls.Set(cmd)
	return cmd
}

func serverCreateCmd() *cobra.Command {
	body := &models.ServerData{
		Name:      new(string),
		Connected: []string{},
	}
	bodyConf := &models.ServerConfig{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsServersCreateParams()
			params.SetNamespace(cli.Namespace)
			body.Config = bodyConf
			params.SetServerData(body)
			projectID, err := cli.GetProjectID()
			if err != nil {
				return err
			}
			params.SetProjectID(projectID)
			resp, err := cli.Projects.ProjectsServersCreate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(body.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&body.ImageName, "image", "", "Server image")
	cmd.Flags().StringVar(&body.ServerSize, "resources", "", "Server resources")
	cmd.Flags().StringVar(&body.StartupScript, "startup-script", "", "Server startup script")
	cmd.Flags().StringVar(&bodyConf.Function, "function", "", "Function to run")
	cmd.Flags().StringVar(&bodyConf.Script, "script", "", "Script to run")
	cmd.Flags().StringVar(&bodyConf.Command, "command", "", "Command to run")
	cmd.Flags().StringVar(&bodyConf.Type, "type", "", "Server type [restful,cron,jupyter]")
	cmd.Flags().StringVar(&body.Host, "host", "", "Host id to run server on.")
	return cmd
}

func serverDescribeCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Server details",
		RunE: func(cmd *cobra.Command, args []string) error {
			var server *models.Server
			var err error
			cli := api.Client()
			if name == "" && serverID == "" {
				return errors.New("You must specify either name or id")
			}
			if serverID != "" {
				server, err = cli.GetServerByID(serverID)
			} else {
				server, err = cli.GetServerByName(name)
			}
			if err != nil {
				return err
			}
			return api.Render("server_format", server)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}

func serverUpdateCmd() *cobra.Command {
	var serverID, name string
	body := &models.ServerData{
		Connected: []string{},
		Name:      new(string),
	}
	bodyConf := &models.ServerConfig{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update server",
		RunE: func(cmd *cobra.Command, args []string) error {
			body.Config = bodyConf
			cli := api.Client()
			params := projects.NewProjectsServersUpdateParams()
			params.SetNamespace(cli.Namespace)
			err := setServerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			params.SetServerData(body)
			resp, err := cli.Projects.ProjectsServersUpdate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("server_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&name, "server-name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "server-id", "", "Server id")
	cmd.Flags().StringVar(body.Name, "name", "", "Server name")
	cmd.Flags().StringVar(&body.ImageName, "image", "", "Server image")
	cmd.Flags().StringVar(&body.StartupScript, "resources", "", "Server resources")
	cmd.Flags().StringVar(&body.StartupScript, "startup-script", "", "Server startup script")
	cmd.Flags().StringVar(&bodyConf.Function, "function", "", "Function to run")
	cmd.Flags().StringVar(&bodyConf.Script, "script", "", "Script to run")
	cmd.Flags().StringVar(&bodyConf.Command, "command", "", "Command to run")
	cmd.Flags().StringVar(&bodyConf.Type, "type", "", "Server type [restful,cron,jupyter]")
	return cmd
}

func serverStartCmd() *cobra.Command {
	var serverID, name string
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsServersStartParams()
			params.SetNamespace(cli.Namespace)
			err := setServerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			_, err = cli.Projects.ProjectsServersStart(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Server started")
			return nil
		},
	}
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	return cmd
}

func serverStopCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsServersStopParams()
			params.SetNamespace(cli.Namespace)
			err := setServerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			_, err = cli.Projects.ProjectsServersStop(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Server stopped")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}

func serverTerminateCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "Terminate server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsServersTerminateParams()
			params.SetNamespace(cli.Namespace)
			err := setServerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			_, err = cli.Projects.ProjectsServersTerminate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Server terminated")
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}

func serverLogsCmd() *cobra.Command {
	var name, serverID, logsURL string
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Server logs",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverID == "" && name == "" {
				return errors.New("You have to specify server id or name")
			}
			cli := api.Client()
			var server *models.Server
			var err error
			if serverID == "" {
				server, err = cli.GetServerByName(name)
			} else {
				server, err = cli.GetServerByID(serverID)
			}
			if err != nil {
				return err
			}
			serverID = server.ID
			logsURL = server.LogsURL
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)
			ws, err := url.Parse(logsURL)
			if err != nil {
				return err
			}
			header := make(http.Header)
			header.Add("Origin", viper.GetString("root"))
			c, _, err := websocket.DefaultDialer.Dial(ws.String(), header)
			if err != nil {
				return err
			}
			defer c.Close()
			done := make(chan struct{})
			go func() {
				defer c.Close()
				defer close(done)

				for {
					_, message, err := c.ReadMessage()
					if err != nil {
						return
					}
					jww.FEEDBACK.Println(string(message))
				}
			}()
			for {
				select {
				case <-interrupt:
					select {
					case <-done:
					case <-time.After(time.Second):
					}
					c.Close()
					return nil
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "uuid", "", "Server id")
	return cmd
}

func serverTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Handle server triggers",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("server_trigger_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func serverTriggerListCmd() *cobra.Command {
	var name, serverID string
	var lf utils.ListFlags
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List server triggers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewServiceTriggerListParams()
			params.SetNamespace(cli.Namespace)
			err := setTriggerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			lf.Apply(params)
			resp, err := cli.Projects.ServiceTriggerList(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("server_trigger_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	cmd.Flags().StringVar(&name, "server-name", "", "Server name")
	cmd.Flags().StringVar(&serverID, "server-id", "", "Server id")
	return cmd
}

func serverTriggerDescribeCmd() *cobra.Command {
	var serverName, serverID, triggerName, triggerID string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe server trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewServiceTriggerDeleteParams()
			params.SetNamespace(cli.Namespace)
			projectID, serverID, err := getPathIDs(serverID, serverName)
			if err != nil {
				return err
			}
			if triggerName == "" && triggerID == "" {
				return errors.New("You have to specify trigger id or name")
			}
			var trigger *models.ServerAction
			if triggerName != "" {
				trigger, err = cli.GetServerTriggerByName(projectID, serverID, triggerName)
			} else {
				trigger, err = cli.GetServerTriggerByID(projectID, serverID, triggerID)
			}
			if err != nil {
				return err
			}
			return api.Render("server_trigger_format", trigger)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&serverName, "server-name", "", "Server name")
	flags.StringVar(&serverID, "server-id", "", "Server id")
	flags.StringVar(&triggerName, "name", "", "Trigger name")
	flags.StringVar(&triggerID, "id", "", "Trigger id")
	return cmd
}

func serverTriggerCreateCmd() *cobra.Command {
	var name, serverID string
	body := &models.ServerAction{
		Webhook: &models.Webhook{
			URL: new(string),
		},
	}
	webhookPayload := api.NewJSONVal()
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create server trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewServiceTriggerCreateParams()
			params.SetNamespace(cli.Namespace)
			err := setTriggerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			body.Webhook.Payload = webhookPayload.Value
			params.SetServerAction(body)
			resp, err := cli.Projects.ServiceTriggerCreate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("server_trigger_format", resp.Payload)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&name, "server-name", "", "Server name")
	flags.StringVar(&serverID, "server-id", "", "Server id")
	flags.StringVar(&body.Name, "name", "", "Trigger name")
	flags.StringVar(&body.Operation, "operation", "", "Server operation [start, terminate]")
	flags.StringVar(body.Webhook.URL, "webhook-url", "", "Webhook url")
	flags.VarP(webhookPayload, "webhook-payload", "", "Webhook payload")
	return cmd
}

func serverTriggerUpdateCmd() *cobra.Command {
	var name, serverID string
	body := &models.ServerAction{
		Webhook: &models.Webhook{
			URL: new(string),
		},
	}
	webhookPayload := api.NewJSONVal()
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update server trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewServiceTriggerUpdateParams()
			params.SetNamespace(cli.Namespace)
			err := setTriggerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			body.Webhook.Payload = webhookPayload.Value
			params.SetServerAction(body)
			resp, err := cli.Projects.ServiceTriggerUpdate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("server_trigger_format", resp.Payload)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&name, "server-name", "", "Server name")
	flags.StringVar(&serverID, "server-id", "", "Server id")
	flags.StringVar(&body.Name, "name", "", "Trigger name")
	flags.StringVar(&body.Operation, "operation", "", "Server operation [start, terminate]")
	flags.StringVar(body.Webhook.URL, "webhook-url", "", "Webhook url")
	flags.VarP(webhookPayload, "webhook-payload", "", "Webhook payload")
	return cmd
}

func serverTriggerDeleteCmd() *cobra.Command {
	var name, serverID string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete server trigger",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewServiceTriggerDeleteParams()
			params.SetNamespace(cli.Namespace)
			err := setTriggerPathParams(params, serverID, name)
			if err != nil {
				return err
			}
			_, err = cli.Projects.ServiceTriggerDelete(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Trigger deleted")
			return nil
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&name, "server-name", "", "Server name")
	flags.StringVar(&serverID, "server-id", "", "Server id")
	return cmd
}

func getPathIDs(serverID, serverName string) (string, string, error) {
	cli := api.Client()
	var projectID string
	if serverID == "" && serverName == "" {
		return "", "", errors.New("You have to specify server id or name")
	}
	if serverID == "" {
		server, err := cli.GetServerByName(serverName)
		if err != nil {
			return "", "", err
		}
		serverID = server.ID
	}
	projectID, err := cli.GetProjectID()
	if err != nil {
		return "", "", err
	}
	return projectID, serverID, nil
}

type (
	ProjectIDSetter interface {
		SetProjectID(string)
	}
	ServerPathParamsSetter interface {
		ProjectIDSetter
		SetID(string)
	}
	TriggerPathParamsSetter interface {
		ProjectIDSetter
		SetServerID(string)
	}
)

func setServerPathParams(target ServerPathParamsSetter, serverID, serverName string) error {
	projectID, serverID, err := getPathIDs(serverID, serverName)
	if err != nil {
		return err
	}
	target.SetProjectID(projectID)
	target.SetID(serverID)
	return nil
}

func setTriggerPathParams(target TriggerPathParamsSetter, serverID, serverName string) error {
	projectID, serverID, err := getPathIDs(serverID, serverName)
	if err != nil {
		return err
	}
	target.SetProjectID(projectID)
	target.SetServerID(serverID)
	return nil
}
