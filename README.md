# LLMOS
[![main-build](https://github.com/llmos-ai/llmos/actions/workflows/main-release.yaml/badge.svg)](https://github.com/llmos-ai/llmos/actions/workflows/main-release.yaml)
[![Releases](https://img.shields.io/github/release/llmos-ai/llmos.svg)](https://github.com/llmos-ai/llmos/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/llmos-ai/llmos)](https://goreportcard.com/report/github.com/llmos-ai/llmos)
[![Discord](https://img.shields.io/discord/1178957864300191754?logo=discord&label=discord)](https://discord.gg/5BnNqC5ccB)

[LLMOS](https://llmos.1block.ai/) is an open-source, cloud-native infrastructure software tailored for managing AI applications and Large Language Models(LLMs).

## Key Features

- **Easy [Installation](https://llmos.1block.ai/docs/installation/):** Simple to install on both x86_64 and ARM64 architectures, delivering an out-of-the-box user experience.
- **Seamless [Notebook](https://llmos.1block.ai/docs/user_guide/llm_management/notebooks) Integration:** Integrates with popular notebook environments such as **Jupyter**, **VSCode**, and **RStudio**, allowing data scientists and developers to work efficiently in familiar tools without complex setup.
- **[ModelService](https://llmos.1block.ai/docs/user_guide/llm_management/serve) for LLM Serving:** Easily serve LLMs using ModelService with **OpenAI-compatible APIs**.
- **[Machine Learning Cluster](https://llmos.1block.ai/docs/user_guide/ml_clusters):** Supports distributed computing with parallel processing capabilities and access to leading AI libraries, improving the performance of machine learning workflows—especially for large-scale models and datasets.
- **Built-in [Distributed Storage](https://llmos.1block.ai/docs/user_guide/storage/system-storage):** Provides built-in distributed storage with high-performance, fault-tolerant features. Offers robust, scalable block and filesystem storage tailored to the demands of AI and LLM applications.
- **[User](https://llmos.1block.ai/docs/user_and_auth/user) & [RBAC Management](https://llmos.1block.ai/docs/user_and_auth/role-template):** Simplifies user management with role-based access control (RBAC) and role templates, ensuring secure and efficient resource allocation.
- **Optimized for Edge & Branch Deployments:** Supports private deployments with optimized resource usage for running models and workloads in edge and branch networks. It also allows for horizontal scaling to accommodate future business needs.


## Use Cases

- **AI Research & Development:** Simplifies LLM and AI infrastructure management, enabling researchers to focus on innovation rather than operational complexities.
- **Enterprise AI Solutions:** Streamline the deployment of AI applications with scalable infrastructure, making it easier to manage models, storage, and resources across multiple teams.
- **Data Science Workflows:** With notebook integration and powerful cluster computing, LLMOS is ideal for data scientists looking to run complex experiments at scale.
- **AI-Driven Products:** From chatbots to automated content generation, LLMOS simplifies the process of deploying LLM-based products that can serve millions of users and scale up horizontally.


## Quick Start

Make sure your nodes meet the [requirements](https://llmos.1block.ai/docs/installation/requirements) before proceeding.

### Installation Script

LLMOS can be installed to a bare-metal server or a virtual machine. To bootstrap a **new cluster**, follow the steps below:

```shell
curl -sfL https://get-llmos.1block.ai | sh -s - --cluster-init --token mytoken
```

To monitor installation logs, run `journalctl -u llmos -f`.

After installation, you may optionally add a worker node to the cluster with the following command:
```shell
curl -sfL https://get-llmos.1block.ai | LLMOS_SERVER=https://server-url:6443 LLMOS_TOKEN=mytoken sh -s -
```

### Config Proxy
If your environment requires internet access through a proxy, set the `HTTP_PROXY` and `HTTPS_PROXY` environment variables before running the installation script:

```shell
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=127.0.0.0/8,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16 # Replace the CIDRs with your own
```

## Getting Started

After installing LLMOS, access the dashboard by navigating to `https://<server-ip>:8443` in your web browser.

1. LLMOS will create a default `admin` user with a randomly generated password. To retrieve the password, run the following command on the **cluster-init** node:
    ```shell
    kubectl get secret --namespace llmos-system llmos-bootstrap-passwd -o go-template='{{.data.password|base64decode}}{{"\n"}}'
    ```
   ![first-login](./assets/docs/auth-first-login.png)
1. Upon logging in, you will be redirected to the setup page. Configure the following:
    - Set a **new password** for the admin user (strong passwords are recommended).
    - Configure the **server URL** that all other nodes in your cluster will use to connect.
      ![setup](./assets/docs/auth-first-login-setup.png)
1. After setup, you will be redirected to the home page where you can start using LLMOS.
   ![home-page](./assets/docs/home-page.png)

## More Examples

To learn more about using LLMOS, explore the following resources:
- [Chat with LLMOS Models](https://llmos.1block.ai/docs/user_guide/llm_management/serve/)
- [Creating a Machine Learning Cluster](https://llmos.1block.ai/docs/user_guide/ml_clusters)
- [Creating a Jupyter Notebook](https://llmos.1block.ai/docs/user_guide/llm_management/notebooks/#create-a-notebook)

## Documentation
Find more detailed documentation, visit [here](https://llmos.1block.ai/docs/).

## Community
If you're interested, please join us on [Discord](https://discord.gg/5BnNqC5ccB) or participate in [GitHub Discussions](https://github.com/llmos-ai/llmos/discussions) to discuss or contribute the project. We look forward to collaborating with you!

If you have any feedback or issues, feel free to file a GitHub [issue](https://github.com/llmos-ai/llmos/issues).

## License

Copyright (c) 2024 [1Block.AI.](https://1block.ai/)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

