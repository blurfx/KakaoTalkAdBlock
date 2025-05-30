package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func CheckLatestVersion(currentVersion string) (string, bool) {
	response, err := http.Get("https://api.github.com/repos/blurfx/KakaoTalkAdBlock/releases/latest")
	if err != nil {
		return currentVersion, false
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return currentVersion, false
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return currentVersion, false
	}

	tagName := data["tag_name"].(string)

	if hasNewRelease(currentVersion, tagName) {
		return tagName, true
	}

	return currentVersion, false
}

func hasNewRelease(current, latest string) bool {
	v1Parts := strings.Split(current, ".")
	v2Parts := strings.Split(latest, ".")

	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		v1Part, err := strconv.Atoi(v1Parts[i])
		if err != nil {
			return false
		}
		v2Part, err := strconv.Atoi(v2Parts[i])
		if err != nil {
			return false
		}

		if v1Part > v2Part {
			return false
		} else if v1Part < v2Part {
			return true
		}
	}
	return false
}
