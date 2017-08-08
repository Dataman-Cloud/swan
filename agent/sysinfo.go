package agent

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry/gosigar"

	"github.com/Dataman-Cloud/swan/types"
)

type gatherFunc func() error

type gatherer struct {
	info types.SysInfo
}

func (g *gatherer) os() error {
	bs, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return err
	}

	g.info.OS = string(bs)
	return nil
}

func (g *gatherer) hostName() error {
	name, err := os.Hostname()
	if err != nil {
		return err
	}

	g.info.Hostname = name
	return nil
}

func (g *gatherer) uptime() error {
	up := sigar.Uptime{}
	if err := up.Get(); err != nil {
		return err
	}

	g.info.Uptime = strconv.FormatFloat(up.Length, 'f', 6, 64)
	return nil
}

func (g *gatherer) unixTime() error {
	g.info.UnixTime = time.Now().Unix()
	return nil
}

func (g *gatherer) loadAvg() error {
	avg := sigar.LoadAverage{}
	if err := avg.Get(); err != nil {
		return err
	}

	g.info.LoadAvg = avg.One
	return nil
}

func (g *gatherer) cpu() error {
	cpulist := sigar.CpuList{}
	if err := cpulist.Get(); err != nil {
		return err
	}

	var (
		processorN  = int64(len(cpulist.List))
		physicalN   int64
		idle, total uint64
		used        float64
	)

	// used
	for _, cpu := range cpulist.List {
		idle += cpu.Idle
		total += cpu.Total()
	}
	used = 100.0 - float64(idle*100)/float64(total)

	// physical
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return err
	}
	defer f.Close()

	var (
		scanner = bufio.NewScanner(f)
		reg     = regexp.MustCompile(`^physical id\s+:\s+(\d+)$`)
	)
	for scanner.Scan() {
		subMatch := reg.FindSubmatch([]byte(scanner.Text()))
		if len(subMatch) >= 2 {
			physicalN++
		}
	}

	g.info.CPU = types.CPUInfo{
		Processor: processorN,
		Physical:  physicalN,
		Used:      used,
	}
	return nil
}

func (g *gatherer) memory() error {
	mem := sigar.Mem{}
	if err := mem.Get(); err != nil {
		return err
	}

	g.info.Memory = types.MemoryInfo{
		Total: int64(mem.Total),
		Used:  int64(mem.ActualUsed),
	}
	return nil
}

func (g *gatherer) ips() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	ips := make(map[string][]string)
	for _, iface := range ifaces {
		if iface.Name == "" {
			continue
		}
		iface.Name = strings.NewReplacer([]string{
			".", "-",
			"$", "-",
		}...).Replace(iface.Name)
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		ifaddrs := []string{}
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					ifaddrs = append(ifaddrs, ipnet.IP.String())
				}
			}
		}
		if len(ifaddrs) > 0 {
			ips[iface.Name] = ifaddrs
		}
	}

	g.info.IPs = ips
	return nil
}

func (g *gatherer) osListenings() error {
	var (
		tcpf, tcp6f = "/proc/net/tcp", "/proc/net/tcp6"
		bufReader   = bufio.NewReader(nil)
	)
	defer bufReader.Reset(nil)

	for _, file := range []string{tcpf, tcp6f} {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		bufReader.Reset(f)
		for {
			line, err := bufReader.ReadString('\n')
			if err != nil {
				break
			}

			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}

			if parts[3] != "0A" {
				continue
			}

			parts = strings.SplitN(parts[1], ":", 2)
			if len(parts) != 2 {
				continue
			}

			listening, err := strconv.ParseInt(parts[1], 16, 32)
			if err != nil {
				continue
			}

			g.info.Listenings = append(g.info.Listenings, listening)
		}
	}

	return nil
}

func (g *gatherer) containers() error {
	return nil // TODO
}

func Gather() (*types.SysInfo, error) {
	g := new(gatherer)
	funcs := []gatherFunc{
		g.os,
		g.hostName,
		g.uptime,
		g.unixTime,
		g.loadAvg,
		g.cpu,
		g.memory,
		g.ips,
		g.containers,
		g.osListenings,
	}

	for _, fun := range funcs {
		if err := fun(); err != nil {
			return nil, fmt.Errorf("Gather.Sysinfo.%v error: %v", funcName(fun), err)
		}
	}

	return &g.info, nil
}

func funcName(f gatherFunc) string {
	v := reflect.ValueOf(f)
	fname := runtime.FuncForPC(v.Pointer()).Name()
	return path.Base(fname)
}
