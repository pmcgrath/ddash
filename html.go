package main

import (
	"html/template"
)

var (
	rootTemplate *template.Template
)

func init() {
	rootTemplate = template.Must(template.New("root").Parse(rootHTMLTemplate))
}

// Could have used https://github.com/jteeuwen/go-bindata
const rootHTMLTemplate = `
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
            var scheme = "http", wsScheme = "ws";
            if (window.location.protocol == "https") {
                scheme = "https"; wsScheme = "wss";
            }

            var containerUrlPrefix = scheme + "://" + window.location.host + "{{.ContainerPathPrefix}}";
            var containersUrl = scheme + "://" + window.location.host + "{{.ContainersPath}}";
            var eventsUrl = wsScheme + "://" + window.location.host + "{{.SocketPath}}";

            var eventsSocket = null;
            var eventsSocketRetryIntervalInMilliseconds = 2000;
            var eventsSocketRetryMaxIntervalInMilliseconds = 5 * 60 * 1000; // 5 minutes
            var eventsSocketRetryAttempts = 0;
            var eventsSocketRetryAttempt = 0;

            function getContainerStatus(container) {
                if (!container.State.Running) { return "stopped"; }
                if (container.State.Paused) { return "paused"; }

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
                var name = container.Name.substring(1, container.Name.length -1); 
                var containerUrl = containerUrlPrefix + container.Id;
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
                    if (portSetting != null) { ports += portSetting[0].HostPort + "->"; }
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
                content.querySelector(".id").textContent = shortId;
                content.querySelector(".name").textContent = name;
                content.querySelector(".pid").textContent = container.State.Pid;
                content.querySelector(".pid").className += ' ' + status;
                content.querySelector(".started").textContent = started;
                content.querySelector(".finished").textContent = finished;
                content.querySelector(".restart-policy").textContent = restartPolicy;
                content.querySelector(".volumes-from").textContent = container.HostConfig.VolumesFrom;
                content.querySelector(".ports").innerHTML = ports;
                content.querySelector(".volumes").innerHTML = volumes;

                containersElement.appendChild(content);
            }

            function setConnectionStatus(connected) {
                var connectionStatusElement = document.getElementById("connectionStatus");
                connectionStatusElement.textContent = connected;
            }

            function rePopulateContainersView(containers) {
                var containersElement = document.getElementById("containers");

                while (containersElement.childElementCount > 1) {
                    containersElement.removeChild(containersElement.lastElementChild);
                }

                for (var index = 0; index < containers.length; index++) {
                    addContainer(containers[index]);
                }
                
                var lastPopulationElement = document.getElementById("lastPopulation");
                lastPopulationElement.textContent = new Date();
            }
            
            function getData(url, completionFunc, errorFunc) {
                var xhr = new XMLHttpRequest();
                xhr.onreadystatechange = function() {
                    if (xhr.readyState == 4) {
                        if(xhr.status == 200 || xhr.status == 201) {
                            var data = JSON.parse(xhr.response);
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

            function configureEventsSocket() {
                console.log("Attempting to configure events socket for " + eventsUrl + " sequence " + eventsSocketRetryAttempts);
                eventsSocket = new WebSocket(eventsUrl);
   
                eventsSocket.onopen = function() {
                    console.log("WebSocket: Connected to " + eventsSocket.url);
                    setConnectionStatus(true);
                    if (eventsSocketRetryAttempts > 0) {
                        console.log("Repopulating views due to socket connection being reopened, attempt sequence " + eventsSocketRetryAttempts);
                        rePopulateViews();
                        eventsSocketRetryAttempts = 0;
                    }
                }

                eventsSocket.onclose = function(e) {
                    console.log("WebSocket: Connection closed (" + e.code + ")");
                    eventsSocket.onopen = null;
                    eventsSocket.onclose = null;
                    eventsSocket.onerror = null;
                    eventsSocket.onmessage = null;
                    eventsSocket = null;
                    setConnectionStatus(false);

                    // Try to re-establish the connection after interval
                    eventsSocketRetryAttempts++;
                    var retryInterval = eventsSocketRetryIntervalInMilliseconds * eventsSocketRetryAttempts;
                    if (retryInterval > eventsSocketRetryMaxIntervalInMilliseconds) {
                        retryInterval = eventsSocketRetryMaxIntervalInMilliseconds;
                    }
                    window.setTimeout(configureEventsSocket, retryInterval);
                }
 
                eventsSocket.onerror = function(e) {
                    console.log("WebSocket: Connection error detected (" + e + ")");
                }

                eventsSocket.onmessage = function(e) {
                    console.log("WebSocket: Message received: " + e.data);
                    rePopulateViews(); // Lazy but plenty good here, don't have to deal with a lack of data and changed data
                }
            }

            window.onload = function() {
                rePopulateViews();
                configureEventsSocket();
            }
        </script>
    </head>
    <body>
        <div id="lastPopulation"></div>
        <div id="connectionStatus"></div>
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
