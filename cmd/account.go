package cmd

import (
	"blink-liveview-websocket/account"
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Start a liveview stream by logging in with your Blink account and selecting a camera",
	Long: `This command will authenticate with your Blink account, fetch a list of available cameras,
and start a liveview stream from the selected camera.

Additionally, you can provide an API token, account ID, and region to bypass the login process.

Credentials can be provided via flags or environment variables:
- BLINK_TOKEN
- BLINK_ACCOUNT_ID
- BLINK_REGION

Use this command if you want to start a liveview stream, but do not have the
full connection credentials already.`,
	Run: func(cmd *cobra.Command, args []string) {
		printExports, _ := cmd.Flags().GetBool("print-exports")

		if cmd.Flag("email").Value.String() != "" {
			fmt.Print("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				os.Exit(1)
			}
			pass := string(passwordBytes)
			fmt.Println()

			account.RunWithCredentials(cmd.Flag("email").Value.String(), pass, printExports)
			return
		}

		token := getStringFlagOrEnv(cmd, "token", "BLINK_TOKEN")
		accountId := getIntFlagOrEnv(cmd, "account-id", "BLINK_ACCOUNT_ID")
		region := getStringFlagOrEnv(cmd, "region", "BLINK_REGION")

		if token == "" || accountId == 0 || region == "" {
			fmt.Println("Error: token, account-id, and region must be provided via flags or environment variables")
			os.Exit(1)
		}

		account.Run(token, accountId, region, printExports)
	},
}

func init() {
	rootCmd.AddCommand(accountCmd)

	accountCmd.Flags().StringP("email", "e", "", "Blink account email address")
	accountCmd.Flags().StringP("token", "t", "", "Blink auth token")
	accountCmd.Flags().IntP("account-id", "a", 0, "Blink account ID")
	accountCmd.Flags().StringP("region", "r", "", "Blink API region")
	accountCmd.Flags().Bool("print-exports", false, "Print shell export statements for credentials")

	accountCmd.MarkFlagsMutuallyExclusive("email", "token")
}
