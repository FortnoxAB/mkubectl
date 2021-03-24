package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "mkubectl",
		Usage: "run kubectl command in multiple contexts",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "context",
				Value:   "",
				Usage:   "regexp kubectl context name",
				Aliases: []string{"c"},
			},
			&cli.StringFlag{
				Name:    "namespace",
				Value:   "",
				Usage:   "kubectl namespace",
				Aliases: []string{"n"},
			},
			&cli.StringFlag{
				Name:  "log-level",
				Value: "info",
			},
		},
		Action: func(c *cli.Context) error {
			lvl, err := logrus.ParseLevel(c.String("log-level"))
			if err != nil {
				return err
			}

			logrus.SetLevel(lvl)
			return run(c.Context, c.String("context"), c.String("namespace"), c.Args().Slice())
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, kubeContext string, namespace string, kubectlCommands []string) error {
	r, err := regexp.Compile(kubeContext)
	if err != nil {
		return fmt.Errorf("failed to compile regexp: %w", err)
	}
	cmd := exec.CommandContext(ctx, "kubectl", "config", "get-contexts", "-o", "name")
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to execute kubectl config: %w", err)
	}
	s := strings.TrimRight(string(bytes), "\n")
	contexts := strings.Split(s, "\n")

	var clusters []string
	for _, context := range contexts {
		context := strings.TrimSpace(context)
		s := r.FindString(context)
		if s == "" {
			continue
		}
		clusters = append(clusters, s)
	}
	for _, context := range clusters {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		logrus.Debugf("running in context %s", context)
		var commands []string
		commands = append(commands, "--context", context)
		if namespace != "" {
			commands = append(commands, "--namespace", namespace)
		}
		commands = append(commands, kubectlCommands...)
		cmd := exec.CommandContext(ctx, "kubectl", commands...)

		cmd.Stdout = NewContextWriter(context)
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run command %s on context %s: %w", strings.Join(kubectlCommands, " "), context, err)
		}
	}
	return nil
}

type contextWriter struct {
	context string
}

func NewContextWriter(context string) *contextWriter {
	return &contextWriter{context: context}
}

func (cw *contextWriter) Write(p []byte) (n int, err error) {
	for _, a := range p {
		if string(a) == "\n" {
			fmt.Fprint(os.Stdout, "\t", cw.context)
		}
		fmt.Fprint(os.Stdout, string(a))
	}
	return len(p), nil
}
