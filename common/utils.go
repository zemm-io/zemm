package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

// URLIsValidAndHTTP check if the given string is something to download
func URLIsValidAndHTTP(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}

	u, err := url.Parse(toTest)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}

	return true
}

// URLAndPathReplace replaces the file in base with new
// For Example: base: https://a.com/a.html, new: b.html = https://a.com/b.html
//				base: https://a.com/dir1/dir2/, new: b.html https://a.com/dir1/dir2/b.html
func URLAndPathReplace(base, new string) string {
	if URLIsValidAndHTTP(base) {
		// Seems to be an URL
		u, _ := url.Parse(base) // can ignore err here cause of "URLIsValid"
		if base[len(base)-1:] != "/" {
			u.Path = path.Join(path.Dir(u.Path), new)
		} else {
			u.Path = path.Join(u.Path, new)
		}

		return u.String()
	}

	// Seems to be a file path
	result := base
	if result[len(result)-1:] != "/" {
		result = path.Dir(result)
	}

	return path.Join(result, new)
}

func URLAndPathJoin(base, new string) string {
	if URLIsValidAndHTTP(base) {
		// Seems to be an URL
		u, _ := url.Parse(base) // can ignore err here cause of "URLIsValid"
		u.Path = path.Join(u.Path, new)
		return u.String()
	}

	// Seems to be a file path
	result := base
	return path.Join(result, new)
}

// DownloadURLToByte downloads a file and returns it contents as bytearray
func DownloadURLToByte(url string) ([]byte, error) {
	c := http.Client{}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to download %v, error was: %s", url, err)
	}
	req.Header.Set("User-Agent", "zemmaschaffa-go")

	res, err := c.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to download %v, error was: %s", url, err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Failed to download %v, error was: %s", url, err)
	}

	return body, nil
}

// URLToStruct reads a URL/File and parses it into the interface out
func URLToStruct(url string, out interface{}) (err error) {
	var contents []byte

	if URLIsValidAndHTTP(url) {
		contents, err = DownloadURLToByte(url)
		if err != nil {
			return err
		}
	} else {
		fp, err := os.Open(url)
		if err != nil {
			return fmt.Errorf("Failed to open %v, error was: %s", url, err)
		}
		contents, err = ioutil.ReadAll(fp)
		if err != nil {
			return fmt.Errorf("Failed to read %v, error was: %s", url, err)
		}
	}

	// Its yaml?
	if len(url) > 5 && url[len(url)-5:] == ".yaml" {
		err = yaml.Unmarshal(contents, out)
		if err != nil {
			return fmt.Errorf("Failed to decode %v, error was: %s", url, err)
		}

		return nil
	}

	err = json.Unmarshal(contents, &out)
	if err != nil {
		return fmt.Errorf("Failed to decode %v, error was: %s", url, err)
	}

	return nil
}

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
