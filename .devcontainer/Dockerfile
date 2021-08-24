# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.191.1/containers/go/.devcontainer/base.Dockerfile

# [Choice] Go version: 1, 1.16, 1.15
ARG VARIANT="1.16"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Option] Enable non-root Docker access in container
ARG ENABLE_NONROOT_DOCKER="true"
# [Option] Use the OSS Moby CLI instead of the licensed Docker CLI
ARG USE_MOBY="false"

# [Optional] Uncomment this section to install additional OS packages.
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends bash-completion

# Install operator-sdk
ARG OPERATOR_SDK_VERSION="1.11.0"
ARG OPERATOR_SDK_KEY_SERVER="keyserver.ubuntu.com"
ARG OPERATOR_SDK_GPG_KEY="052996E2A20B5C7E"
COPY library-scripts/install-operator-sdk.sh /tmp/library-scripts/install-operator-sdk.sh
RUN /tmp/library-scripts/install-operator-sdk.sh "${OPERATOR_SDK_VERSION}" "${OPERATOR_SDK_KEY_SERVER}" "${OPERATOR_SDK_GPG_KEY}"

# Install kubectl
RUN curl -sSL -o /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && kubectl completion bash > /etc/bash_completion.d/kubectl

# Replace localhost with host.docker.internal
COPY library-scripts/copy-kube-config.sh /usr/local/share/
RUN echo "source /usr/local/share/copy-kube-config.sh" | tee -a /home/vscode/.bashrc >> /home/vscode/.zshrc

# Configure for Docker access
COPY library-scripts/docker-debian.sh /tmp/library-scripts/
RUN apt-get update \
    && bash /tmp/library-scripts/docker-debian.sh "${ENABLE_NONROOT_DOCKER}" "/var/run/docker-host.sock" "/var/run/docker.sock" "${USERNAME}" "${USE_MOBY}"
ENTRYPOINT ["/usr/local/share/docker-init.sh"]
CMD ["sleep", "infinity"]
