package myhttputil

func JSONDecoder (response http.ResponseWriter, request *http.Request, next http.HandlerFunc) {
    
    // If the request is a GET doesn't contain the body
    // so there's no json
    if request.Method == "GET" {
        next(response, request)
    }
    
    json.NewDecoder(request).Decode
}