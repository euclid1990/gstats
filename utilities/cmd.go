package utilities

import (
	"sync"
	"fmt"
	"strings"
	"os/exec"
	"os"
)

func exe_cmd(cmd string, wg *sync.WaitGroup) {
	fmt.Println(cmd)
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0],parts[1]).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
	wg.Done()
}

func AddSystemCrontab(min, hour, dayofmonth, month, dayofweek, command string)  {
	wg := new(sync.WaitGroup)
	wg.Add(1)
	os.Setenv("MIN", min)
	os.Setenv("HOUR", hour)
	os.Setenv("DAYOFMONTH", dayofmonth)
	os.Setenv("MONTH", month)
	os.Setenv("DAYOFWEEK", dayofweek)
	os.Setenv("COMMAND", command)
	go exe_cmd("/bin/sh crontab.sh", wg)
	wg.Wait()
}
