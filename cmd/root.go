/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ma6254/bookcocoon-server/build"
	"github.com/ma6254/bookcocoon-server/config"
	"github.com/ma6254/bookcocoon-server/server"
	"github.com/spf13/cobra"
)

var (
	configFilePath string
	workDir        string
)

const asciiArt = `
================================================================================
 ____    ___    ___   _  __   ____   ___    ____   ___    ___   _   _
| __ )  / _ \  / _ \ | |/ /  / ___| / _ \  / ___| / _ \  / _ \ | \ | |
|  _ \ | | | || | | || ' /  | |    | | | || |    | | | || | | ||  \| |
| |_) || |_| || |_| || . \  | |___ | |_| || |___ | |_| || |_| || |\  |
|____/  \___/  \___/ |_|\_\  \____| \___/  \____| \___/  \___/ |_| \_| 

================================================================================
`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bookcocoon-server",
	Short: "BookCocoon Server is a backend service",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {

		log.SetFlags(log.LstdFlags | log.Llongfile)

		if workDir != "" {
			err := os.Chdir(workDir)
			if err != nil {
				return fmt.Errorf("Failed to change working directory: %v", err)
			}
		}

		log := log.New(os.Stdout, "[main]", log.LstdFlags|log.Lshortfile)
		fmt.Print(asciiArt)
		log.Printf("hello")
		log.Printf("build time: %s", build.BuildTime)
		log.Printf("version: %s", build.BuildVersion)
		log.Printf("config: %s", configFilePath)
		log.Printf("work dir: %s", workDir)

		c, err := config.NewConfigPath(configFilePath)
		if err != nil {
			return fmt.Errorf("Failed to read config file: %v", err)
		}

		s := server.NewServer(c)

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		log.Printf("Server is running...\r\n")

		err = s.Run()
		if err != nil {
			s.Stop()
			return fmt.Errorf("Failed to run server: %v", err)
		}

		exit := false

		// Wait for exit signal
		for {
			select {
			case sig := <-sigs:
				{
					s.Stop()
					log.Printf("Received signal: %v, shutting down...\r\n", sig)
					exit = true
				}

			case <-s.ExitChan:
				{
					s.Stop()
					log.Printf("Server exit signal received, shutting down...\r\n")
					exit = true
				}

			}

			if exit {
				break
			}
		}

		log.Printf("Server exiting...\r\n")
		log.Printf("Server exited\r\n")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.bookcocoon-server.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&configFilePath, "config", "c", "./config.yml", "config file (default is ./config.yml)")
	rootCmd.Flags().StringVarP(&workDir, "dir", "d", "", "work directory (default is ./)")
}
