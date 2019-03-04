package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/se7entyse7en/github-activity-tracker/pkg/client"
	"github.com/spf13/cobra"
)

var (
	user     string
	sinceStr string
	toStr    string
)

var rootCmd = &cobra.Command{
	Use:   "github-activity-tracker",
	Short: "Generates a report for a user's public activity",
	Run: func(cmd *cobra.Command, args []string) {
		c := client.NewClient(user)

		since, err := time.Parse(time.RFC3339, sinceStr)

		if err != nil {
			fmt.Println(err)
			return
		}

		to, err := time.Parse(time.RFC3339, toStr)

		if err != nil {
			fmt.Println(err)
			return
		}

		activity := c.GetActivity(context.Background(), true, &since, &to)
		fmt.Printf("%s\n", activity)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&user, "user", "u", "", "User")
	rootCmd.Flags().StringVarP(&sinceStr, "since", "s", "", "Since in RFC3339 format")
	rootCmd.Flags().StringVarP(&toStr, "to", "t", "", "To in RFC3339 format")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
