
# Boxy

<div align="center">
	<img src="docs/boxy.png" alt="Boxy Logo" width="220"/>
	<br/>
	<img src="https://img.shields.io/badge/Go-1.25.5-blue?logo=go" alt="Golang"/>
	<img src="https://img.shields.io/badge/License-GPLv3-blue.svg" alt="GPLv3 License"/>
</div>



Boxy is a lightweight containerization system built to run isolated applications using core Linux features. It uses namespaces, cgroups, and filesystem isolation to create independent environments. Designed for simplicity and learning, Boxy demonstrates how container runtimes manage processes, resources, and system isolation.

## Makefile

```makefile
# build CLI (default)
make build

# build daemon
make build TARGET=daemon

# install daemon to GOBIN
make install TARGET=daemon

# convenience
make build-daemon
make install-cli
```
