package cmd

import (
	"blink-liveview-websocket/liveview"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var liveviewCmd = &cobra.Command{
	Use:   "liveview",
	Short: "Start a liveview stream directly with specified credentials",
	Long: `The liveview command exposes a direct method to start a liveview stream
without the need for logging in or utilizing the WebSocket server.

You can use this command if you already have all of the connection credentials. 
If you do not have all of the required information, use the account command instead.

Credentials can be provided via flags or environment variables:
- BLINK_REGION
- BLINK_TOKEN
- BLINK_DEVICE_TYPE
- BLINK_ACCOUNT_ID
- BLINK_NETWORK_ID
- BLINK_CAMERA_ID
- LIVEVIEW_OUTPUT (ffplay or rtsp)
- RTSP_BASE_URL (required when LIVEVIEW_OUTPUT=rtsp)`,
	Run: func(cmd *cobra.Command, args []string) {
		region := getStringFlagOrEnv(cmd, "region", "BLINK_REGION")
		token := getStringFlagOrEnv(cmd, "token", "BLINK_TOKEN")
		deviceType := getStringFlagOrEnv(cmd, "device-type", "BLINK_DEVICE_TYPE")
		accountId := getIntFlagOrEnv(cmd, "account-id", "BLINK_ACCOUNT_ID")
		networkId := getIntFlagOrEnv(cmd, "network-id", "BLINK_NETWORK_ID")
		cameraId := getIntFlagOrEnv(cmd, "camera-id", "BLINK_CAMERA_ID")
		output := getStringFlagOrEnv(cmd, "output", "LIVEVIEW_OUTPUT")
		rtspBaseURL := os.Getenv("RTSP_BASE_URL")

		if output == "" {
			output = "ffplay"
		}

		liveview.Run(region, token, deviceType, accountId, networkId, cameraId, output, rtspBaseURL)
	},
}

func getStringFlagOrEnv(cmd *cobra.Command, flagName, envName string) string {
	val := cmd.Flag(flagName).Value.String()
	if val == "" {
		val = os.Getenv(envName)
	}
	return val
}

func getIntFlagOrEnv(cmd *cobra.Command, flagName, envName string) int {
	val, _ := cmd.Flags().GetInt(flagName)
	if val == 0 {
		if envVal := os.Getenv(envName); envVal != "" {
			val, _ = strconv.Atoi(envVal)
		}
	}
	return val
}

func init() {
	rootCmd.AddCommand(liveviewCmd)

	liveviewCmd.Flags().StringP("region", "r", "", "The Blink API subdomain/region to use (e.g. u011)")
	liveviewCmd.Flags().StringP("token", "t", "", "The Blink API token to use for authentication")
	liveviewCmd.Flags().StringP("device-type", "d", "", "The Blink device type (e.g. owl, doorbell, etc)")
	liveviewCmd.Flags().IntP("account-id", "a", 0, "The Blink account ID")
	liveviewCmd.Flags().IntP("network-id", "n", 0, "The Blink network ID")
	liveviewCmd.Flags().IntP("camera-id", "c", 0, "The Blink camera ID")
	liveviewCmd.Flags().StringP("output", "o", "", "Output mode: ffplay (default) or rtsp")
}
