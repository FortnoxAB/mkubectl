package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
		},
		Action: func(c *cli.Context) error {
			return run(c.Context, c.String("context"), c.Args().Slice())
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, kubeContext string, kubectlCommands []string) error {
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
		var commands []string
		commands = append(commands, "--context", context)
		commands = append(commands, kubectlCommands...)
		cmd := exec.CommandContext(ctx, "kubectl", commands...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to run command %s on context %s: %w", strings.Join(kubectlCommands, " "), context, err)
		}
	}
	return nil
}
