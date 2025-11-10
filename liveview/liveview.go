package liveview

import (
	"blink-liveview-websocket/common"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

func Run(region string, token string, deviceType string, accountId int, networkId int, cameraId int, output string, rtspBaseURL string) {
	var cmd *exec.Cmd
	var inputPipe io.WriteCloser

	if output == "rtsp" {
		// RTSP mode: use ffmpeg to publish to RTSP server
		if rtspBaseURL == "" {
			log.Println("RTSP_BASE_URL must be set when using RTSP output mode")
			os.Exit(1)
		}
		rtspURL := fmt.Sprintf("%s/blink-%d", rtspBaseURL, cameraId)
		log.Printf("Publishing to RTSP URL: %s\n", rtspURL)
		
		cmd = exec.Command("ffmpeg",
			"-f", "mpegts",
			"-i", "pipe:0",
			"-c", "copy",
			"-f", "rtsp",
			rtspURL,
		)
		var err error
		inputPipe, err = cmd.StdinPipe()
		if err != nil {
			log.Println("error creating ffmpeg stdin pipe", err)
			os.Exit(1)
		}
	} else {
		// Default ffplay mode
		cmd = exec.Command("ffplay",
			"-f", "mpegts",
			"-err_detect", "ignore_err",
			"-window_title", "Blink Liveview Middleware",
			"-",
		)
		var err error
		inputPipe, err = cmd.StdinPipe()
		if err != nil {
			log.Println("error creating ffplay stdin pipe", err)
			os.Exit(1)
		}
	}

	if err := cmd.Start(); err != nil {
		log.Printf("error starting %s: %v\n", cmd.Path, err)
		os.Exit(1)
	}
	defer cmd.Process.Kill()

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
		DeviceType: deviceType,
		AccountId:  accountId,
		NetworkId:  networkId,
		CameraId:   cameraId,
	}
	if err := common.Livestream(ctx, accountDetails, inputPipe); err != nil {
		log.Println("error during livestream", err)
	}

	inputPipe.Close()
	if err := cmd.Wait(); err != nil {
		log.Printf("error waiting for command: %v\n", err)
	}
}
