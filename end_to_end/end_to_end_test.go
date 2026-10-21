package end_to_end

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

var apiHost string

func warn(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func isPortOpen(host string, port string) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func runServer(port string) *exec.Cmd {
	here, err := os.Getwd()
	warn(err)
	fmt.Println("Current working directory:", here)
	serverPath, err := filepath.Abs(filepath.Join(here, "../cmd/app/bin/pb-go-api-starter"))
	warn(err)
	cmd := exec.Command(serverPath)
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error starting the server:", err)
		os.Exit(1)
	}
	retries := 0
	for retries < 10 && !isPortOpen("127.0.0.1", port) {
		fmt.Println("waiting for server to start")
		time.Sleep(2 * time.Second)
	}
	if !isPortOpen("127.0.0.1", port) {
		fmt.Println("Error starting the server:", err)
	}
	return cmd
}

func TestMain(m *testing.M) {
	var cmd *exec.Cmd
	needsLocalServer := false
	host := os.Getenv("API_HOST")
	if host == "" {
		host = "127.0.0.1"
		needsLocalServer = true
	}
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8000"
	}
	if needsLocalServer {
		cmd = runServer(port)
	}
	apiHost = "http://" + host + ":" + port
	fmt.Printf("Running e2e tests against %s\n", apiHost)

	exitCode := m.Run()
	if cmd != nil {
		_ = cmd.Process.Kill()
	}
	os.Exit(exitCode)
}

func postRequest(url string, payload any) (map[string]interface{}, int, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	var response map[string]interface{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, 0, err
	}

	return response, resp.StatusCode, nil
}
