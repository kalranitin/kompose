package integration

import (
	"bytes"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/docker"
	dockerclient "github.com/fsouza/go-dockerclient"

	. "gopkg.in/check.v1"
)

const (
	SimpleTemplate = `
        hello:
          image: busybox
          stdin_open: true
          tty: true
        `
	SimpleTemplateWithVols = `
        hello:
          image: busybox
          stdin_open: true
          tty: true
          volumes:
          - /root:/root
          - /home:/home
          - /var/lib/vol1
          - /var/lib/vol2
          - /var/lib/vol4
        `

	SimpleTemplateWithVols2 = `
        hello:
          image: busybox
          stdin_open: true
          tty: true
          volumes:
          - /tmp/tmp-root:/root
          - /var/lib/vol1
          - /var/lib/vol3
          - /var/lib/vol4
        `
)

func Test(t *testing.T) { TestingT(t) }

func asMap(items []string) map[string]bool {
	result := map[string]bool{}
	for _, item := range items {
		result[item] = true
	}
	return result
}

var random = rand.New(rand.NewSource(time.Now().Unix()))

func RandStr(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[random.Intn(len(letters))]
	}
	return string(b)
}

type RunSuite struct {
	command  string
	projects []string
}

func (s *RunSuite) TearDownTest(c *C) {
	// Delete all containers
	client := GetClient(c)
	containers, err := client.ListContainers(dockerclient.ListContainersOptions{All: true})
	c.Assert(err, IsNil)
	for _, container := range containers {
		// Unpause container (if paused) and ignore error (if wasn't paused)
		client.UnpauseContainer(container.ID)
		// And remove force \o/
		err := client.RemoveContainer(dockerclient.RemoveContainerOptions{ID: container.ID, Force: true, RemoveVolumes: true})
		c.Assert(err, IsNil)
	}
}

var _ = Suite(&RunSuite{
	command: "../bundles/kompose_linux-amd64",
})

func (s *RunSuite) CreateProjectFromText(c *C, input string) string {
	return s.ProjectFromText(c, "create", input)
}

func (s *RunSuite) RandomProject() string {
	return "test-project-" + RandStr(7)
}

func (s *RunSuite) ProjectFromText(c *C, command, input string) string {
	projectName := s.RandomProject()
	return s.FromText(c, projectName, command, input)
}

func (s *RunSuite) FromText(c *C, projectName, command string, argsAndInput ...string) string {
	command, args, input := s.createCommand(c, projectName, command, argsAndInput)

	cmd := exec.Command(s.command, args...)
	cmd.Stdin = bytes.NewBufferString(strings.Replace(input, "\t", "  ", -1))
	if os.Getenv("TESTVERBOSE") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}

	err := cmd.Run()
	if err != nil {
		logrus.Errorf("Failed to run %s %v: %v\n with input:\n%s", s.command, err, args, input)
	}

	c.Assert(err, IsNil)

	return projectName
}

// Doesn't assert that command runs successfully
func (s *RunSuite) FromTextCaptureOutput(c *C, projectName, command string, argsAndInput ...string) (string, string) {
	command, args, input := s.createCommand(c, projectName, command, argsAndInput)

	cmd := exec.Command(s.command, args...)
	cmd.Stdin = bytes.NewBufferString(strings.Replace(input, "\t", "  ", -1))

	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Errorf("Failed to run %s %v: %v\n with input:\n%s", s.command, err, args, input)
	}

	return projectName, string(output[:])
}

func (s *RunSuite) createCommand(c *C, projectName, command string, argsAndInput []string) (string, []string, string) {
	args := []string{"--verbose", "-p", projectName, "-f", "-", command}
	args = append(args, argsAndInput[0:len(argsAndInput)-1]...)

	input := argsAndInput[len(argsAndInput)-1]

	if command == "up" {
		args = append(args, "-d")
	} else if command == "down" {
		args = append(args, "--timeout", "0")
	} else if command == "restart" {
		args = append(args, "--timeout", "0")
	} else if command == "stop" {
		args = append(args, "--timeout", "0")
	}

	logrus.Infof("Running %s %v", command, args)

	return command, args, input
}

func GetClient(c *C) *dockerclient.Client {
	client, err := docker.CreateClient(docker.ClientOpts{})

	c.Assert(err, IsNil)

	return client
}

func (s *RunSuite) GetContainerByName(c *C, name string) *dockerclient.Container {
	client := GetClient(c)
	container, err := docker.GetContainerByName(client, name)

	c.Assert(err, IsNil)

	if container == nil {
		return nil
	}

	info, err := client.InspectContainer(container.ID)

	c.Assert(err, IsNil)

	return info
}

func (s *RunSuite) GetContainersByProject(c *C, project string) []dockerclient.APIContainers {
	client := GetClient(c)
	containers, err := docker.GetContainersByFilter(client, docker.PROJECT.Eq(project))

	c.Assert(err, IsNil)

	return containers
}
