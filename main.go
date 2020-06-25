package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

type Auth struct {
	Auth_token string
}

func getUCPAuthToken(ucp_address string, username string, password string) string {

	// curl -k option
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	loginUri := fmt.Sprint("https://", ucp_address, "/auth/login")

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
func downloadBundle(ucp_address string, authToken string, path string) {
	fmt.Println("Downlowding client bundle")
	bundleUri := fmt.Sprint("https://", ucp_address, "/api/clientbundle")

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

/* Read user input from stdin */
func readInput() (string, string, string, string) {
	reader := bufio.NewReader(os.Stdin)
	// UCP Address
	fmt.Print("UCP Address: ")
	ucp_address, _ := reader.ReadString('\n')

	// UCP Username
	fmt.Print("ucp username: ")
	username, _ := reader.ReadString('\n')

	// UCP Password
	fmt.Print("ucp password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println(err)
	}
	password := string(bytePassword)

	fmt.Println("")
	fmt.Print("Context Name: ")
	contextName, _ := reader.ReadString('\n')
	return strings.TrimSpace(ucp_address), strings.TrimSpace(username), strings.TrimSpace(password), strings.TrimSpace(contextName)
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

/**** Main ****/
func main() {
	// Read Input
	ucp_address, username, password, contextName := readInput()
	// Obtain Auth Token
	authToken := getUCPAuthToken(strings.TrimSpace(ucp_address), username, password)
	// Download bundle
	// From
	// To
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
  }
  bundleBasePath := filepath.Join(usr.HomeDir, ".ucp-bundle", ucp_address)
  os.MkdirAll(bundleBasePath, 0755)
	bundlePath := filepath.Join(bundleBasePath, fmt.Sprint(contextName, "-", username, ".zip"))
	downloadBundle(ucp_address, authToken, bundlePath)

	// add bundle to docker context
	fmt.Println("Adding bundle to docker context")
	AddBundleToContext(contextName, bundlePath)
}
