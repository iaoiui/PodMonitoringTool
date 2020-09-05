kubectl delete secret envvar -n pod-monitoring
kubectl create secret generic --from-env-file .env envvar -n pod-monitoring
