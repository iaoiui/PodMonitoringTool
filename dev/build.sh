#go build -o main app/*
docker build . -t pod_monitoring_tool 
kind load docker-image pod_monitoring_tool:latest
