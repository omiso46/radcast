package main

import (
	"fmt"
	"os/exec"
	"regexp"
)

func lookConverterCommand() (string, error) {
	for _, p := range []string{"ffmpeg", "avconv"} {
		cmd, err := exec.LookPath(p)
		if err == nil {
			return cmd, nil
		}
	}
	return "", fmt.Errorf("not found converter cmd such also ffmpeg, avconv.")
}

func newConverterCmd(path, output string) (*exec.Cmd, error) {

	switch {
	case regexp.MustCompile("ffmpeg$").MatchString(path):
		return newFfmpegCmd(path, output), nil
	case regexp.MustCompile("avconv$").MatchString(path):
		return newAvconvCmd(path, output), nil
	}

	return nil, fmt.Errorf("path should be ffmpeg or avconv")
}

func newFfmpegCmd(ffmpeg, output string) *exec.Cmd {
	return exec.Command(
		ffmpeg,
		"-y",
		"-i", "-",
		"-vn",
		"-acodec", "copy",
		output,
	)
}

func newAvconvCmd(avconv, output string) *exec.Cmd {
	return exec.Command(
		avconv,
		"-y",
		"-i", "-",
		"-vn",
		"-c:a", "copy",
		output,
	)
}
