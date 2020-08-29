kubectl delete secret envvar
kubectl create secret generic --from-env-file .env envvar 
