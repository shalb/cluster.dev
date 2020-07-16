# Info

Monitoring stack based on [chart](https://github.com/helm/charts/tree/master/stable/prometheus-operator)

See application here `kubernetes/apps/monitoring.yaml`

# Access

## Grafana

Create a proxy connection to your machine

```
export POD_NAME=$(kubectl get pods -n monitoring -l "app=prometheus" -o jsonpath="{.items[0].metadata.name}")
kubectl -n monitoring port-forward --address 127.0.0.1 $POD_NAME 19090:9090
```

Access it via http://127.0.0.1:13000/

# Prometheus

Create a proxy connection to your machine

```
POD_NAME=$(kubectl get pods -n monitoring -l "app.kubernetes.io/name=grafana" -o jsonpath="{.items[0].metadata.name}
kubectl -n monitoring port-forward --address 127.0.0.1 $POD_NAME 13000:3000
```

Access it via http://127.0.0.1:19090/

