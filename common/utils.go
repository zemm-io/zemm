package common

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

const (
	OS_READ        = 04
	OS_WRITE       = 02
	OS_EX          = 01
	OS_USER_SHIFT  = 6
	OS_GROUP_SHIFT = 3
	OS_OTH_SHIFT   = 0

	OS_USER_R   = OS_READ << OS_USER_SHIFT
	OS_USER_RX  = OS_READ | OS_EX
	OS_USER_W   = OS_WRITE << OS_USER_SHIFT
	OS_USER_X   = OS_EX << OS_USER_SHIFT
	OS_USER_RW  = OS_USER_R | OS_USER_W
	OS_USER_RWX = OS_USER_RW | OS_USER_X

	OS_GROUP_R   = OS_READ << OS_GROUP_SHIFT
	OS_GROUP_RX  = OS_READ | OS_EX
	OS_GROUP_W   = OS_WRITE << OS_GROUP_SHIFT
	OS_GROUP_X   = OS_EX << OS_GROUP_SHIFT
	OS_GROUP_RW  = OS_GROUP_R | OS_GROUP_W
	OS_GROUP_RWX = OS_GROUP_RW | OS_GROUP_X

	OS_OTH_R   = OS_READ << OS_OTH_SHIFT
	OS_OTH_RX  = OS_READ | OS_EX
	OS_OTH_W   = OS_WRITE << OS_OTH_SHIFT
	OS_OTH_X   = OS_EX << OS_OTH_SHIFT
	OS_OTH_RW  = OS_OTH_R | OS_OTH_W
	OS_OTH_RWX = OS_OTH_RW | OS_OTH_X

	OS_ALL_R   = OS_USER_R | OS_GROUP_R | OS_OTH_R
	OS_ALL_RX  = OS_USER_RX | OS_GROUP_RX | OS_OTH_RX
	OS_ALL_W   = OS_USER_W | OS_GROUP_W | OS_OTH_W
	OS_ALL_X   = OS_USER_X | OS_GROUP_X | OS_OTH_X
	OS_ALL_RW  = OS_ALL_R | OS_ALL_W
	OS_ALL_RWX = OS_ALL_RW | OS_GROUP_X
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

// DirExists if the given directory exists and is a directory
func DirExists(directory string) bool {
	info, err := os.Stat(directory)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func CopyFile(src string, destination string, makedirs bool) error {
	if !FileExists(src) {
		return fmt.Errorf("Copy: source file '%s' doesn't exists", src)
	}

	td := path.Dir(destination)
	if makedirs == false && !DirExists(td) {
		return fmt.Errorf("Copy: destination directory '%s' doesn't exists", td)
	} else if makedirs {
		err := os.MkdirAll(td, os.ModeDir|(OS_USER_RWX|OS_GROUP_RX|OS_OTH_RX))
		if err != nil {
			return err
		}
	}

	srcFp, err := os.Open(src)
	if err != nil {
		return err
	}
	dstFp, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer dstFp.Close()

	buf := make([]byte, 1000000)
	for {
		n, err := srcFp.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := dstFp.Write(buf[:n]); err != nil {
			return err
		}
	}

	return err
}
