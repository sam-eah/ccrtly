package main

func main() {
	// var strTmp string
	// myCmd := &cobra.Command{
	//     Use: "myaction",
	//     Run: func(cmd *cobra.Command, args []string) {
	//         if len(args) == 1 {
	//             fmt.Println(strTmp)
	//         }
	//     },
	// }
	// myCmd.Flags().StringVarP((&strTmp), "test", "t", "", "Source directory to read from")
	// myCmd.Execute()
	rootCmd.Execute()
}

// var wg = sync.WaitGroup{}

// func Shellout(command string) (string, string, error) {
// 	var stdout bytes.Buffer
// 	var stderr bytes.Buffer
// 	cmd := exec.Command("bash", "-c", command)
// 	cmd.Stdout = &stdout
// 	cmd.Stderr = &stderr
// 	err := cmd.Run()
// 	return stdout.String(), stderr.String(), err
// }

// func asyncCall(env string, tenant string){
// 	fmt.Println(env, string(':'), tenant)
// 	wg.Done()
// }

// func main() {
// 	fmt.Println("Hello")

// 	buf, _ := os.ReadFile("./project.yaml")

// 	c := &[]project{}
// 	_ = yaml.Unmarshal(buf, c)

// 	// obj := struct {
// 	// 	a []project `json:"a"`
// 	// }{a: *c}

// 	// aaa, _ := json.Marshal(obj)

// 	// fmt.Println(string(aaa))

// 	proj := *c

// 	for i:=0; i<len(proj); i++ {
// 		env := proj[i]
// 		for j:=0; j<len(env.Tenants); j++ {
// 			tenant := env.Tenants[j]
// 			wg.Add(1)
// 			go asyncCall(env.Env, tenant)
// 		}
// 	}
// 	wg.Wait()

// 	fmt.Println("All done.")

// 	// out, errout, err := Shellout("ls -ltr")
// 	// if err != nil {
// 	//     log.Printf("error: %v\n", err)
// 	// }
// 	// fmt.Println("--- stdout ---")
// 	// fmt.Println(out)
// 	// fmt.Println("--- stderr ---")
// 	// fmt.Println(errout)
// }
