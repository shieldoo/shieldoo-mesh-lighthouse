# Shieldoo Mesh Lighthouse

[![Build](https://github.com/shieldoo/shieldoo-mesh-lighthouse/actions/workflows/build-release.yml/badge.svg)](https://github.com/shieldoo/shieldoo-mesh-lighthouse/actions/workflows/build-release.yml) 
[![Release](https://img.shields.io/github/v/release/shieldoo/shieldoo-mesh-lighthouse?logo=GitHub&style=flat-square)](https://github.com/shieldoo/shieldoo-mesh-lighthouse/releases/latest) 
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=shieldoo_shieldoo-mesh-lighthouse&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=shieldoo_shieldoo-mesh-lighthouse) 
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=shieldoo_shieldoo-mesh-lighthouse&metric=bugs)](https://sonarcloud.io/summary/new_code?id=shieldoo_shieldoo-mesh-lighthouse) 
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=shieldoo_shieldoo-mesh-lighthouse&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=shieldoo_shieldoo-mesh-lighthouse)

The Shieldoo Mesh Lighthouse is a core component of the Shieldoo platform, specifically designed for the creation and management of Nebula-based networks. Built upon the robust foundations of the open-source Nebula project, the Lighthouse serves as the primary host tracker within a Managed Nebula network. 

## What is Lighthouse?

In a Nebula network, a Lighthouse is responsible for keeping track of all other hosts and aiding them in discovering each other within the network, regardless of their geographic location. By design, the Lighthouse is the only node within a Managed Nebula network whose IP address should remain constant.

Operating with minimal compute resources, the Lighthouse can easily be deployed using the most cost-effective options from cloud hosting providers. It is crucial to note that a UDP port (defaulted to 4242 but can be customized when setting up the Lighthouse) should be made accessible to the internet. This enables hosts to effectively communicate with the Lighthouse, ensuring the seamless operation of the network.

## Building Lighthouse

To build the Lighthouse, you need to run the following commands:

```bash
# Set environment for GOOS and GOARCH
env GOOS=linux GOARCH=amd64 go build -o out/shieldoo-mesh-lighthouse ./main

# Build the Docker image
docker build . --tag ghcr.io/shieldoo/shieldoo-mesh-lighthouse:latest

# Run the Docker container with test configuration
docker run -p 1053:53/udp --cap-add=NET_ADMIN -e DEBUG=true -e PUBLICIP=111.111.111.111 -e URI=http://192.168.1.133:9000/ -e SECRET=00008RCjReWTnn2a6p2ba4GVHc7wGweJKZuq8RCjReWTnn2a6p2ba4GVHc7wGweJKZuq0000 -e SENDINTERVAL=30 lh
```

The above commands first set the environment variables for the Go operating system (GOOS) and the architecture (GOARCH), then build the Docker image and finally run the Docker container with a test configuration. The container is run with a UDP port 1053 opened to allow communication. Several environment variables are passed in during the run command to configure the lighthouse.

For more advanced use cases and configurations, please refer to the Nebula documentation or contact the Shieldoo support team.