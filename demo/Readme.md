<!-- SPDX-License-Identifier: MIT -->

# demo

## `tcpdump`

1. run `make run-dpm` to get a debug container
2. start a `tcpdump` session in the container `tcpdump -i any dst port 9090`
3. start a port-forwarding session `kubectl port-forward svc/demo 9090:9090`
4. send a request to the service `curl localhost:9090`

