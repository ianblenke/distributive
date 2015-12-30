// dockerstatus provides a few functions for getting very simple data out
// of Docker, mostly for use in simple status checks.
package dockerstatus

import (
	"fmt"
	"github.com/CiscoCloud/distributive/tabular"
	log "github.com/Sirupsen/logrus"
	"os/exec"
	"strings"
)

// DockerImageRepositories returns a slice of the names of the Docker images
// present on the host (what's under the REPOSITORIES column of `docker images`)
func DockerImageRepositories() (images []string, err error) {
	cmd := exec.Command("docker", "images")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// try escalating to sudo, the error might have been one of permissions
		cmd = exec.Command("sudo", "docker", "images")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return images, err
		}
	}
	table := tabular.ProbabalisticSplit(string(out))
	return tabular.GetColumnByHeader("REPOSITORIES", table), nil
}

func parseRunningContainers(output string) (containers []string) {
	// the output of `docker ps -a` has spaces in columns, but each column
	// is separated by 2 or more spaces. Just what Probabalistic was made for!
	lines := tabular.ProbabalisticSplit(string(output))
	log.Debugf("Docker containers table: %s", tabular.ToString(lines))
	names := tabular.GetColumnByHeader("IMAGE", lines)
	statuses := tabular.GetColumnByHeader("STATUS", lines)
	for i, status := range statuses {
		// index error caught by second condition in if clause
		if strings.Contains(status, "Up") && len(names) > i {
			containers = append(containers, names[i])
		}
	}
	return containers
}

// RunningContainers returns a list of names of running docker containers
// (what's under the IMAGE column of `docker ps -a` if it has status "Up".
func RunningContainers() (containers []string, err error) {
	cmd := exec.Command("docker", "ps", "-a")
	out, err := cmd.CombinedOutput()
	if err != nil {
		cmd = exec.Command("sudo", "docker", "ps", "-a")
		out, err = cmd.CombinedOutput()
		if err != nil {
			return containers, err
		}
	} else if out == nil {
		err = fmt.Errorf("The command %v produced no output", cmd.Args)
		return containers, err
	}
	return parseRunningContainers(string(out)), nil
}
