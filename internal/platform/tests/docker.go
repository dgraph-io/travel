package tests

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"
)

// Container tracks information about a docker container started for tests.
type Container struct {
	ID      string
	DBHost  string
	APIHost string
}

// StartContainer runs a postgres container to execute commands.
func StartContainer(t *testing.T) *Container {
	t.Helper()

	cmd := exec.Command("docker", "run", "-d", "-P", "dgraph/standalone:v20.03.0")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container: %v", err)
	}

	id := out.String()[:12]
	t.Log("DB ContainerID:", id)

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}

	var doc []struct {
		NetworkSettings struct {
			Ports struct {
				TCP9080 []struct {
					HostIP   string `json:"HostIp"`
					HostPort string `json:"HostPort"`
				} `json:"9080/tcp"`
				TCP8080 []struct {
					HostIP   string `json:"HostIp"`
					HostPort string `json:"HostPort"`
				} `json:"8080/tcp"`
			} `json:"Ports"`
		} `json:"NetworkSettings"`
	}
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("could not decode json: %v", err)
	}

	dbNet := doc[0].NetworkSettings.Ports.TCP9080[0]
	apiNet := doc[0].NetworkSettings.Ports.TCP8080[0]

	c := Container{
		ID:      id,
		DBHost:  dbNet.HostIP + ":" + dbNet.HostPort,
		APIHost: apiNet.HostIP + ":" + apiNet.HostPort,
	}

	t.Log("DB Host[", c.DBHost, "]  API Host[", c.APIHost, "]")

	return &c
}

// StopContainer stops and removes the specified container.
func StopContainer(t *testing.T, c *Container) {
	t.Helper()

	if err := exec.Command("docker", "stop", c.ID).Run(); err != nil {
		t.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", c.ID)

	if err := exec.Command("docker", "rm", c.ID, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", c.ID)
}

// DumpContainerLogs runs "docker logs" against the container and send it to t.Log
func DumpContainerLogs(t *testing.T, c *Container) {
	t.Helper()

	out, err := exec.Command("docker", "logs", c.ID).CombinedOutput()
	if err != nil {
		t.Fatalf("could not log container: %v", err)
	}
	t.Logf("Logs for %s\n%s:", c.ID, out)
}
