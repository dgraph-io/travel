package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"testing"
)

// DBContainer tracks information about the DB docker container started for tests.
type DBContainer struct {
	ID  string
	URL string
}

// startDBContainer runs a postgres container to execute commands.
func startDBContainer(t *testing.T, image string) *DBContainer {
	cmd := exec.Command("docker", "run", "-d", "-P", image)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not start container %s: %v", image, err)
	}

	id := out.String()[:12]

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("could not inspect container %s: %v", id, err)
	}

	var doc []struct {
		NetworkSettings struct {
			Ports struct {
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

	endpoint := doc[0].NetworkSettings.Ports.TCP8080[0]

	c := DBContainer{
		ID:  id,
		URL: fmt.Sprintf("http://%s", net.JoinHostPort(endpoint.HostIP, endpoint.HostPort)),
	}

	t.Logf("Image:          %s", image)
	t.Logf("DB ContainerID: %s", c.ID)
	t.Logf("URL:            %s", c.URL)
	return &c
}

// stopContainer stops and removes the specified container.
func stopContainer(t *testing.T, id string) {
	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		t.Fatalf("could not stop container: %v", err)
	}
	t.Log("Stopped:", id)

	if err := exec.Command("docker", "rm", id, "-v").Run(); err != nil {
		t.Fatalf("could not remove container: %v", err)
	}
	t.Log("Removed:", id)
}
