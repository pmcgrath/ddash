package main

import (
	"fmt"
	"html/template"
	"time"
)

var (
	containersTemplate *template.Template
)

func init() {
	containersTemplate = template.Must(template.New("containers").Funcs(templateFuncMap).Parse(containersHtmlTemplate))
}

var templateFuncMap = template.FuncMap{
	"displayPort": func(p port) string {
		// Ape "docker ps"
		if p.Host > 0 {
			return fmt.Sprintf("%d->%d/%s", p.Host, p.Container, p.Type)
		}

		return fmt.Sprintf("%d->", p.Host)

	},
	"displayRestartPolicy": func(c container) string {
		if c.RestartPolicyName == "" {
			return "None"
		}
		if c.RestartPolicyMaximumRetryCount > 0 {
			return c.RestartPolicyName + " : " + string(c.RestartPolicyMaximumRetryCount)
		}

		return c.RestartPolicyName
	},
	"displayStatus": func(c container) string {
		if !c.Running {
			return "stopped"
		}
		if c.Paused {
			return "paused"
		}

		return "running"
	},
	"displayTimestamp": func(t time.Time) string {
		if t.Year() == 1 {
			return ""
		}
		return t.Format("Jan 2 2006 15:04:05")
	},
	"displayVolume": func(v volume) string {
		writeable := ""
		if v.ReadWrite {
			writeable = "RW: "
		}
		return fmt.Sprintf("%s%s->%s", writeable, v.Container, v.Host)
	},
}

// Could have used https://github.com/jteeuwen/go-bindata
const containersHtmlTemplate = `
<html>
    <head>
        <title>Containers</title>
        <style type="text/css">
            .table          { display: table; }
            .heading        { display: table-row; font-weight: bold; }
            .row            { display: table-row; }
	    .cell           { display: table-cell; padding-left: 10px; padding-top: 5px; }
	    .status         { color: black; font-weight: bold; }
            .status.running { color: green; }
            .status.paused  { color: yellow; }
            .status.stopped { color: red; }
        </style>
        <script type="text/javascript">
            var sock = null;
            var uri = "ws://" + window.location.host + "{{.SocketPath}}";

            window.onload = function() {
                sock = new WebSocket(uri);
   
                sock.onopen = function() {
                    console.log("WebSocket: Connected to " + uri);
                }

                sock.onclose = function(e) {
                    console.log("WebSocket: Connection closed (" + e.code + ")");
                }
 
                sock.onerror = function(e) {
                    console.log("WebSocket: Connection error detected (" + e + ")");
                }

                sock.onmessage = function(e) {
                    console.log("WebSocket: Message received: " + e.data);
                    window.location.reload(); // Lazy but plenty good here, don't have to deal with a lack of data and changed data
                }
            };
        </script>
    </head>
    <body>
        <h1>Containers</h1>
        <div class="table">
            <div class="heading">
                <div class="cell">Id</div>
                <div class="cell">Name</div>
                <div class="cell">Pid</div>
                <div class="cell">Started</div>
                <div class="cell">Finished</div>
                <div class="cell">Restart Policy</div>
                <div class="cell">Ports</div>
                <div class="cell">Volumes From</div>
                <div class="cell">Volumes</div>
            </div>
            {{range .Containers}}
                <div class="row" id="{{.Id}}">
                    <div class="cell"><a href="/containers/{{.Id}}">{{.ShortId}}</a></div>
                    <div class="cell">{{.Name}}</div>
                    <div class="cell status {{displayStatus .}}">{{.Pid}}</div>
                    <div class="cell">{{displayTimestamp .StartedAt}}</div>
                    <div class="cell">{{displayTimestamp .FinishedAt}}</div>
                    <div class="cell">{{displayRestartPolicy .}}</div>
                    <div class="cell">{{range .Ports}}{{displayPort .}}<br/>{{end}}</div>
                    <div class="cell">{{.VolumesFrom}}</div>
                    <div class="cell">{{range .Volumes}}{{displayVolume .}}<br/>{{end}}</div>
                </div>
            {{end}}
        </div>
    </body>
</html>
`
