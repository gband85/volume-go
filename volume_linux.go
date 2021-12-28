//go:build !windows && !darwin
// +build !windows,!darwin

package volume

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var (
	useAmixer bool
)

const (
//	outputdevice string = "Master"
//	outputdevice string = "Speaker" // Modified from Master to Speaker for talkkonnect for raspberry pi
)

func init() {
	if _, err := exec.LookPath("pactl"); err != nil {
		useAmixer = true
	}
}

func cmdEnv() []string {
	return []string{"LANG=C", "LC_ALL=C"}
}

func getVolumeCmd(outputdevice string) []string {
	if useAmixer {
		return []string{"amixer", "-M", "get", outputdevice}
	}
	return []string{"pactl", "list", "sinks"}
}

var volumePattern = regexp.MustCompile(`\d+%`)

func parseVolume(out string) (int, error) {
	lines := strings.Split(out, "\t")
	for _, line := range lines {
		s := strings.TrimLeft(line, " \t")
		if useAmixer && (strings.Contains(s, "Playback") || strings.Contains(s, "Front Left:")) && strings.Contains(s, "%") || !useAmixer && strings.HasPrefix(s, "Volume:") {
			volumeStr := volumePattern.FindString(s)
			return strconv.Atoi(volumeStr[:len(volumeStr)-1])
		}
	}
	return 0, errors.New("no volume found")
}

func setVolumeCmd(volume int, outputdevice string) []string {
	if useAmixer {
		return []string{"amixer", "set", outputdevice, strconv.Itoa(volume) + "%"}
	} else if _, err := strconv.Atoi(outputdevice); err == nil {
		return []string{"pactl", "set-sink-volume", outputdevice, strconv.Itoa(volume) + "%"}
	}
	return []string{"pactl", "set-sink-volume", "0", strconv.Itoa(volume) + "%"}
}

func increaseVolumeCmd(diff int, outputdevice string) []string {

	OrigVolume, err := GetVolume(outputdevice)
	if err != nil {
		log.Println("error: Cannot Get Volume of Current Output Device")
		return nil
	}

	var Sign string
	if diff >= 0 {
		Sign = "+"
		log.Printf("debug: Increasing Volume From %v to %v\n", OrigVolume, OrigVolume+1)
	} else if useAmixer {
		diff = -diff
		Sign = "-"
		log.Printf("debug: Decreasing Volume From %v to %v\n", OrigVolume, OrigVolume-1)
	}

	if useAmixer {
		if Sign == "+" {
			return []string{"amixer", "sset", "-M", outputdevice, strconv.Itoa(OrigVolume+1) + "%"}
		}
		if Sign == "-" {
			return []string{"amixer", "sset", "-M", outputdevice, strconv.Itoa(OrigVolume-1) + "%"}
		}
	} else if _, err := strconv.Atoi(outputdevice); err == nil {
		return []string{"pactl", "--", "set-sink-volume", outputdevice, Sign + strconv.Itoa(diff) + "%"}
	}
	return []string{"pactl", "--", "set-sink-volume", "0", Sign + strconv.Itoa(diff) + "%"}
}

func getMutedCmd(outputdevice string) []string {
	if useAmixer {
		return []string{"amixer", "sget", outputdevice}
	}
	return []string{"pactl", "list", "sinks"}
}

func parseMuted(out string) (bool, error) {
	lines := strings.Split(out, "\t")
	for _, line := range lines {
		s := strings.TrimLeft(line, "\t")
		if (useAmixer && strings.Contains(s, "Playback")) && (strings.Contains(s, "on") || strings.Contains(s, "off")) {
			if strings.Contains(s, "[off]") || strings.Contains(s, "off") || strings.Contains(s, "yes") {
				return true, nil
			}
			if strings.Contains(s, "[on]") || strings.Contains(s, "on") || strings.Contains(s, "no") {
				return false, nil
			}
		}
		if !useAmixer && strings.HasPrefix(s, "Mute: ") {
			if strings.Contains(s, "[off]") || strings.Contains(s, "off") || strings.Contains(s, "yes") {
				return true, nil
			}
			if strings.Contains(s, "[on]") || strings.Contains(s, "on") || strings.Contains(s, "no") {
				return false, nil
			}
		}
	}
	return false, errors.New("no muted information found")
}

func muteCmd(outputdevice string) []string {
	if useAmixer {
		return []string{"amixer", "set", outputdevice, "mute"}
	} else if _, err := strconv.Atoi(outputdevice); err == nil {
		return []string{"pactl", "set-sink-mute", outputdevice, "1"}
	}
	return []string{"pactl", "set-sink-mute", "0", "1"}
}

func unmuteCmd(outputdevice string) []string {
	if useAmixer {
		return []string{"amixer", "set", outputdevice, "unmute"}
	} else if _, err := strconv.Atoi(outputdevice); err == nil {
		return []string{"pactl", "set-sink-mute", outputdevice, "0"}
	}
	return []string{"pactl", "set-sink-mute", "0", "0"}
}
