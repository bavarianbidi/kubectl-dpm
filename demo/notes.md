```
k debug --image nicolaka/netshoot:v0.13 prometheus-server-9d47b6bf5-gvbrp -it
```

openssl s_client -verify_return_error -connect kubernetes.default.svc:443

k debug --image nicolaka/netshoot:v0.13 prometheus-server-9d47b6bf5-gvbrp -it -- openssl s_client -verify_return_error -connect kubernetes.default.svc:443


ok, something with certs, but we know, k8s is injecting the certs into the pods, so we need to check the certs in the pod
but they are not there in our debug container

...
... this is where the debug profiles came in


openssl s_client -verifyCAfile /run/secrets/kubernetes.io/serviceaccount/ca.crt -connect kubernetes.default.svc:443

openssl s_client -verifyCAfile /run/secrets/kubernetes.io/serviceaccount/ca.crt -connect kubernetes.default.svc:443 -brief -no-interactive
openssl s_client -connect kubernetes.default.svc:443 -brief -no-interactive


k dpm run --profile=prometheus-kcd