package configparser

import (
	"fmt"
	"io/ioutil"
	"strings"
)

const NEW_LINE string = "\n"
const SEPARATOR string = "="

type ConfigParser struct {
	Name string
	Content string
	ConfigOpts map[string]string
}

func (c *ConfigParser) Load(filename string){
  c.Name = filename
  c.ConfigOpts = make(map[string]string)
  content,err := ioutil.ReadFile(c.Name)
  if err != nil {
	fmt.Printf("[ERROR] Loading config file: %s\n",c.Name)
	return
  }
  c.Content = strings.Trim(string(content),NEW_LINE)
  lines := strings.Split(c.Content,NEW_LINE)
  for _,line := range lines {
	  current := strings.Split(line,SEPARATOR)
	  c.ConfigOpts[current[0]] = current[1]
  }
}

func (c ConfigParser) GetConfigOpts() map[string] string {
   return c.ConfigOpts
}

func (c ConfigParser) Show() {
    fmt.Println(c.Content)
}
