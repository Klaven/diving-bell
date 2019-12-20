package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
	"k8s.io/klog"

	cluster "github.com/SUSE/skuba/pkg/skuba/actions/cluster/init"
)

type Node struct {
	User     string
	Target   string
	HostName string
}
type ClusterConfig struct {
	ClusterName        string
	ControlPlaneTarget string
	Managers           []Node
	Workers            []Node
}

func initCluster(clusterName string, controlPlaneTarget string) {
	// Get current user
	usr, err := user.Current()
	if err != nil {
		klog.Fatalf("getting current user failed: %s", err)
	}

	// Init the cluster
	initConfig, err := cluster.NewInitConfiguration(
		fmt.Sprintf("%s/%s", usr.HomeDir, clusterName),
		"",
		controlPlaneTarget,
		"",
		false)
	if err != nil {
		klog.Fatalf("init failed due to error: %s", err)
	}

	if err = cluster.Init(initConfig); err != nil {
		klog.Fatalf("init failed due to error: %s", err)
	}
}

func runShell(shellCmd string) {
	args := strings.Fields(shellCmd)
	cmd := exec.Command(args[0], args[1:len(args)]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		klog.Fatal(err)
	}
	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		klog.Infoln(scanner.Text())
	}
}

func bootstrapControlPlane(firstMaster Node) {
	klog.Infof("Bootstrapping %+v\n", firstMaster)
	cmd := fmt.Sprintf("skuba node bootstrap --user %s --sudo --target %s %s", firstMaster.User, firstMaster.Target, firstMaster.HostName)
	runShell(cmd)
}

func joinNodes(nodes []Node) {

	for _, node := range nodes {
		klog.Infof("Joining %+v\n", node)
		cmd := fmt.Sprintf("skuba node join --user %s --sudo --target %s %s", node.User, node.Target, node.HostName)
		runShell(cmd)
		time.Sleep(10 * time.Second)
	}
}

func main() {
	var clusterConfig ClusterConfig
	reader, _ := os.Open("cluster-config.yaml")
	buf, _ := ioutil.ReadAll(reader)
	yaml.Unmarshal(buf, &clusterConfig)

	initCluster(clusterConfig.ClusterName, clusterConfig.ControlPlaneTarget)
	bootstrapControlPlane(clusterConfig.Managers[0])

	if len(clusterConfig.Managers) > 1 {
		joinNodes(clusterConfig.Managers[1:len(clusterConfig.Managers)])
	}

	joinNodes(clusterConfig.Workers)
}
