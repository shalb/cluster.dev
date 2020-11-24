# Info

Logging stack based on [chart](https://github.com/helm/charts/tree/master/stable/graylog)

# Access

## Web UI

Create a proxy connection to your machine

Default credentials:  
User: `admin`  
Password: `admin`


```
export POD_NAME=$(kubectl get pods -n monitoring -l "app=graylog" -o jsonpath="{.items[0].metadata.name}")
kubectl -n monitoring port-forward --address 127.0.0.1 $POD_NAME 19000:9000
```

Access it via http://127.0.0.1:19000/

