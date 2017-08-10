package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/cli-tools/tbs/utils"
	"github.com/3Blades/go-sdk/client/projects"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	cmd := projectsCmd()
	cmd.AddCommand(
		projectListCmd(),
		projectCreateCmd(),
		projectUpdateCmd(),
		projectDeleteCmd(),
		addUserToProjectCmd(),
	)
	RootCmd.AddCommand(cmd)
}

func projectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Handle projects",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "Output format")
	viper.BindPFlag("project_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func projectListCmd() *cobra.Command {
	var lf utils.ListFlags
	filters := api.NewFilterVal()
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := projects.NewProjectsListParams()
			params.SetNamespace(cli.Namespace)
			lf.Apply(params)
			params.SetName(filters.Get("name"))
			params.SetPrivate(filters.Get("private"))
			resp, err := cli.Projects.ProjectsList(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("project_format", resp.Payload)
		},
	}
	lf.Set(cmd)
	cmd.Flags().Var(filters, "filter", "Filter results (ex. --filter name=test)")
	return cmd
}

func projectCreateCmd() *cobra.Command {
	var members []string
	body := &models.ProjectData{
		Name:          new(string),
		Private:       false,
		Collaborators: []string{},
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
			params.SetNamespace(cli.Namespace)
			params.SetProjectData(body)
			resp, err := cli.Projects.ProjectsCreate(params, cli.AuthInfo)
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
	cmd.Flags().StringSliceVar(&members, "members", []string{}, "Project members (comma separated)")
	return cmd
}

func projectDeleteCmd() *cobra.Command {
	var projectID, projectName string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete project",
		RunE: func(cmd *cobra.Command, args []string) error {
			if projectName == "" && projectID == "" {
				return errors.New("You must specify project name or id")
			}
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
			if projectID == "" {
				projectID, err = cli.GetProjectIDByName(projectName)
				if err != nil {
					return err
				}
			}
			deleteParams := projects.NewProjectsDeleteParams()
			deleteParams.SetNamespace(cli.Namespace)
			deleteParams.SetID(projectID)
			_, err = cli.Projects.ProjectsDelete(deleteParams, cli.AuthInfo)
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
			var projectID string
			var err error
			cli := api.Client()
			if len(args) < 1 {
				return errors.New("You must specify user email")
			}
			email := args[0]
			if projectName != "" {
				projectID, err = cli.GetProjectIDByName(projectName)
			} else {
				projectID, err = cli.GetProjectID()
			}
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
		params.SetNamespace(cli.Namespace)
		data := &models.CollaboratorData{
			Owner:  false,
			Member: &member,
		}
		params.SetProjectID(projectID)
		params.SetCollaboratorData(data)
		_, err := cli.Projects.ProjectsCollaboratorsCreate(params, cli.AuthInfo)
		if err != nil {
			if nerr, ok := err.(*projects.ProjectsCollaboratorsCreateBadRequest); ok {
				for _, msg := range nerr.Payload.Member {
					jww.ERROR.Print(msg)
				}
			} else {
				jww.ERROR.Printf("Error adding memeber: %s.", member)
			}
			continue
		}
		jww.FEEDBACK.Printf("Member added: %s\n", member)
	}
	return nil
}

func projectUpdateCmd() *cobra.Command {
	var projectID string
	var members []string
	updateBody := &models.ProjectData{
		Name: new(string),
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			var err error
			if *updateBody.Name == "" && projectID == "" {
				return errors.New("You must provide either project name or id")
			}
			if projectID == "" {
				projectID, err = cli.GetProjectIDByName(*updateBody.Name)
			}
			err = addMembers(projectID, members...)
			if err != nil {
				return err
			}
			params := projects.NewProjectsUpdateParams()
			params.SetNamespace(cli.Namespace)
			params.SetID(projectID)
			params.SetProjectData(updateBody)
			resp, err := cli.Projects.ProjectsUpdate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("Project updated.")
			return api.Render("project_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&projectID, "uuid", "", "Project id")
	cmd.Flags().StringVar(updateBody.Name, "name", "", "Project name")
	cmd.Flags().StringVar(&updateBody.Description, "description", "", "Project description")
	cmd.Flags().BoolVar(&updateBody.Private, "privacy", false, "Should this project be private?")
	cmd.Flags().StringSliceVar(&members, "members", []string{}, "Project members")
	return cmd
}
