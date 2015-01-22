# Docker dashboard

## Simple docker dashboard for viewing docker containers
- Use the docker cli, this is just something I used to explore the docker api
- Web app that listens on port 8090, showing a listing of containers (Probably should add images sometime)
- App uses no package to interface with docker, just plain http access, so no models exist
- App only does reads, surfaces no modification functionality
- App listens for docker events
- Clients establish a web socket connection to get docker update events
- Clients reload the docker information on receiving a docker event

## Build and run options
- To build use : ./build.sh build
- To build docker image use : ./build.sh image
- To run the app locally use : ./dash
- To run the app within a container use : sudo docker run --name=ddash --detach=true --volume=/var/run/docker.sock:/var/run/docker.sock:ro --publish=8090:8090 ${USER}/ddash:latest

## Not completed
- UI is terrible
 - Use AWS EC2 style table at top with selected container info appearing at the bottom of the view 
 - Display full container info on an inline panel rather than showing in a seperate page (Would reduce web socket connections)
- Could alter so it supported multiple docker hosts rather than just the one on the app's host 
- Support TLS connections (app and docker itself)
- Include docker images
- Filter docker events so we can ignore some (i.e. Docker kill and stop events)

