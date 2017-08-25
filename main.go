package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli"
)

var endpoint string

type Config struct {
	BrokerUrl string `json:"brokerUrl"`
}

type ClustersAPIResponse struct {
	Status       int       `json:"status"`
	ErrorMessage string    `json:"errorMessage"`
	Data         []Cluster `json:"data"`
}

type NewTaskRequest struct {
	SourceClusterId string `json:"sourceClusterId"`
	TargetClusterId string `json:"targetClusterId"`
}

type NewTaskAPIResponse struct {
	Status string `json:"status"`
	Error  string `json:"errorMessage"`
}

type TasksAPIResponse struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	Data         []Task `json:"data"`
}

type TaskAPIResponse struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"errorMessage"`
	Data         Task   `json:"data"`
}

type ClusterAPIResponse struct {
	Status       int     `json:"status"`
	ErrorMessage string  `json:"errorMessage"`
	Data         Cluster `json:"data"`
}

type Task struct {
	Id            string `json:"id"`
	Status        string `json:"status"`
	SourceCluster string `json:"sourceClusterId"`
	TargetCluster string `json:"targetClusterId"`
}

type Cluster struct {
	Name                   string `json:"id"`
	Version                string `json:"kubeletVersion"`
	ConfigMaps             int    `json:"configMapsCount"`
	Deployments            int    `json:"depsCount"`
	Pods                   int    `json:"podCount"`
	ReplicationControllers int    `json:"rcCount"`
	ReplicaSets            int    `json:"rsCount"`
	Services               int    `json:"svcCount"`
}

func main() {
	setEndpoint()

	app := cli.NewApp()
	app.Name = "kaptaind cli"
	app.EnableBashCompletion = true
	app.Version = "0.1"
	app.Usage = "controls the kaptaind broker api"

	app.Commands = []cli.Command{
		{
			Name:  "get",
			Usage: "get kaptaind resources",
			Subcommands: []cli.Command{
				{
					Name:  "clusters",
					Usage: "get Kubernetes clusters",
					Action: func(c *cli.Context) error {
						getClusters()
						return nil
					},
				},
				{
					Name:  "tasks",
					Usage: "get tasks",
					Action: func(c *cli.Context) error {
						getTasks()
						return nil
					},
				},
				{
					Name:  "cluster",
					Usage: "get cluster information",
					Action: func(c *cli.Context) error {
						id := c.Args().First()
						if id == "" {
							fmt.Println("missing cluster id. try get cluster <id>")
						} else {
							getCluster(id)
						}

						return nil
					},
				},
				{
					Name:  "task",
					Usage: "get task information",
					Action: func(c *cli.Context) error {
						id := c.Args().First()
						if id == "" {
							fmt.Println("missing task id. try get task <id>")
						} else {
							getTask(id)
						}

						return nil
					},
				},
			},
		},
		{
			Name:  "delete",
			Usage: "deletes kaptaind resources",
			Subcommands: []cli.Command{
				{
					Name:  "task",
					Usage: "delete a task",
					Action: func(c *cli.Context) error {
						id := c.Args().First()
						if id == "" {
							fmt.Println("missing task id. try get task <id>")
						} else {
							deleteTask(id)
						}

						return nil
					},
				},
			},
		},
		{
			Name:  "run",
			Usage: "runs a new task",
			Subcommands: []cli.Command{
				{
					Name: "task",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "sourceClusterId",
							Usage: "id of the source kubernetes cluster to snapshot",
							Value: "",
						},
						cli.StringFlag{
							Name:  "targetClusterId",
							Usage: "id of the target kubernetes cluster to restore",
							Value: "",
						}},
					Action: func(c *cli.Context) error {
						source := c.String("sourceClusterId")
						target := c.String("targetClusterId")

						if source == "" {
							fmt.Println("--sourceClusterId is required")
						} else if target == "" {
							fmt.Println("--targetClusterId is required")
						} else {
							newTask(source, target)
						}

						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func newTask(source string, target string) {
	task := NewTaskRequest{}
	task.SourceClusterId = source
	task.TargetClusterId = target

	buffer := new(bytes.Buffer)
	json.NewEncoder(buffer).Encode(task)

	resp, err := http.Post(endpoint+"/tasks", "application/json; charset=utf-8", buffer)
	if err != nil {
		fmt.Println("error connecting to broker")
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	apiResponse := NewTaskAPIResponse{}
	err = json.Unmarshal(body, &apiResponse)

	if apiResponse.Error != "" {
		fmt.Println(apiResponse.Error)
	} else {
		fmt.Println("Task started successfully")
	}
}

func getClusters() {
	resp, err := http.Get(endpoint + "/clusters")
	if err != nil {
		fmt.Println("error connecting to broker")
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	apiResponse := ClustersAPIResponse{}
	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("error parsing response from broker")
		fmt.Println(err.Error())
		return
	}

	csvContent, err := gocsv.MarshalString(&apiResponse.Data)
	printTable(csvContent)
}

func printTable(csvContent string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	table.SetCenterSeparator("|")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	scanner := bufio.NewScanner(strings.NewReader(csvContent))
	header := true

	for scanner.Scan() {
		text := strings.Split(scanner.Text(), ",")

		if header {
			table.SetHeader(text)
			header = false
		} else {
			table.Append(text)
		}
	}

	table.Render()
}

func getCluster(id string) {
	resp, err := http.Get(endpoint + "/clusters/" + id)
	if err != nil {
		fmt.Println("error connecting to broker")
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	apiResponse := ClusterAPIResponse{}
	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("error parsing response from broker")
		fmt.Println(err.Error())
		return
	}

	clusterArr := []*Cluster{&apiResponse.Data}
	csvContent, err := gocsv.MarshalString(clusterArr)

	printTable(csvContent)
}

func deleteTask(id string) {
	req, err := http.NewRequest(http.MethodDelete, endpoint+"/tasks/"+id, nil)
	if err != nil {
		fmt.Println("error deleting task")
		return
	}

	_, err = http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("error deleting task")
		return
	}

	fmt.Println("task deleted")
}

func getTask(id string) {
	resp, err := http.Get(endpoint + "/tasks/" + id + "/state")
	if err != nil {
		fmt.Println("error connecting to broker")
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	apiResponse := TaskAPIResponse{}
	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("error parsing response from broker")
		fmt.Println(err.Error())
		return
	}

	tasksArr := []*Task{&apiResponse.Data}
	csvContent, err := gocsv.MarshalString(tasksArr)

	printTable(csvContent)
}

func getTasks() {
	resp, err := http.Get(endpoint + "/tasks")
	if err != nil {
		fmt.Println("error connecting to broker")
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	apiResponse := TasksAPIResponse{}
	err = json.Unmarshal(body, &apiResponse)

	if err != nil {
		fmt.Println("error parsing response from broker")
		fmt.Println(err.Error())
		return
	}

	csvContent, err := gocsv.MarshalString(&apiResponse.Data)
	printTable(csvContent)
}

func setEndpoint() {
	file, _ := ioutil.ReadFile("/.kap/config")
	config := Config{}
	err := json.Unmarshal(file, &config)

	if err != nil {
		panic("error reading config file")
	}

	endpoint = config.BrokerUrl
}
