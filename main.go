package main

import (
	"archive/zip"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

type Auth struct {
	Auth_token string
}

func getUCPAuthToken(ucp_address string, username string, password string) string {

	fmt.Println("Authenticating...")
	// curl -k option
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	loginUri := fmt.Sprint(ucp_address, "/auth/login")

	requestBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	if err != nil {
		fmt.Println(err)
	}

	response, err := http.Post(loginUri, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Request failed with error %s\n", err)
	}

	defer response.Body.Close()

	data, _ := ioutil.ReadAll(response.Body)

	var auth Auth
	jsonErr := json.Unmarshal(data, &auth)
	if jsonErr != nil {
		fmt.Println(jsonErr)
	}
	return auth.Auth_token
}

/** Download UCP Bundle **/
func getBundle(ucp_address string, authToken string, path string) {
	fmt.Println("Downlowding...")

	bundleUri := fmt.Sprint(ucp_address, "/api/clientbundle")

	client := &http.Client{}
	request, _ := http.NewRequest("GET", bundleUri, nil)
	request.Header.Set("Authorization", fmt.Sprint("Bearer ", authToken))
	response, _ := client.Do(request)

	defer response.Body.Close()

	out, err := os.Create(path)

	if err != nil {
		fmt.Println(err)
	}
	defer out.Close()
	io.Copy(out, response.Body)
}

/* Adding bundle zip file to docker context */
func AddBundleToContext(name string, path string) {
	cmd := exec.Command("docker", "context", "import", name, path)
	_, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}

	cmd = exec.Command("docker", "context", "use", name)
	_, err = cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
}

func Unzip(src string, dest string) ([]string, error) {
	fmt.Println("Extracting...")
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

/**** Main ****/
func main() {
	// Parse arguments
	address := flag.String("ucp-url", "", "MKE Url")
	username := flag.String("ucp-username", "", "MKE Username")
	password := flag.String("ucp-password", "", "MKE User password")
	flag.Parse()

	// Obtain Auth Token
	authToken := getUCPAuthToken(*address, *username, *password)

	bundlePath, bundleBasePath := func() (string, string) {
		usr, err := user.Current()
		if err != nil {
			fmt.Println(err)
		}

		u, err := url.Parse(*address)
		if err != nil {
			log.Fatal(err)
		}
		parts := strings.Split(u.Hostname(), ".")
		domain := strings.Join(parts, ".")

		bundleBasePath := filepath.Join(usr.HomeDir, ".mirantis/mke-bundle", domain, *username)
		os.MkdirAll(bundleBasePath, 0755)
		bundlePath := filepath.Join(bundleBasePath, "bundle.zip")
		fmt.Println(bundlePath)
		return bundlePath, bundleBasePath
	}()

	getBundle(*address, authToken, bundlePath)
	Unzip(bundlePath, bundleBasePath)
}
