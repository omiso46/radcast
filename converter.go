package main

import (
	"fmt"
	"os/exec"
)

func lookConverterCommand() (string, error) {
	cmd, err := exec.LookPath("ffmpeg")
	if err == nil {
		return cmd, nil
	}
	return "", fmt.Errorf("not found converter cmd, %s", "ffmpeg.")
}

func hlsFfmpegCmd(ffmpeg, streamURL string, authtoken string, sec string, output string) *exec.Cmd {
	return exec.Command(
		ffmpeg,
		"-loglevel", "error",
		"-fflags", "+discardcorrupt",
		"-headers", "X-Radiko-Authtoken:"+authtoken,
		"-i", streamURL,
		"-acodec", "copy",
		"-vn",
		"-bsf:a", "aac_adtstoasc",
		"-y",
		"-t", sec,
		"-movflags", "+faststart",
		output,
	)

}
