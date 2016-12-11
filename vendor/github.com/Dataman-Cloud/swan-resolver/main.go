package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/Dataman-Cloud/swan-resolver/nameserver"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func ServerCommand() cli.Command {
	return cli.Command{
		Name:      "server",
		Usage:     "start a dns proxy server",
		ArgsUsage: "[name]",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "domain",
				Value: "swan",
				Usage: "default doamin prefix",
			},

			cli.StringFlag{
				Name:  "listener",
				Value: "0.0.0.0",
				Usage: "default ip addr",
			},

			cli.StringFlag{
				Name:  "log-level",
				Value: "debug",
			},

			cli.IntFlag{
				Name:  "port",
				Value: 53,
				Usage: "default port",
			},
		},
		Action: func(c *cli.Context) error {
			level, _ := logrus.ParseLevel(c.String("log-level"))
			logrus.SetLevel(level)

			resolver := nameserver.NewResolver(nameserver.NewConfig(c))
			go func() {
				i := 0
				for {
					e := nameserver.RecordGeneratorChangeEvent{
						Change:       "add",
						Type:         aOrSrv(),
						Ip:           "192.168.1.1",
						Port:         "1234",
						DomainPrefix: domainGen(i),
					}
					fmt.Println(e)
					resolver.RecordGeneratorChangeChan() <- &e
					time.Sleep(time.Second * 8)
					i += 1
				}
			}()
			resolver.Start(context.Background())

			return nil
		},
	}
}

func aOrSrv() string {
	rand.Seed(time.Now().UnixNano())
	if rand.Intn(1024)%2 == 0 {
		return "a"
	} else {
		return "srv"
	}
}

func domainGen(i int) string {
	return fmt.Sprintf("task%d.appname.username", i)
}

func main() {
	resolver := cli.NewApp()
	resolver.Name = "swan-resolver"
	resolver.Usage = "command-line client for resolver"
	resolver.Version = "0.1"
	resolver.Copyright = "(c) 2016 Dataman Cloud"

	resolver.Commands = []cli.Command{
		ServerCommand(),
	}

	if err := resolver.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
