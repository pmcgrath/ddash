package main

import (
	"html/template"
)

var (
	containersTemplate *template.Template
)

func init() {
	containersTemplate = template.Must(template.New("containers").Parse(containersHtmlTemplate))
}

// Could have used https://github.com/jteeuwen/go-bindata
const containersHtmlTemplate = `
<html>
    <head>
        <title>Containers</title>
        <script type="text/javascript">
            var sock = null;
            var uri = "ws://" + window.location.host + "{{.SocketPath}}";

            window.onload = function() {
                console.log("onload");
                sock = new WebSocket(uri);
   
                sock.onopen = function() {
                    console.log("Connected to " + uri);
                }

                sock.onclose = function(e) {
                    console.log("Connection closed (" + e.code + ")");
                }
 
                sock.onerror = function(e) {
                    console.log("Connection error detected (" + e + ")");
                }

                sock.onmessage = function(e) {
                    console.log("message received: " + e.data);
                }
            };
        </script>
    </head>
    <body>
        <h1>Containers</h1>
        {{range .Containers}}
            Id: {{.Id}}
            Name: {{.Name}}
            Pid: {{.Pid}}
            <br/>
        {{end}}
    </body>
</html>
`
