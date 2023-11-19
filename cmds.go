package main

import (
	"fmt"
	"os/exec"
	"sort"
	"sync"

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

type Command struct {
	name       string
	command    string
	sequential bool
}

var base_commands = []Command{
	{
		name: "setup_remote_state",
		command: `echo $REMOTE_TF_WORKSPACE
asp $ASP_PROFILE
cd remote_states
terraform workspace new $REMOTE_TF_WORKSPACE
terraform apply
asp`,
	},
	{
		name: "setup_workspace",
		command: `echo $TF_WORKSPACE
asp $ASP_PROFILE
terraform init -backend-config=variables/backend_$TF_WORKSPACE.tfvars -reconfigure
terraform workspace new $TF_WORKSPACE
asp`,
	},
	{
		name: "plan",
		command: `echo $TF_WORKSPACE
asp $ASP_PROFILE
terraform init -backend-config=variables/backend_$TF_WORKSPACE.tfvars -reconfigure
terraform plan -var-file=variables/$(tf workspace show).tfvars -parallelism=200 -out=scripts/$TF_WORKSPACE.plan
asp`,
	},
	{
		name: "apply_plan",
		command: `echo $TF_WORKSPACE
asp $ASP_PROFILE
terraform init -backend-config=variables/backend_$TF_WORKSPACE.tfvars -reconfigure
terraform apply -auto-approve scripts/$TF_WORKSPACE.plan
rm scripts/$TF_WORKSPACE.plan
asp`,
	},
	{
		name: "show_plan",
		command: `echo $TF_WORKSPACE
asp $ASP_PROFILE
terraform show scripts/$TF_WORKSPACE.plan
echo "*********************************"
asp`,
		sequential: true,
	},
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

func generateEnvVariables(combo Combo, vars map[string]string) string {
	str := "source ~/.zshrc\n"

	all_vars := map[string]string{
		"TF_ENV":              combo.env,
		"TF_TENANT":           combo.tenant,
		"ASP_PROFILE":         combo.tenant + "-" + combo.env,
		"TF_WORKSPACE":        combo.tenant + "_" + combo.env,
		"REMOTE_TF_WORKSPACE": combo.tenant + "-" + combo.env,
	}

	// Section to add/override variables
	for k, v := range vars {
		all_vars[k] = v
	}

	// Add all vars to str
	for k, v := range all_vars {
		str += fmt.Sprintf("export %s=\"%s\"\n", k, v)
	}

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

		// p := tea.NewProgram(m)
		// if _, err := p.Run(); err != nil {
		// 	fmt.Println("could not start program:", err)
		// 	os.Exit(1)
		// }

		if len(m.selected) == 0 {
			fmt.Println("Nothing selected")
			return nil
		}

		// fmt.Println(m.selected)

		commands := base_commands

		for k, v := range config.Scripts {
			commands = append(commands, Command{
				name:    k,
				command: v,
			})
		}

		list := []list.Item{}
		for _, v := range commands {
			list = append(list, item(v.name))
		}
		t := createList(list)

		if t.choice == "" {
			return nil
		}

		fmt.Println("	> ", t.choice)

		var command Command
		for _, v := range commands {
			if v.name == t.choice {
				command = v
			}
		}

		if command.sequential {
			fmt.Printf("Running %s sequentially for all selected profiles\n", command.name)
			for selected := range m.selected {
				// fmt.Println(selected)

				str := generateEnvVariables(selected, m.available[selected])
				str += command.command
				Native3(str, selected)
			}
		} else {
			fmt.Printf("Running %s concurrently for all selected profiles\n", command.name)
			var wg = sync.WaitGroup{}

			for selected := range m.selected {
				str := generateEnvVariables(selected, m.available[selected])
				str += command.command
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
