package cmd

import (
	"errors"
	"fmt"
)

type App struct {
	Name     string
	Commands []Command
}

type Command struct {
	Name        string
	Description string
	Flags       []Flag
	action      func(*context)
	args        func() commandArgs
}

type Flag struct {
	Name        string
	Value       string
	EnvVariable string
}

type commandArgs interface {
	init(flags []Flag, args []string) error
}

type context struct {
	command *Command
	args    commandArgs
	//flags   commandArgs
}

func (app *App) Run(args []string) {
	ctx, err := app.parseContext(args)
	if err != nil {
		fmt.Println(err)
		app.help()
		return
	}

	ctx.command.action(ctx)
}

func (app *App) parseContext(args []string) (*context, error) {
	if len(args) <= 1 {
		return nil, errors.New("no command specified")
	}

	var command *Command
	commandName := args[1]
	for _, com := range app.Commands {
		if commandName == com.Name {
			command = &com
			break
		}
	}

	if command == nil {
		return nil, fmt.Errorf("unknown command %s", commandName)
	}

	commandArgs := command.args()
	err := commandArgs.init(command.Flags, args[2:])
	if err != nil {
		return nil, err
	}

	return &context{command, commandArgs}, nil
}

func (app *App) help() {
	fmt.Println(app.Name)
	fmt.Println("Commands:")
	for _, c := range app.Commands {
		fmt.Println(c.Name)
		fmt.Println(c.Description)

		if len(c.Flags) > 0 {
			fmt.Println("    flags:")
			for _, f := range c.Flags {
				fmt.Printf("    %s (default: %s)\n", f.Name, f.Value)
			}
		}
		fmt.Println()
	}

}
