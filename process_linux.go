package psps

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

// LinuxProcess is an implementation of Process that contains Unix-specific
// fields and information.
type LinuxProcess struct {
	pid   int

	state rune

	ppid  int
	pgid  int
	sid   int

	name string
	imagePath string
	cmdLine []string
	cwd string
}

func (p *LinuxProcess) Pid() int {
	return p.pid
}

func (p *LinuxProcess) PPid() int {
	return p.ppid
}

func (p *LinuxProcess) PGid() int {
	return p.pgid
}

func (p *LinuxProcess) Name() string {
	return p.name
}

func (p *LinuxProcess) ImagePath() string {
	return p.imagePath
}

func (p *LinuxProcess) CmdLine() []string {
	return p.cmdLine
}

func (p *LinuxProcess) Cwd() string {
	return p.cwd
}

// Refresh all the data associated with this process.
func (p *LinuxProcess) Refresh() error {
	processDir := fmt.Sprintf("/proc/%d", p.pid)

	statPath := path.Join(processDir, "stat")
	dataBytes, err := ioutil.ReadFile(statPath)
	if err != nil {
		return err
	}

	// First, parse out the image name
	data := string(dataBytes)
	binStart := strings.IndexRune(data, '(') + 1
	binEnd := strings.IndexRune(data[binStart:], ')')
	p.name = data[binStart : binStart+binEnd]

	// Move past the image name and start parsing the rest
	data = data[binStart+binEnd+2:]
	_, err = fmt.Sscanf(data,
		"%c %d %d %d",
		&p.state,
		&p.ppid,
		&p.pgid,
		&p.sid)

	if err != nil {
		return err
	}

	exePath := path.Join(processDir, "exe")
	fileInfo, err := os.Lstat(exePath)
	if err != nil {
		return err
	}
	if fileInfo.Mode() & os.ModeSymlink != 0 {
		link, err := os.Readlink(exePath)
		if err == nil {
			p.imagePath = link
		}
	}

	cwdPath := path.Join(processDir, "cwd")
	fileInfo, err = os.Lstat(cwdPath)
	if err != nil {
		return err
	}
	if fileInfo.Mode() & os.ModeSymlink != 0 {
		link, err := os.Readlink(cwdPath)
		if err == nil {
			p.cwd = link
		}
	}

	cmdPath := path.Join(processDir, "cmdline")
	cmdBytes, err := ioutil.ReadFile(cmdPath)
	if err != nil {
		return err
	}
	if len(cmdBytes) > 0 {
		p.cmdLine = strings.Split(string(bytes.TrimRight(cmdBytes, string("\x00"))), string(byte(0)))
	}

	return err
}

//func findProcess(pid int) (Process, error) {
//	dir := fmt.Sprintf("/proc/%d", pid)
//	_, err := os.Stat(dir)
//	if err != nil {
//		if os.IsNotExist(err) {
//			return nil, nil
//		}
//
//		return nil, err
//	}
//
//	return newLinuxProcess(pid)
//}

func Processes() ([]Process, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	results := make([]Process, 0, 50)
	for {
		fis, err := d.Readdir(10)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		for _, fi := range fis {
			// We only care about directories, since all pids are dirs
			if !fi.IsDir() {
				continue
			}

			// We only care if the name starts with a numeric
			name := fi.Name()
			if name[0] < '0' || name[0] > '9' {
				continue
			}

			// From this point forward, any errors we just ignore, because
			// it might simply be that the process doesn't exist anymore.
			pid, err := strconv.ParseInt(name, 10, 0)
			if err != nil {
				continue
			}

			p, err := newLinuxProcess(int(pid))
			if err != nil {
				continue
			}

			results = append(results, p)
		}
	}

	return results, nil
}

func newLinuxProcess(pid int) (*LinuxProcess, error) {
	p := &LinuxProcess{pid: pid}
	return p, p.Refresh()
}