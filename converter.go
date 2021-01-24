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
		"-f", "concat",
		"-safe", "0",
		"-y",
		"-c", "copy",
		"-movflags",
		"+faststart",
		output,
	)
	// return exec.Command(
	// 	ffmpeg,
	// 	"-y",
	// 	"-i", "-",
	// 	"-vn",
	// 	"-acodec", "copy",
	// 	"-movflags",
	// 	"+faststart",
	// 	output,
	// )
	//	return exec.Command(
	//		ffmpeg,
	//		"-y",
	//		"-i", "-",
	//		"-vn",
	//		"-acodec", "libmp3lame",
	//		"-ar", "44100",
	//		"-ac", "2",
	//		output,
	//	)

}

func newAvconvCmd(avconv, output string) *exec.Cmd {
	return exec.Command(
		avconv,
		"-y",
		"-i", "-",
		"-vn",
		"-c:a", "copy",
		"-movflags",
		"+faststart",
		output,
	)
	//	return exec.Command(
	//		avconv,
	//		"-y",
	//		"-i", "-",
	//		"-vn",
	//		"-c:a", "libmp3lame",
	//		"-ar", "44100",
	//		"-ac", "2",
	//		output,
	//	)
}

func hlsFfmpegCmd(ffmpeg, streamURL string, authtoken string, sec string, output string) *exec.Cmd {
	return exec.Command(
		ffmpeg,
		"-loglevel", "error",
		"-fflags", "+discardcorrupt",
		"-headers", `"X-Radiko-Authtoken: `+authtoken+`"`,
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
