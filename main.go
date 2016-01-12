package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
)

const (
	memoryUsagePath  = "/sys/fs/cgroup/memory/memory.usage_in_bytes"
	eventControlPath = "/sys/fs/cgroup/memory/cgroup.event_control"

	// stress -m 40
	threshold = 4037235968
)

func main() {
	sysEventfd, _, syserr := syscall.RawSyscall(
		syscall.SYS_EVENTFD2, 0, syscall.FD_CLOEXEC, 0,
	)
	if syserr != 0 {
		log.Fatal(syserr.Error())
	}

	eventfd := os.NewFile(sysEventfd, "eventfd")

	memoryUsage, err := os.Open(memoryUsagePath)
	if err != nil {
		log.Fatal(err)
	}

	eventControlData := fmt.Sprintf(
		"%d %d %d",
		eventfd.Fd(), memoryUsage.Fd(), threshold,
	)

	err = ioutil.WriteFile(
		eventControlPath,
		[]byte(eventControlData),
		0222,
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(eventControlData)

	for {
		buf := make([]byte, 8)
		_, err := eventfd.Read(buf)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("TRESHOLD CROSSED")

		// check that cgroup is not destroyed
		_, err = os.Lstat(eventControlPath)
		if os.IsNotExist(err) {
			log.Println("cgroup is destroyed, exiting")
			return
		}
	}

}
