package stt

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	providers "github.com/polyfire/api/stt/providers"
)

type Silence struct {
	Start    float64
	End      float64
	Duration float64
}

/*	WARNING: This function requires silences to be sorted by ascending start
 *  timestamp. As of 2024-02-07 the RemoveSilence function should already return
 *  sorted silences. If it changes at any point in the future we would need to
 *  sort it manually or timestamps could be wrong. */
func AddSilenceToTimestamp(silences []Silence, timestamp float64) float64 {
	res := timestamp
	for _, s := range silences {
		if s.Start <= res {
			res += s.Duration
		}
	}

	return res
}

func AddSilenceToWordTimestamps(silences []Silence, words []providers.Word) []providers.Word {
	res := make([]providers.Word, 0)
	for _, w := range words {
		res = append(res, providers.Word{
			Word:              w.Word,
			PunctuatedWord:    w.PunctuatedWord,
			Start:             AddSilenceToTimestamp(silences, w.Start),
			End:               AddSilenceToTimestamp(silences, w.End),
			Confidence:        w.Confidence,
			Speaker:           w.Speaker,
			SpeakerConfidence: w.SpeakerConfidence,
		})
	}

	return res
}

func AddSilenceToDialogueTimestamps(silences []Silence, words []providers.DialogueElement) []providers.DialogueElement {
	res := make([]providers.DialogueElement, 0)
	for _, w := range words {
		res = append(res, providers.DialogueElement{
			Speaker: w.Speaker,
			Text:    w.Text,
			Start:   AddSilenceToTimestamp(silences, w.Start),
			End:     AddSilenceToTimestamp(silences, w.End),
		})
	}

	return res
}

func RemoveSilence(file io.Reader) ([]Silence, io.Reader, func(), error) {
	id := "split_transcribe-" + uuid.New().String()
	_ = os.Mkdir("/tmp/"+id, 0700)
	closeFunc := func() {
		os.RemoveAll("/tmp/" + id)
	}

	f, err := os.Create("/tmp/" + id + "/audio-file")
	if err != nil {
		return nil, nil, closeFunc, err
	}

	_, err = io.Copy(f, file)
	if err != nil {
		return nil, nil, closeFunc, err
	}

	b, err := exec.Command("bash", "-c", "ffmpeg -i \"/tmp/"+id+"/audio-file\" -af \"silencedetect=d=1\" -f null - |& tr '\\r' '\\n' | grep silence_end").
		CombinedOutput()
	if err != nil {
		return nil, nil, closeFunc, err
	}

	res := strings.Split(strings.ReplaceAll(string(b), "\r", "\n"), "\n")

	silences := make([]Silence, 0)
	vfOpts := make([]string, 0)
	afOpts := make([]string, 0)
	for _, s := range res {
		r, err := regexp.Compile(
			`^\[silencedetect.*\] silence_end: (?P<silenceend>[0-9.]+) \| silence_duration: (?P<silenceduration>[0-9.]+)$`,
		)
		if err != nil {
			return nil, nil, closeFunc, err
		}
		m := r.FindStringSubmatch(s)
		if len(m) < 2 {
			continue
		}

		result := make(map[string]string)
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				result[name] = m[i]
			}
		}

		end, err := strconv.ParseFloat(result["silenceend"], 64)
		if err != nil {
			return nil, nil, closeFunc, err
		}
		duration, err := strconv.ParseFloat(result["silenceduration"], 64)
		if err != nil {
			return nil, nil, closeFunc, err
		}
		start := end - duration
		silences = append(silences, Silence{
			Start:    start,
			Duration: duration,
			End:      end,
		})

		vfOpts = append(vfOpts, fmt.Sprintf("select='not(between(t\\,%f\\,%f))'", start, end))
		afOpts = append(afOpts, fmt.Sprintf("aselect='not(between(t\\,%f\\,%f))'", start, end))
	}

	_, err = exec.Command("bash", "-c", "ffmpeg -i \"/tmp/"+id+"/audio-file\" -vf "+strings.Join(vfOpts, ",")+" -af "+strings.Join(afOpts, ",")+" /tmp/"+id+"/audio-file-no-silence.mp3").
		CombinedOutput()
	if err != nil {
		return nil, nil, closeFunc, err
	}

	noSilenceReader, err := os.Open("/tmp/" + id + "/audio-file-no-silence.mp3")
	if err != nil {
		return nil, nil, closeFunc, err
	}

	return silences, noSilenceReader, closeFunc, nil
}
