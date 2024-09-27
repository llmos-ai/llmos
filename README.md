# LLMOS
[![main-build](https://github.com/llmos-ai/llmos/actions/workflows/main-release.yaml/badge.svg)](https://github.com/llmos-ai/llmos/actions/workflows/main-release.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/llmos-ai/llmos)](https://goreportcard.com/report/github.com/llmos-ai/llmos)
[![Releases](https://img.shields.io/github/release/llmos-ai/llmos.svg)](https://github.com/llmos-ai/llmos/releases)

[LLMOS](https://llmos.1block.ai/) is an open-source, cloud-native infrastructure software tailored for managing AI applications and Large Language Models(LLMs).

## Key Features
- **Easy to Install:** Install directly on the x86_64 or ARM64 architecture, offering an out-of-the-box user experience.
- **Complete Infrastructure & LLM Lifecycle Management:** Provides a unified interface for both developers and non-developers to manage the LLM infrastructure, ML Cluster, models and workloads.
- **Easy to Use:** Build models and AI applications in your own way, without needing to managing Kubernetes & infrastructure directly.
- **Perfect for Edge & Branch:** Better resource optimization, simplify the deployment of models and workloads to edge and branch networks, but can also scale up horizontally to handle large workloads.

## Quick Start

### Installation Script

LLMOS can be installed to a bare-metal server or a virtual machine. To bootstrap a **new cluster**, follow the steps below:

```shell
curl -sfL https://get-llmos.1block.ai | sh -s - --cluster-init --token mytoken
```

To watch the installation logs, run `journalctl -u llmos -f`.

After the installation completes, it is optional to add a additional worker node to the cluster with the following command:
```shell
curl -sfL https://get-llmos.1block.ai | LLMOS_SERVER=https://server-url:6443 LLMOS_TOKEN=mytoken sh -s -
```

### Config Proxy
If you environment needs to access the internet through a proxy, you can set the `HTTP_PROXY` and `HTTPS_PROXY` environment variables to configure the installation script to use the proxy.

```shell
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16 # Replace the CIDRs with your own
```

## Getting Started

After installing LLMOS, you can access the dashboard by navigating to `https://<server-ip>:8443` in your web browser.

1. LLMOS will bootstrap a default admin user with the username `admin` and a random password. To get the password, you can run the following command on the **cluster-init** node:
    ```shell
    kubectl get secret --namespace llmos-system llmos-bootstrap-passwd -o go-template='{{.data.password|base64decode}}{{"\n"}}'
    ```
   ![first-login](./assets/docs/auth-first-login.png)
1. After logging in, you will be redirected to the setup page, you will need to configure the following:
    - Set a **new password** for the admin user, using strong passwords is recommended.
    - Config the **server URL** where all other nodes in your cluster will be able to reach this.
      ![setup](./assets/docs/auth-first-login-setup.png)
1. After that, you will be redirected to the home page where you can start using LLMOS.
   ![home-page](./assets/docs/home-page.png)

## More Examples

To learn more about how to use LLMOS, check out the examples below:

- [Chat with LLMOS Models](https://llmos.1block.ai/docs/user_guide/llm_management/serve/)
- [Creating a Machine Learning Cluster](https://llmos.1block.ai/docs/user_guide/ml_clusters)
- [Creating a Jupyter Notebook](https://llmos.1block.ai/docs/user_guide/llm_management/notebooks/#create-a-notebook)

## Documentation
Find more documentation [here](https://llmos.1block.ai/docs/).

## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

