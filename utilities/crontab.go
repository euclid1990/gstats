package utilities

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
)

func execCmd(cmd string, wg *sync.WaitGroup) {
	parts := strings.Fields(cmd)
	out, err := exec.Command(parts[0], parts[1], parts[2]).Output()
	if err != nil {
		fmt.Println("error occured")
		fmt.Printf("%s", err)
	}
	fmt.Printf("%s", out)
	wg.Done()
}

func reMakeCrontab(handle func(s string) string) {
	filename := "mycron"
	wg := new(sync.WaitGroup)
	wg.Add(1)
	execCmd("/bin/sh crontab.sh -l", wg)
	wg.Wait()

	bCrontabs, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	crontab := handle(string(bCrontabs))

	ioutil.WriteFile(filename, []byte(crontab), 0644)
	wg.Add(1)
	execCmd("/bin/sh crontab.sh -c", wg)
	wg.Wait()
}

// job string. ex: "* * * * * /bin/date >> /tmp/output"
func AddCrontab(job string) {
	reMakeCrontab(func(s string) string {
		s += job + "\n"
		return s
	})
}

// job string. ex: "* * * * * /bin/date >> /tmp/output"
func RemoveCrontab(job string) {
	reMakeCrontab(func(s string) string {
		return strings.Replace(s, job, "", -1)
	})
}
