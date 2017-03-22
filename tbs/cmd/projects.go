package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jkp85/cli-tools/tbs/api"
	"github.com/jkp85/cli-tools/tbs/utils"
	"github.com/jkp85/go-sdk/client/projects"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	projectCmd := projectsCmd()
	projectCmd.AddCommand(projectCreateCmd())
	projectCmd.AddCommand(projectUpdateCmd())
	projectCmd.AddCommand(projectDeleteCmd())
	projectCmd.AddCommand(addUserToProjectCmd())
	RootCmd.AddCommand(projectCmd)
}

func projectsCmd() *cobra.Command {
	var lf utils.ListFlags
	filters := api.NewFilterVal()
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Handle projects",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsListParams()
			ns := viper.GetString("namespace")
			if ns == "" {
				return errors.New("You must provide a namespace")
			}
			params.SetNamespace(ns)
			lf.Apply(params)
			params.SetName(filters.Get("name"))
			params.SetPrivate(filters.Get("private"))
			resp, err := cli.Projects.ProjectsList(params)
			if err != nil {
				return err
			}
			return api.Render("project_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	cmd.Flags().Var(filters, "filter", "Filter results")
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("project_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func projectCreateCmd() *cobra.Command {
	var members []string
	body := projects.ProjectsCreateBody{
		Name: new(string),
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if *body.Name == "" {
				return errors.New("You need to provide name for your project")
			}
			cli := api.Client()
			params := projects.NewProjectsCreateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			params.SetData(body)
			resp, err := cli.Projects.ProjectsCreate(params)
			if err != nil {
				return fmt.Errorf("There was an error creating project: %s\n", err)
			}
			err = addMembers(resp.Payload.ID, members...)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Project successfully created")
			return api.Render("project_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(body.Name, "name", "", "Project name")
	cmd.Flags().StringVar(&body.Description, "description", "", "Project description")
	cmd.Flags().BoolVar(&body.Private, "privacy", false, "Should this project be private?")
	cmd.Flags().StringSliceVar(&members, "members", []string{}, "Project members")
	return cmd
}

func getProjectIDByName(name string) (string, error) {
	cli := api.Client()
	params := projects.NewProjectsListParams()
	ns := viper.GetString("namespace")
	params.SetNamespace(ns)
	params.SetName(&name)
	resp, err := cli.Projects.ProjectsList(params)
	if err != nil {
		return "", err
	}
	if len(resp.Payload) == 0 {
		return "", fmt.Errorf("There is no project with name: '%s'", name)
	}
	return resp.Payload[0].ID, nil
}

func projectDeleteCmd() *cobra.Command {
	var projectID, projectName string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete project",
		RunE: func(cmd *cobra.Command, args []string) error {
			confirm, err := readStdin(
				fmt.Sprintf("Are you sure you want to delete project '%s'? (Y/n): ", projectName))
			if err != nil {
				return err
			}
			confirm = strings.ToLower(confirm)
			if confirm == "n" || confirm == "no" {
				jww.FEEDBACK.Println("Aborted")
				return nil
			}
			cli := api.Client()
			ns := viper.GetString("namespace")
			if projectID == "" {
				projectID, err = getProjectIDByName(projectName)
				if err != nil {
					return err
				}
			}
			deleteParams := projects.NewProjectsDeleteParams()
			deleteParams.SetNamespace(ns)
			deleteParams.SetID(projectID)
			_, err = cli.Projects.ProjectsDelete(deleteParams)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Project deleted")
			return nil
		},
	}
	cmd.Flags().StringVar(&projectName, "name", "", "Project name")
	cmd.Flags().StringVar(&projectID, "uuid", "", "Project uuid")
	return cmd
}

func addUserToProjectCmd() *cobra.Command {
	var projectName, format string
	cmd := &cobra.Command{
		Use:   "adduser [email]",
		Short: "Add collaborator to project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("You must specify user email")
			}
			email := args[0]
			projectID, err := getProjectIDByName(projectName)
			if err != nil {
				return err
			}
			return addMembers(projectID, email)
		},
	}
	cmd.Flags().StringVar(&format, "format", "json", "Collaborator format")
	cmd.Flags().StringVar(&projectName, "project", "", "Project name")
	return cmd
}

func addMembers(projectID string, members ...string) error {
	cli := api.Client()
	for _, member := range members {
		params := projects.NewProjectsCollaboratorsCreateParams()
		ns := viper.GetString("namespace")
		params.SetNamespace(ns)
		data := projects.ProjectsCollaboratorsCreateBody{
			Owner: false,
			Email: &member,
		}
		params.SetProjectPk(projectID)
		params.SetData(data)
		_, err := cli.Projects.ProjectsCollaboratorsCreate(params)
		if err != nil {
			jww.ERROR.Printf("Error adding memeber: %s\n", member)
			continue
		}
		jww.FEEDBACK.Printf("Member added: %s\n", member)
	}
	return nil
}

func projectUpdateCmd() *cobra.Command {
	var projectID string
	var members []string
	updateBody := projects.ProjectsPartialUpdateBody{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update project",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if updateBody.Name == "" || projectID == "" {
				return errors.New("You must provide either project name or id")
			}
			if projectID == "" {
				projectID, err = getProjectIDByName(updateBody.Name)
			}
			err = addMembers(projectID, members...)
			if err != nil {
				return err
			}
			cli := api.Client()
			params := projects.NewProjectsPartialUpdateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			params.SetID(projectID)
			params.SetData(updateBody)
			resp, err := cli.Projects.ProjectsPartialUpdate(params)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Project updated.")
			return api.Render("project_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&projectID, "uuid", "", "Project id")
	cmd.Flags().StringVar(&updateBody.Name, "name", "", "Project name")
	cmd.Flags().StringVar(&updateBody.Description, "description", "", "Project description")
	cmd.Flags().BoolVar(&updateBody.Private, "privacy", false, "Should this project be private?")
	cmd.Flags().StringSliceVar(&members, "members", []string{}, "Project members")
	return cmd
}
