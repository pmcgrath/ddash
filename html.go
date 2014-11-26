package main

import (
	"html/template"
)

var (
	rootTemplate *template.Template
)

func init() {
	rootTemplate = template.Must(template.New("root").Parse(rootHtmlTemplate))
}

// Could have used https://github.com/jteeuwen/go-bindata
const rootHtmlTemplate = `
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
            var containerUrlPrefix = "http://" + window.location.host + "{{.ContainerPathPrefix}}";
            var containersUrl = "http://" + window.location.host + "{{.ContainersPath}}";
            var eventsUrl = "ws://" + window.location.host + "{{.SocketPath}}";

            var eventsSocket = null;

            function getContainerStatus(container) {
                if (!container.State.Running) {
                    return "stopped";
                }
                if (container.State.Paused) {
                    return "paused";
                }

                return "running";
            }

            function getTimestamp(dateString) {
                    // See https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/DateTimeFormat
                    var options = { day: "numeric", month: "short", year: "numeric", hour: "2-digit", minute: "2-digit", second: "2-digit" };
                    var date = new Date(dateString);
                    if (date.getFullYear() == 1) { return ""; }
                    return date.toLocaleTimeString(navigator.language, options);
            }

            function addContainer(container) {
                var shortId = container.Id.substring(0, 12);
                var containerUrl = "/containers/" + container.Id;
                var status = getContainerStatus(container);
                var started = getTimestamp(container.State.StartedAt);
                var finished = getTimestamp(container.State.FinishedAt);

                var restartPolicy = container.HostConfig.RestartPolicy.Name;
                if (container.HostConfig.RestartPolicy.MaximumRetryCount > 0) {
                    restartPolicy += " : " + container.HostConfig.RestartPolicy.MaximumRetryCount;
                }

                var ports = "";
                for (var containerPort in container.NetworkSettings.Ports) {
                    var portSetting = container.NetworkSettings.Ports[containerPort];
                    if (portSetting != null) {
                        ports += portSetting[0].HostPort + "->"; 
                    }
                    ports += containerPort + "<br/>" 
                }

                var volumes = "";
                for (var containerPath in container.Volumes) {
                    volumes += containerPath + " : " + container.Volumes[containerPath] + "<br/>" 
                }

                var containersElement = document.getElementById("containers");
                var template = document.querySelector("#containerTemplate");
                var content = document.importNode(template.content, true);
                content.querySelector(".id").href = containerUrl;
                content.querySelector(".id").innerText = shortId;
                content.querySelector(".name").innerText = container.Name;
                content.querySelector(".pid").innerText = container.State.Pid;
                content.querySelector(".pid").className += ' ' + status;
                content.querySelector(".started").innerText = started;
                content.querySelector(".finished").innerText = finished;
                content.querySelector(".restart-policy").innerText = restartPolicy;
                content.querySelector(".volumes-from").innerText = container.HostConfig.VolumesFrom;
                content.querySelector(".ports").innerHTML = ports;
                content.querySelector(".volumes").innerHTML = volumes;

                var containerRowsElement = document.getElementById("containerRows");
                containerRowsElement.appendChild(content);
            }

            function rePopulateContainersView(containers) {
                var containerRowsElement = document.getElementById("containerRows");
                while (containerRowsElement.firstChild) {
                    containerRowsElement.removeChild(containerRowsElement.firstChild);
                }

                for (var index = 0; index < containers.length; index++) {
                    addContainer(containers[index]);
                }
            }
            
            function getData(url, completionFunc, errorFunc) {
                var xhr = new XMLHttpRequest();
                xhr.onreadystatechange = function() {
                    if (xhr.readyState == 4) {
                        if(xhr.status == 200 || xhr.status == 201) {
                            data = JSON.parse(xhr.response);
                            completionFunc(data);
                        } else {
                            errorFunc(xhr.status);
                        }
                    }
                };
                xhr.open("GET", url);
                xhr.setRequestHeader('Accept', 'application/json');
                xhr.send();
            }

            function rePopulateViews() {
                getData(containersUrl, rePopulateContainersView, console.log);
            }
            
	    window.onload = function() {
                rePopulateViews();
               
                eventsSocket = new WebSocket(eventsUrl);
   
                eventsSocket.onopen = function() {
                    console.log("WebSocket: Connected to " + eventsSocket.url);
                }

                eventsSocket.onclose = function(e) {
                    console.log("WebSocket: Connection closed (" + e.code + ")");
                }
 
                eventsSocket.onerror = function(e) {
                    console.log("WebSocket: Connection error detected (" + e + ")");
                }

                eventsSocket.onmessage = function(e) {
                    console.log("WebSocket: Message received: " + e.data);
                    rePopulateViews(); // Lazy but plenty good here, don't have to deal with a lack of data and changed data
                }
            }
        </script>
    </head>
    <body>
        <h1>Containers</h1>
        <div class="table" id="containers">
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
        </div>
        <div id="containerRows">
	</div>
        <template id="containerTemplate">
            <div class="row">
                <div class="cell"><a href="" class="id"></a></div>
                <div class="cell name"></div>
                <div class="cell pid status"></div>
                <div class="cell started"></div>
                <div class="cell finished"></div>
                <div class="cell restart-policy"></div>
                <div class="cell ports"></div>
                <div class="cell volumes-from"></div>
                <div class="cell volumes"></div>
            </div>
        </template>
    </body>
</html>
`
