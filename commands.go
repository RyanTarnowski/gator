package main

import (
	"errors"
	"fmt"
)

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	run_cmd, ok := c.commands[cmd.name]
	if !ok {
		return errors.New("command not found")
	}

	return run_cmd(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Login requires a username input")
	}

	username := cmd.args[0]

	err := s.config.SetUser(username)
	if err != nil {
		return fmt.Errorf("Not able to set current user: %w", err)
	}

	fmt.Printf("User %v has been logged in.\n", username)
	return nil
}

