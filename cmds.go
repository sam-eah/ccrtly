package main

import (
	"bufio"
	"fmt"
	"html/template"

	"time"

	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"

	// "sync"

	// "time"

	// "golang.org/x/net/websocket"

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
	time.Sleep(2 * time.Second)
	log.SetLevel(log.DebugLevel)
	cmd := exec.Command("/bin/zsh", "-c", str)

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
			chans[combo] <- scanner.Text()
			if strings.Contains(scanner.Text(), "Error") {
				log.Errorf("[%-4s|%12s] %s\n", combo.env, combo.tenant, scanner.Text())
			} else {
				log.Infof("[%-4s|%12s] %s\n", combo.env, combo.tenant, scanner.Text())
			}
		}
		close(chans[combo])
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

		for combo := range m.selected {
			chans[combo] = make(chan string)
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

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// fmt.Println("request received /")
			// sls := make([][]string, 0)
			// for combo := range m.selected {
			// 	sl := ChanToSlice[string](chans[combo])
			// 	sls = append(sls, sl)
			// }
			// fmt.Println(sls)
			tmpl := template.Must(template.ParseFiles("index.html"))
			tmpl.Execute(w, nil)
		})

		http.HandleFunc("/random", chanHandler)

		if command.Sequentially {
			fmt.Printf("Running %s sequentially for all selected profiles\n", command.Name)
			for selected := range m.selected {
				// fmt.Println(selected)

				str := generateEnvVariables(selected, m.available[selected], config, command.Command)
				Native3(str, selected)
			}
		} else {
			fmt.Printf("Running %s concurrently for all selected profiles\n", command.Name)
			// var wg = sync.WaitGroup{}

			for selected := range m.selected {
				str := generateEnvVariables(selected, m.available[selected], config, command.Command)
				// fmt.Println(str)
				// wg.Add(1)
				go func(selected Combo) {
					// defer wg.Done()
					// time.Sleep(10 * time.Second)
					Native3(str, selected)
				}(selected)
			}
			// for m := range ch {
			// 	// We got a message. Do the usual SSE stuff.
			// 	fmt.Println("msg received in chan", m)
			// }
			// wg.Wait()
			// fmt.Println(ch)
		}
		fmt.Println("listen")
		log.Fatal(http.ListenAndServe(":8000", nil))

		open("localhost:8000")

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

var chans = make(map[Combo](chan string))

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// type Server struct {
//     conns map[*websocket.Conn]bool
// }

// func NewServer() *Server {
//     return &Server{
//         conns: make(map[*websocket.Conn]bool),
//     }
// }

func randomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	// send a random number every 2 seconds
	for {
		rand.Seed(time.Now().UnixNano())
		fmt.Fprintf(w, "data: %d \n\n", rand.Intn(100))
		w.(http.Flusher).Flush()
		time.Sleep(2 * time.Second)
	}
}

func chanHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	// ch := make(chan string)

	for _, ch := range chans {
		fmt.Println("here1")
		go func(ch chan string) {

			fmt.Println("here")
			for {
				select {
				case <-r.Context().Done():
					// User closed the connection. We are out of here.
					return
				case m, err := <-ch:
					// We got a message. Do the usual SSE stuff.
					if err {
						fmt.Println(err)
						return
					}

					fmt.Println("here>", m)
					fmt.Fprintf(w, "data: %s \n\n", m)
					w.(http.Flusher).Flush()
				}
			}
		}(ch)

	}
}

func ChanToSlice[T any](ch chan T) []T {
	ts := make([]T, 0)
	for t := range ch {
		ts = append(ts, t)
	}
	return ts
}
