package cmd

import (
	"errors"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/go-sdk/client/users"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	accountCmd := accountCmd()
	accountCmd.AddCommand(accountCreateCmd())
	accountCmd.AddCommand(accountDescribeCmd())
	accountCmd.AddCommand(accountUpdateCmd())
	accountCmd.AddCommand(accountDeleteCmd())
	RootCmd.AddCommand(accountCmd)
}

func accountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage accounts",
	}
	cmd.PersistentFlags().StringP("format", "f", "json", "User format")
	viper.BindPFlag("user_format", cmd.PersistentFlags().Lookup("format"))
	return cmd
}

func accountCreateCmd() *cobra.Command {
	accountBody := users.UsersCreateBody{
		Username: new(string),
		Password: new(string),
		Profile:  &models.UserProfile{},
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := users.NewUsersCreateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			params.SetData(accountBody)
			resp, err := cli.Users.UsersCreate(params)
			if err != nil {
				return err
			}
			return api.Render("user_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&accountBody.FirstName, "first-name", "", "New account first name")
	cmd.Flags().StringVar(&accountBody.LastName, "last-name", "", "New account last name")
	cmd.Flags().StringVar(accountBody.Username, "username", "", "New account username")
	cmd.Flags().StringVar(accountBody.Password, "password", "", "New account password")
	cmd.Flags().StringVar(&accountBody.Profile.URL, "url", "", "New account url")
	cmd.Flags().StringVar(&accountBody.Profile.AvatarURL, "avatar-url", "", "New account avatar-url")
	cmd.Flags().StringVar(&accountBody.Profile.Bio, "bio", "", "New account bio")
	cmd.Flags().StringVar(&accountBody.Profile.Location, "location", "", "New account location")
	cmd.Flags().StringVar(&accountBody.Email, "email", "", "New account email")
	cmd.Flags().StringVar(&accountBody.Profile.Company, "company", "", "New account company")
	cmd.Flags().StringVar(&accountBody.Profile.Timezone, "timezone", "", "New account timezone")
	return cmd
}

func getUserByID(userID string) (*models.User, error) {
	cli := api.Client()
	ns := viper.GetString("namespace")
	params := users.NewUsersReadParams()
	params.SetID(userID)
	params.SetNamespace(ns)
	resp, err := cli.Users.UsersRead(params)
	return resp.Payload, err
}

func getUserByName(username string) (*models.User, error) {
	cli := api.Client()
	ns := viper.GetString("namespace")
	params := users.NewUsersListParams()
	params.SetNamespace(ns)
	params.SetUsername(&username)
	resp, err := cli.Users.UsersList(params)
	if err != nil {
		return &models.User{}, err
	}
	return resp.Payload[0], nil
}

func accountDescribeCmd() *cobra.Command {
	var username, userID string
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Get information for existing account",
		RunE: func(cmd *cobra.Command, args []string) error {
			var user *models.User
			var err error
			if userID != "" {
				user, err = getUserByID(userID)
			} else {
				user, err = getUserByName(username)
			}
			if err != nil {
				return err
			}
			return api.Render("user_format", user)
		},
	}
	cmd.Flags().StringVar(&username, "username", "", "Username")
	cmd.Flags().StringVar(&userID, "uuid", "", "User id")
	return cmd
}

func accountUpdateCmd() *cobra.Command {
	var userID string
	accountBody := users.UsersPartialUpdateBody{
		Profile: &models.UserProfile{},
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := users.NewUsersPartialUpdateParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			params.SetData(accountBody)
			params.SetID(userID)
			resp, err := cli.Users.UsersPartialUpdate(params)
			if err != nil {
				return err
			}
			return api.Render("user_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&userID, "uuid", "", "User id")
	cmd.Flags().StringVar(&accountBody.FirstName, "first-name", "", "Update account first name")
	cmd.Flags().StringVar(&accountBody.LastName, "last-name", "", "Update account last name")
	cmd.Flags().StringVar(&accountBody.Username, "username", "", "Update account username")
	cmd.Flags().StringVar(&accountBody.Password, "password", "", "Update account password")
	cmd.Flags().StringVar(&accountBody.Profile.URL, "url", "", "Update account url")
	cmd.Flags().StringVar(&accountBody.Profile.AvatarURL, "avatar-url", "", "Update account avatar-url")
	cmd.Flags().StringVar(&accountBody.Profile.Bio, "bio", "", "Update account bio")
	cmd.Flags().StringVar(&accountBody.Profile.Location, "location", "", "Update account location")
	cmd.Flags().StringVar(&accountBody.Email, "email", "", "Update account email")
	cmd.Flags().StringVar(&accountBody.Profile.Company, "company", "", "Update account company")
	cmd.Flags().StringVar(&accountBody.Profile.Timezone, "timezone", "", "Update account timezone")
	return cmd
}

func accountDeleteCmd() *cobra.Command {
	var userID, name, email string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete user",
		RunE: func(cmd *cobra.Command, args []string) error {
			var user *models.User
			var err error
			cli := api.Client()
			params := users.NewUsersDeleteParams()
			ns := viper.GetString("namespace")
			params.SetNamespace(ns)
			if userID == "" {
				user, err = getUserByName(name)
				userID = user.ID
			} else {
				user, err = getUserByID(userID)
				if name != "" && *user.Username != name {
					return errors.New("Provided username is not this user username")
				}
			}
			if user.Email != email {
				return errors.New("Provided email is not this user email")
			}
			params.SetID(userID)
			_, err = cli.Users.UsersDelete(params)
			if err != nil {
				return err
			}
			jww.FEEDBACK.Println("User deleted.")
			return nil
		},
	}
	cmd.Flags().StringVar(&userID, "uuid", "", "User id")
	cmd.Flags().StringVar(&name, "username", "", "Username")
	cmd.Flags().StringVar(&email, "email", "", "User email")
	return cmd
}
