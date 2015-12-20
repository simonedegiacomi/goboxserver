package configuration

import (
    "encoding/json"
    "os"
)

type Configuration struct {
    User        map[string]string   `json:"user"`
    onReload    []func()
}

func NewConfiguration () *Configuration {
    return &Configuration{
        User: make(map[string]string),
        onReload: make([]func(), 1),
    }
}

func (c *Configuration) Reload () {
    // Save the file
    file, err := os.Create("config.json")
    
    if err != nil {
        return
    }
    
    json.NewEncoder(file).Encode(c)
    // Execute the reload
    for _, callback := range c.onReload {
        callback()
    }
}

func (c *Configuration) AddOnReload (callback func() ) {
    c.onReload = append(c.onReload, callback)
}