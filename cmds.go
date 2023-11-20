package main

import (
	"fmt"
	"os/exec"
	"sort"
	"sync"
    "strings"

	"github.com/charmbracelet/bubbles/list"
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
	cmd := exec.Command("/bin/zsh", "-c",
		str)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
		return
	}
	fmt.Println(string(output))
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

func generateEnvVariables(combo Combo, vars map[string]string, config Config, command string) string {
	str := fmt.Sprintf(`source ~/.zshrc
export CCRTLY_ENV="%s"
export CCRTLY_TENANT="%s"
`, combo.env, combo.tenant)

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

	fmt.Println(str)

	return str
}

var rootCmd = &cobra.Command{
	Use:   "ccrtly",
	Short: "A CLI for running scripts multiple envs/tenants.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := GetConfig()

		envs := config.Envs

		columns := make([]Column, 0)
		rows := make([]Row, 0)

		available := make(map[Combo]map[string]string)

		for _, env := range envs {
			columns = append(columns, Column(env.Name))
			for tenant, profiles := range env.Tenants {
				rows = append(rows, Row(tenant))
				available[Combo{env.Name, tenant}] = profiles
			}
		}

		rows = removeDuplicate(rows)

		sort.Slice(rows, func(i, j int) bool {
			return rows[i] < rows[j]
		})

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

				str := generateEnvVariables(selected, m.available[selected], config, command.Command)
				Native3(str, selected)
			}
		} else {
			fmt.Printf("Running %s concurrently for all selected profiles\n", command.Name)
			var wg = sync.WaitGroup{}

			for selected := range m.selected {
				str := generateEnvVariables(selected, m.available[selected], config, command.Command)
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
