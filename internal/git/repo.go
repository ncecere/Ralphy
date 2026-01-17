package git

import (
	"errors"
	"strings"
)

func ParseRepoFromRemote() (owner, repo string, err error) {
	url, err := RemoteURL()
	if err != nil {
		return "", "", err
	}
	return ParseRepoFromURL(url)
}

func ParseRepoFromURL(url string) (owner, repo string, err error) {
	url = strings.TrimSuffix(url, ".git")

	if strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return "", "", errors.New("invalid git URL")
		}
		path := parts[1]
		return splitOwnerRepo(path)
	}

	if strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://") {
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")
		parts := strings.SplitN(url, "/", 2)
		if len(parts) != 2 {
			return "", "", errors.New("invalid https URL")
		}
		return splitOwnerRepo(parts[1])
	}

	return "", "", errors.New("unsupported URL format")
}

func splitOwnerRepo(path string) (string, string, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", errors.New("cannot parse owner/repo")
	}
	return parts[0], parts[1], nil
}
