package stt

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func SplitFile(file io.Reader) ([]io.Reader, int, func(), error) {
	id := "split_transcribe-" + uuid.New().String()
	_ = os.Mkdir("/tmp/"+id, 0700)
	closeFunc := func() {
		os.RemoveAll("/tmp/" + id)
	}
	fmt.Println(id)

	f, err := os.Create("/tmp/" + id + "/audio-file")
	if err != nil {
		return nil, 0, nil, err
	}

	_, err = io.Copy(f, file)
	if err != nil {
		return nil, 0, nil, err
	}

	err = exec.Command("ffmpeg", "-i", "/tmp/"+id+"/audio-file", "/tmp/"+id+"/audio-file.ts", "-ar", "44100").
		Run()
	if err != nil {
		return nil, 0, nil, err
	}

	ffprobeResult, err := exec.Command("ffprobe", "-i", "/tmp/"+id+"/audio-file.ts", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", "-v", "quiet", "-of", "csv=p=0").
		Output()
	if err != nil {
		return nil, 0, nil, err
	}

	durationFfprobe := strings.Split(strings.Trim(string(ffprobeResult), " \t\n"), ".")

	durationSeconds, err := strconv.Atoi(durationFfprobe[0])
	if err != nil {
		return nil, 0, nil, err
	}

	durationSeconds = durationSeconds + 1
	os.Remove("/tmp/" + id + "/audio-file")

	splitCmd := exec.Command("split", "-b", "20971400", "/tmp/"+id+"/audio-file.ts")
	splitCmd.Dir = "/tmp/" + id
	err = splitCmd.Run()
	if err != nil {
		return nil, 0, nil, err
	}

	os.Remove("/tmp/" + id + "/audio-file.ts")
	files, err := os.ReadDir("/tmp/" + id)
	if err != nil {
		fmt.Println("readdir")
		return nil, 0, nil, err
	}

	res := make([]io.Reader, 0)
	for _, file := range files {
		if file.Name() != "audio-file" && file.Name() != "audio-file.mpeg" {
			err := exec.Command("ffmpeg", "-i", "/tmp/"+id+"/"+file.Name(), "/tmp/"+id+"/"+file.Name()+".mp3").
				Run()
			if err != nil {
				return nil, 0, nil, err
			}

			os.Remove("/tmp/" + id + "/" + file.Name())
			audioPartR, err := os.Open("/tmp/" + id + "/" + file.Name() + ".mp3")
			if err != nil {
				fmt.Println("open part")
				return nil, 0, nil, err
			}
			res = append(res, audioPartR)
			fmt.Println("adding:", file.Name())
		}
	}

	return res, durationSeconds, closeFunc, nil
}
