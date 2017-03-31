package cmd

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gobuffalo/replo/repl"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var session = repl.NewSession(`fmt`)

var RootCmd = &cobra.Command{
	Use:   "replo",
	Short: "A brief description of your application",
	RunE: func(cmd *cobra.Command, args []string) error {
		return session.Start()
	},
}

func historyFilePath() string {
	pwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	home, err := homedir.Dir()
	if err != nil {
		return ""
	}

	h := sha1.New()
	h.Write([]byte(pwd))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	return filepath.Join(home, "."+sha1_hash+".replo")
}

func init() {
	RootCmd.Flags().BoolVarP(&session.Debug, "debug", "d", false, "enable debugging")
	RootCmd.Flags().BoolVarP(&session.SkipHistory, "skip-history", "s", false, "skip loading the previous history file")
	RootCmd.Flags().StringVarP(&session.History, "file", "f", historyFilePath(), "file to read/write the session to")
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
