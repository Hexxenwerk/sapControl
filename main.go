package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

type args struct {
	file, pass, cmd *string
	prod, debug     *bool
}

type system map[string]struct {
	Host, User string
	Inst       []string
	Prod       bool
}

func main() {
	flags := getFlags()
	systems := getSystems(flags)
	execSAPControl(flags, systems)
}

func getFlags() args {
	flags := args{
		cmd:   flag.String("cmd", "GetProcessList", "sapcontrol function"),
		file:  flag.String("file", "config.json", "Config file"),
		pass:  flag.String("pass", "", "Password for OS user <sid>adm"),
		prod:  flag.Bool("prod", false, "Set to include production systems"),
		debug: flag.Bool("debug", false, "Set to generate additional output"),
	}
	flag.Parse()
	return flags
}

func execSAPControl(flags args, systems system) {
	for sid := range systems {
		if systems[sid].Prod && !*flags.prod {
			fmt.Println("Skipping production system:", sid)
			continue
		}
		fmt.Printf("%s: Executing function %s\n", sid, *flags.cmd)
		arg := []string{"-host", systems[sid].Host, "-user", systems[sid].User, *flags.pass, "-nr", systems[sid].Inst[0], "-function", *flags.cmd}
		if *flags.debug {
			fmt.Println("/usr/sap/hostctrl/exe/sapcontrol", strings.Join(arg, " "))
		}
		cmd := exec.Command("/usr/sap/hostctrl/exe/sapcontrol", arg...)
		out, err := cmd.CombinedOutput()
		fmt.Printf("%s", out)
		if err != nil {
			switch cmd.ProcessState.ExitCode() {
			case 2, 3, 4:
				continue
			default:
				log.Fatal("Error: Failed executing OS command: ", err)
			}
		}
	}
}

func getSystems(flags args) (systems system) {
	data, err := ioutil.ReadFile(*flags.file)
	if err != nil {
		log.Fatal("Error: Failed reading file: ", err)
	}
	if err := json.Unmarshal(data, &systems); err != nil {
		log.Fatal("Error: Failed to unmarshal JSON: ", err)
	}
	return systems
}
