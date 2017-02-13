package main

import (
	"fmt"
	"os"

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
				Value: "swan.com",
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
				a := nameserver.RecordGeneratorChangeEvent{
					Change:       "add",
					Type:         "a",
					Ip:           "192.168.1.1",
					DomainPrefix: "0.mysql.xcm.cluster",
				}
				resolver.RecordGeneratorChangeChan() <- &a

				a1 := nameserver.RecordGeneratorChangeEvent{
					Change:       "add",
					Type:         "a",
					Ip:           "192.168.1.2",
					DomainPrefix: "1.mysql.xcm.cluster",
				}
				resolver.RecordGeneratorChangeChan() <- &a1

				srv := nameserver.RecordGeneratorChangeEvent{
					Change:       "add",
					Type:         "srv",
					Ip:           "192.168.1.3",
					Port:         "1234",
					DomainPrefix: "0.nginx.xcm.cluster",
				}
				resolver.RecordGeneratorChangeChan() <- &srv

				srv1 := nameserver.RecordGeneratorChangeEvent{
					Change:       "add",
					Type:         "srv",
					Ip:           "192.168.1.4",
					Port:         "1235",
					DomainPrefix: "1.nginx.xcm.cluster",
				}
				resolver.RecordGeneratorChangeChan() <- &srv1

				proxy1 := nameserver.RecordGeneratorChangeEvent{
					Change:  "add",
					Type:    "a",
					Ip:      "192.168.1.5",
					IsProxy: true,
				}
				resolver.RecordGeneratorChangeChan() <- &proxy1

				proxy2 := nameserver.RecordGeneratorChangeEvent{
					Change:  "add",
					Type:    "a",
					Ip:      "192.168.1.6",
					IsProxy: true,
				}
				resolver.RecordGeneratorChangeChan() <- &proxy2

				da1 := nameserver.RecordGeneratorChangeEvent{
					Change:       "del",
					Type:         "a",
					Ip:           "192.168.1.2",
					DomainPrefix: "1.mysql.xcm.cluster",
				}
				resolver.RecordGeneratorChangeChan() <- &da1
			}()
			resolver.Start(context.Background())

			return nil
		},
	}
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
