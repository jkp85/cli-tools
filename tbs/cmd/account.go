package cmd

import (
	"errors"
	"fmt"

	"github.com/3Blades/cli-tools/tbs/api"
	"github.com/3Blades/go-sdk/client/users"
	"github.com/3Blades/go-sdk/models"
	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func init() {
	cmd := accountCmd()
	cmd.AddCommand(
		accountCreateCmd(),
		accountDescribeCmd(),
		accountUpdateCmd(),
		accountDeleteCmd(),
	)
	RootCmd.AddCommand(cmd)
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
	accountBody := &models.UserData{
		Username: new(string),
		Password: new(string),
		Profile:  &models.UserProfile{},
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create account",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			msg := "You need to provide flags: "
			raise := false
			if *accountBody.Username == "" {
				msg += "username"
				raise = true
			}
			if *accountBody.Password == "" {
				msg += ", password"
				raise = true
			}
			if accountBody.Email == "" {
				msg += ", email."
				raise = true
			}
			if raise {
				return errors.New(msg)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := users.NewUsersCreateParams()
			params.SetUserData(accountBody)
			resp, err := cli.Users.UsersCreate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("user_format", resp.Payload)
		},
	}
	flags := cmd.Flags()
	flags.StringVar(&accountBody.FirstName, "first-name", "", "New account first name")
	flags.StringVar(&accountBody.LastName, "last-name", "", "New account last name")
	flags.StringVar(accountBody.Username, "username", "", "New account username (required)")
	flags.StringVar(accountBody.Password, "password", "", "New account password (required)")
	flags.StringVar(&accountBody.Profile.URL, "url", "", "New account url")
	flags.StringVar(&accountBody.Profile.AvatarURL, "avatar-url", "", "New account avatar-url")
	flags.StringVar(&accountBody.Profile.Bio, "bio", "", "New account bio")
	flags.StringVar(&accountBody.Profile.Location, "location", "", "New account location")
	flags.StringVar(&accountBody.Email, "email", "", "New account email (required)")
	flags.StringVar(&accountBody.Profile.Company, "company", "", "New account company")
	flags.StringVar(&accountBody.Profile.Timezone, "timezone", "", "New account timezone")
	return cmd
}

func getUserByID(userID string) (*models.User, error) {
	cli := api.Client()
	params := users.NewUsersReadParams()
	params.SetID(userID)
	resp, err := cli.Users.UsersRead(params, cli.AuthInfo)
	return resp.Payload, err
}

func getUserByName(username string) (*models.User, error) {
	cli := api.Client()
	params := users.NewUsersListParams()
	params.SetUsername(&username)
	resp, err := cli.Users.UsersList(params, cli.AuthInfo)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) > 0 {
		return resp.Payload[0], nil
	}
	return nil, fmt.Errorf("There is no user with username: %s", username)
}

func getUserByEmail(email string) (*models.User, error) {
	cli := api.Client()
	params := users.NewUsersListParams()
	params.SetEmail(&email)
	resp, err := cli.Users.UsersList(params, cli.AuthInfo)
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) > 0 {
		return resp.Payload[0], nil
	}
	return nil, fmt.Errorf("There is no user with email: %s", email)
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
	accountBody := &models.UserData{
		Username: new(string),
		Password: new(string),
		Profile:  &models.UserProfile{},
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := api.Client()
			params := users.NewUsersUpdateParams()
			params.SetUserData(accountBody)
			params.SetID(userID)
			resp, err := cli.Users.UsersUpdate(params, cli.AuthInfo)
			if err != nil {
				return err
			}
			return api.Render("user_format", resp.Payload)
		},
	}
	cmd.Flags().StringVar(&userID, "uuid", "", "User id")
	cmd.Flags().StringVar(&accountBody.FirstName, "first-name", "", "Update account first name")
	cmd.Flags().StringVar(&accountBody.LastName, "last-name", "", "Update account last name")
	cmd.Flags().StringVar(accountBody.Username, "username", "", "Update account username")
	cmd.Flags().StringVar(accountBody.Password, "password", "", "Update account password")
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
			if name != "" {
				user, err = getUserByName(name)
			} else if email != "" {
				user, err = getUserByEmail(email)
			} else {
				user, err = getUserByID(userID)
			}
			if err != nil {
				return err
			}
			params.SetID(user.ID)
			_, err = cli.Users.UsersDelete(params, cli.AuthInfo)
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
