package account

import (
	"blink-liveview-websocket/common"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Fetches the list of Blink devices and prints them to the console
// Select one of the devices to start a liveview stream
//
// token: the Blink API token
//
// accountId: the Blink account ID
//
// region: the Blink API region
//
// printExports: if true, print shell export statements instead of starting liveview
func Run(token string, accountId int, region string, printExports bool) {
	if printExports {
		fmt.Printf("export BLINK_TOKEN='%s'\n", token)
		fmt.Printf("export BLINK_ACCOUNT_ID='%d'\n", accountId)
		fmt.Printf("export BLINK_REGION='%s'\n", region)
		return
	}

	baseUrl := common.GetApiUrl(region)
	homescreenUrl := fmt.Sprintf("%s/api/v4/accounts/%d/homescreen", baseUrl, accountId)
	devices, err := common.Homescreen(homescreenUrl, token)
	if err != nil {
		log.Println("error getting homescreen", err)
		os.Exit(1)
	}

	fmt.Println("Select a device to start a liveview stream:")
	output, options := common.PrintDeviceOptions(devices)
	if len(options) == 0 {
		log.Println("no devices found")
		os.Exit(1)
	} else {
		fmt.Println(output)
	}

getDevice:
	fmt.Print("Device number: ")
	var deviceNumber int
	if _, err = fmt.Scanln(&deviceNumber); err != nil {
		log.Println("error reading device number", err)
		os.Exit(1)
	}
	fmt.Println()

	if deviceNumber < 1 || deviceNumber > len(options) {
		log.Println("invalid device number")
		goto getDevice
	}

	device := options[deviceNumber-1]
	log.Printf("Selected device: %s\n", device.FormattedName)

	ffplayCmd := exec.Command("ffplay",
		"-f", "mpegts",
		"-err_detect", "ignore_err",
		"-window_title", "Blink Liveview Middleware",
		"-",
	)
	inputPipe, err := ffplayCmd.StdinPipe()
	if err != nil {
		log.Println("error creating ffplay stdin pipe", err)
	}

	if err := ffplayCmd.Start(); err != nil {
		log.Println("error starting ffplay", err)
	}
	defer ffplayCmd.Process.Kill()

	ctx, cancelCtx := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log.Println("Received SIGINT")
		cancelCtx()
	}()

	accountDetails := common.AccountDetails{
		Region:     region,
		Token:      token,
		DeviceType: device.DeviceType,
		AccountId:  accountId,
		NetworkId:  device.NetworkId,
		CameraId:   device.DeviceId,
	}
	if err := common.Livestream(ctx, accountDetails, inputPipe); err != nil {
		log.Println("error starting liveview session", err)
	}

	inputPipe.Close()
	if err := ffplayCmd.Wait(); err != nil {
		log.Println("error waiting for ffplay", err)
	}
}

// Authenticates with the Blink API using the provided email and password
// and fetches the list of Blink devices
// Select one of the devices to start a liveview stream
//
// email: the Blink account email address
//
// password: the Blink account password
//
// printExports: if true, print shell export statements after login instead of starting liveview
func RunWithCredentials(email string, password string, printExports bool) {
	fingerprint, err := common.GetFingerprint("")
	if err != nil {
		log.Println("error getting fingerprint", err)
		os.Exit(1)
	}

	loginResp, err := common.Login(email, password, "", fingerprint)
	if err != nil {
		log.Println("error logging in", err)
		os.Exit(1)
	}

	var tsvResp *common.LoginResponse
	if loginResp.TwoStepVerification == "sms" {
		log.Println("Client verification is required. A SMS code has been sent to your phone.")
		fmt.Print("Code: ")
		codeBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			os.Exit(1)
		}
		code := string(codeBytes)
		fmt.Println()

		var tsvErr error
		if tsvResp, tsvErr = common.Login(email, password, code, fingerprint); tsvErr != nil {
			log.Println("error verifying pin", tsvErr)
			os.Exit(1)
		}
	} else {
		// TODO: Handle other TSV states or skip TSV if not required
		log.Println("Unexpected two-step verification state:", loginResp.TwoStepVerification)
		os.Exit(1)
	}

	tierInfo, err := common.GetTierInfo(tsvResp.AccessToken)
	if err != nil {
		log.Println("error getting tier info", err)
		os.Exit(1)
	}

	if err := fingerprint.Store(); err != nil {
		log.Println("error saving the fingerprint. Next login will require a new SMS code.", err)
	}

	log.Printf("Logged in successfully.\n\tToken: %s,\n\tAccountId: %d,\n\tRegion: %s\n", tsvResp.AccessToken, tierInfo.AccountId, tierInfo.Tier)

	if printExports {
		fmt.Printf("export BLINK_TOKEN='%s'\n", tsvResp.AccessToken)
		fmt.Printf("export BLINK_ACCOUNT_ID='%d'\n", tierInfo.AccountId)
		fmt.Printf("export BLINK_REGION='%s'\n", tierInfo.Tier)
		return
	}

	Run(tsvResp.AccessToken, tierInfo.AccountId, tierInfo.Tier, false)
}
