package cmd

import (
	"bufio"
	"bytes"
	"context-changer/config_parser"
	"fmt"
	"github.com/spf13/cobra"
	"hash/fnv"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var initContext = &cobra.Command{
	Use:   "init",
	Short: "Initializes a context",
	Long: `Initializes a set of projects that belong to a context specified in a config json file. Example:

context-changer init -c ./config.json -x projectName`,
	Run: initializeContext,
}

func init() {
	rootCmd.AddCommand(initContext)

	initContext.Flags().StringP("config", "c", "./config.json", "Path to the config file")
	initContext.Flags().StringP("context", "x", "", "Context name")
}

func initializeContext(cmd *cobra.Command, args []string) {
	configPath, _ := cmd.Flags().GetString("config")
	contextName, _ := cmd.Flags().GetString("context")

	config, err := config_parser.ParseConfig(configPath)
	if err != nil {
		fmt.Printf("Error reading config file:\n%v\n", err)
		return
	}

	var ctx *config_parser.Context = nil
	for _, aContext := range (*config).Contexts {
		if aContext.Name == contextName {
			ctx = &aContext
			break
		}
	}
	if ctx == nil {
		fmt.Printf("Could not find context with name %s\n", contextName)
		return
	}

	var wg sync.WaitGroup

	for _, aProject := range ctx.Projects {
		fmt.Printf("Project:\n%v\n", aProject)

		wg.Add(1)
		go startProject(&aProject, &wg)
	}

	for _, aProject := range ctx.Projects {
		go startIde(&aProject)
	}

	//wg.Wait()

	/*reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Type :q to quit: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == ":q" {
			fmt.Println("Exiting project")
			terminateIde("goland64.exe")
			os.Exit(0)
		}
		fmt.Println("You entered:", input)
	}*/

	fmt.Printf("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)

	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	terminateIde("goland64.exe")
}

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

func terminateIde(ideProcessName string) {
	result, err := findProcessByName(ideProcessName)
	fmt.Println("Terminating")
	fmt.Printf("%v\n", err)
	if err == nil {
		pidStr := strings.Fields(result)[1]
		pid, _ := strconv.Atoi(pidStr)
		process, _ := os.FindProcess(pid)
		process.Kill()
	}
}
func startIde(project *config_parser.Project) {
	cmd := exec.Command("cmd", "/C", project.Ide, "--project", project.Path)

	execCommand(cmd, project.Name)
}

func startProject(project *config_parser.Project, wg *sync.WaitGroup) {
	defer wg.Done()

	command := strings.Split(project.Start, " ")

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = project.Path

	execCommand(cmd, project.Name)
}

func execCommand(cmd *exec.Cmd, projectName string) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error obtaining stdout pipe:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting application:", err)
		return
	}

	scanner := bufio.NewScanner(stdout)

	go func() {
		color := getColor(projectName)
		for scanner.Scan() {
			fmt.Printf("%s%s%s: %s\n", color, projectName, Reset, scanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for command to finish:", err)
		return
	}

	fmt.Println("Command finished successfully.")
}

func getColor(projectName string) string {
	colors := []string{Red, Green, Yellow, Blue, Purple, Cyan, Gray}

	h := fnv.New32()
	h.Write([]byte(projectName))
	hash := int(h.Sum32())

	return colors[hash%len(colors)]
}

func findProcessByName(processName string) (string, error) {
	cmd := exec.Command("taskList.exe")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var process string
	for _, line := range strings.Split(out.String(), "\n") {
		if strings.Contains(line, processName) {
			process = line
			break
		}
	}
	return process, nil
}
