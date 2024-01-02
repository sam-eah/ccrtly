package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// func initialListModel() ListModel {
// 	t := NewListModel()
// 	if _, err := tea.NewProgram(t).Run(); err != nil {
// 		fmt.Println("Error running program:", err)
// 		os.Exit(1)
// 	}

// 	return t
// }

func Native() string {
	cmd, err := exec.Command("terraform", "init").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output := string(cmd)
	return output
}

// this is to execute a script in a file
func Native2(file_path string) string {
	cmd, err := exec.Command("/bin/sh", file_path).Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output := string(cmd)
	return output
}

func Native3(str string, combo Combo) {
	log.SetLevel(log.DebugLevel)
	cmd := exec.Command("/bin/zsh", "-c",
		str)

	// create a pipe for the output of the script
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return
	}
	log.Debug("Created StdoutPipe for Cmd.")
	cmd.Stderr = cmd.Stdout

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "Error") {
				log.Errorf("[%-4s|%12s] %s\n", combo.env, combo.tenant, scanner.Text())
			} else {
				log.Infof("[%-4s|%12s] %s\n", combo.env, combo.tenant, scanner.Text())
			}
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		return
	}
	log.Debug("Started Cmd.")

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		return
	}
	fmt.Println()
	log.Debug("Waited for Cmd.")

	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	fmt.Println(fmt.Sprint(err) + ": " + string(output))
	// 	return
	// }
	// fmt.Println(string(output))
	// cmd.Stdout = os.Stdout
	// cmd.Stderr = os.Stderr
	// fmt.Println("1111")
	// 	// open the out file for writing
	// 	outfile, err := os.Create("./" + combo.env + "-" + combo.tenant + ".txt")
	// 	if err != nil {
	// 		fmt.Println(err)
	// 	}
	// 	defer outfile.Close()
	// 	cmd.Stdout = outfile
	// fmt.Println("2222")

	// err = cmd.Run()
	//
	//	if err != nil {
	//		fmt.Printf("error %s", err)
	//	}
	//
	// // output := string(cmd.)
	// // return output
}

func generateEnvVariables(combo Combo, vars map[string]string, filename string, config Config, command string) string {
	str := fmt.Sprintf(`source ~/.zshrc
export CCRTLY_ENV="%s"
export CCRTLY_TENANT="%s"
export CCRTLY_FILENAME="%s"
`, combo.env, combo.tenant, filename)

	// Config level variables
	for k, v := range config.Variables {
		str += fmt.Sprintf("export %s=\"%s\"\n", strings.ToUpper(k), v)
	}

	// Tenant level variables
	if tenant, ok := config.Tenants[combo.tenant]; ok {
		for k, v := range tenant {
			str += fmt.Sprintf("export %s=\"%s\"\n", strings.ToUpper(k), v)
		}
	}

	// Combo level variables
	for k, v := range vars {
		str += fmt.Sprintf("export %s=\"%s\"\n", strings.ToUpper(k), v)
	}

	str += config.Prescript
	str += command
	str += config.Postscript

	// fmt.Println(str)
	log.Info(str)

	return str
}

var rootCmd = &cobra.Command{
	Use:   "ccrtly",
	Short: "A CLI for running scripts multiple envs/tenants.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfig()

		available, columns, rows := getCombos()

		m := initialModel(columns, rows, available)

		if len(m.selected) == 0 {
			fmt.Println("Nothing selected")
			return nil
		}

		list := []list.Item{}
		for _, v := range config.Scripts {
			list = append(list, item(v.Name))
		}
		t := createList(list)

		if t.choice == "" {
			return nil
		}

		fmt.Println("	> ", t.choice)

		var command Script
		for _, v := range config.Scripts {
			if v.Name == t.choice {
				command = v
			}
		}

		if command.Sequentially {
			fmt.Printf("Running %s sequentially for all selected profiles\n", command.Name)
			for selected := range m.selected {
				// fmt.Println(selected)

				str := generateEnvVariables(selected, m.available[selected].vars, m.available[selected].filename, config, command.Command)
				Native3(str, selected)
			}
		} else {
			fmt.Printf("Running %s concurrently for all selected profiles\n", command.Name)
			var wg = sync.WaitGroup{}

			for selected := range m.selected {
				str := generateEnvVariables(selected, m.available[selected].vars, m.available[selected].filename, config, command.Command)
				// fmt.Println(str)
				wg.Add(1)
				go func(selected Combo) {
					defer wg.Done()
					Native3(str, selected)
				}(selected)
			}
			wg.Wait()
		}

		return nil
	},
}

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "help",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	helpCmd.Flags().StringP(
		"help",
		"h",
		"",
		"help",
	)
	// fmt.Println("help")
	rootCmd.AddCommand(helpCmd)
}

func getCombos() (map[Combo]ComboContent, []Column, []Row) {

	entries, err := os.ReadDir("./variables")
	if err != nil {
		log.Fatal(err)
	}

	columns := make([]Column, 0)
	rows := make([]Row, 0)
	available := make(map[Combo]ComboContent)

	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".tfvars" && !strings.HasPrefix(e.Name(), "backend") {
			filename := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))

			name := strings.ToLower(filename)
			name = strings.ReplaceAll(name, "_", "-")
			name = strings.ReplaceAll(name, "-prod", "-prd")

			suffix := ""

			if strings.HasSuffix(name, "-2") {
				suffix = "-2"
				name = strings.TrimSuffix(name, "-2")
			}
			if strings.HasSuffix(name, "-3") {
				suffix = "-3"
				name = strings.TrimSuffix(name, "-3")
			}

			lastInd := strings.LastIndex(name, "-")
			tenant := name[:lastInd] + suffix
			env := name[lastInd+1:] // o/p: ew

			fmt.Println(tenant, env)
			columns = append(columns, Column(env))
			rows = append(rows, Row(tenant))
			available[Combo{env, tenant}] = ComboContent{
				filename: filename,
				vars: make(map[string]string),
			}
		}
	}

	rows = removeDuplicate(rows)
	columns = removeDuplicate(columns)

	sort.Slice(rows, func(i, j int) bool {
		return rows[i] < rows[j]
	})

	return available, columns, rows
}
