package cmd

import (
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "tbs",
	Short: "3Blades CLI",
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

func init() {
	jww.SetLogThreshold(jww.LevelTrace)
	jww.SetStdoutThreshold(jww.LevelInfo)
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.threeblades.yaml)")
	RootCmd.PersistentFlags().String("namespace", "", "3Blades namespace")
	viper.BindPFlag("namespace", RootCmd.PersistentFlags().Lookup("namespace"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.AddConfigPath("$HOME")
	viper.SetConfigName(".threeblades") // name of config file (without extension)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("THREEBLADES")
	viper.BindEnv("project")
	viper.BindEnv("namespace")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		jww.ERROR.Printf("Error reading config file: %s\n", err)
	}
	token, err := ioutil.ReadFile(tokenFilePath())
	if err == nil {
		viper.Set("token", token)
	}
}
