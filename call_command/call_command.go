package call_command

import (
	"fmt"
	"os/exec"
	//"log"
)


type Command struct {
	CommandName string
	Args []string
}

func New(command string) *Command {
	c := new(Command)
	c.CommandName =  command
	return c
}

func (c *Command) AddArgs(args []string) {
	for _,value := range args{
		c.Args = append(c.Args,value)
	}
}

//Function for calling certain command
func (c Command) Call() string {
	command := exec.Command(c.CommandName,c.Args...)
	Stout,err := command.CombinedOutput()
	if err != nil {
		fmt.Printf("[ERROR] When launching the command: %s\n",c.CommandName)
		//log.Fatal(err)
		return ""
	}

	return string(Stout)
}
