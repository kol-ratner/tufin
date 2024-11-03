# Tufin

A powerful CLI tool for deploying and managing WordPress and MySQL applications on Kubernetes.

## Overview

Tufin streamlines the deployment and configuration of WordPress and MySQL on Kubernetes clusters, offering flexible resource management and real-time monitoring capabilities.

## Installation

```
go install github.com/kol-ratner/tufin@latest
```

## Usage

### Manage Cluster
```
tufin cluster
```


### Deploy Applications
```
tufin deploy
```

Example deployment with custom configurations:

```
tufin deploy --set wordpress.replicas=2,wordpress.memory-request=1Gi,mysql.replicas=3
```

Available configuration options:
- replicas: Number of pod replicas (int)
- cpu-request: Minimum CPU guaranteed (e.g., 250m, 500m)
- memory-request: Minimum memory guaranteed (e.g., 256Mi, 1Gi)
- cpu-limit: Maximum CPU allowed (e.g., 500m, 1)
- memory-limit: Maximum memory allowed (e.g., 512Mi, 2Gi)
- volume-size: Persistent volume size (e.g., 5Gi, 10Gi)


### Monitor Status
```
tufin status
```


## Contributing
We welcome contributions! Please submit pull requests for any enhancements.


## License
MIT License - see LICENSE for details.

