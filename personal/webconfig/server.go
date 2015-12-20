package webconfig

import (
    "net/http"
    "goboxserver/personal/configuration"
    "io"
)

type WebConfig struct {
    
}

func NewWebConfig (conf *configuration.Configuration) *WebConfig{
    // In this case use the default http mux, there's only 2
    // handlers so who cares
    http.HandleFunc("/login", func (response http.ResponseWriter, request *http.Request) {
        conf.User["username"] = request.FormValue("username")
        conf.User["password"] = request.FormValue("password")
        conf.Reload()
    })
    
    http.HandleFunc("/", func (response http.ResponseWriter, request *http.Request) {
        io.WriteString(response, `
                <h1>ConfigurationPage</h1>
                <form action="/login" method="POST">
                    <label>Username: </label><input type="text" name="username"></br>
                    <label>Password: </label><input type="password" name="password">
                    <input type="submit" name="submit">
                </form>
        `)
    })
    
    return &WebConfig{}
}

func (c *WebConfig) ListenAndServer (address string) {
    http.ListenAndServe(address, nil)
}