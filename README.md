# Install requirements

## Install docker
Follow instructions [docker ubuntu](https://docs.docker.com/engine/install/ubuntu/)
Post instructions
````bash
sudo groupadd docker
sudo usermod -aG docker $USER
newgrp docker
````

## Install Go 
Follow instructions [Go](https://go.dev/doc/install)
[Download](https://go.dev/dl/) the corresponding package for your OS.

```bash
 rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz
```

Update your $HOME/.bashrc
```bash
cat <<EOF >> $HOME/.bashrc

# set PATH so it includes go bin if it exists
if [ -d "/usr/local/go/bin" ] ; then
    PATH="/usr/local/go/bin:$PATH"
fi
EOF
source $HOME/.bashrc
```

## Install crictl
Follow instructions [crictl](https://github.com/kubernetes-sigs/cri-tools/blob/master/docs/crictl.md0)
```bash
VERSION="v1.30.1" # check latest version in /releases page
curl -L https://github.com/kubernetes-sigs/cri-tools/releases/download/$VERSION/crictl-${VERSION}-linux-amd64.tar.gz --output crictl-${VERSION}-linux-amd64.tar.gz
sudo tar zxvf crictl-$VERSION-linux-amd64.tar.gz -C /usr/local/bin
rm -f crictl-$VERSION-linux-amd64.tar.gz
```

## Install cri-dockerd
[Download](https://github.com/Mirantis/cri-dockerd/releases) the corresponding package for your OS.
```bash
sudo dpkg -i cri-dockerd_0.3.15.3-0.ubuntu-jammy_amd64.deb 
```

## Install conntrack
```bash
sudo apt install conntrack
```

## Install kubebuilder

```bash
# download kubebuilder and install locally.
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)"
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```

## Install 
Follow instructions [container-networking](https://minikube.sigs.k8s.io/docs/faq/#how-do-i-install-containernetworking-plugins-for-none-driver)
```bash
CNI_PLUGIN_VERSION="v1.5.1"
CNI_PLUGIN_TAR="cni-plugins-linux-amd64-$CNI_PLUGIN_VERSION.tgz" # change arch if not on amd64
CNI_PLUGIN_INSTALL_DIR="/opt/cni/bin"

curl -LO "https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGIN_VERSION/$CNI_PLUGIN_TAR"
sudo mkdir -p "$CNI_PLUGIN_INSTALL_DIR"
sudo tar -xf "$CNI_PLUGIN_TAR" -C "$CNI_PLUGIN_INSTALL_DIR"
rm "$CNI_PLUGIN_TAR"
```


## Install minikube
Follow instructions [minikube](https://kubernetes.io/fr/docs/tasks/tools/install-minikube/)
```bash
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && chmod +x minikube
sudo mkdir -p /usr/local/bin/
sudo install minikube /usr/local/bin/
```

## Setup command completion
```bash
cat <<EOF >> $HOME/.bashrc
# Auto completion kubebuilder
if command -v kubebuilder &> /dev/null
then
    eval $(kubebuilder completion bash)
fi
Create
# Auto completion kubectl
if command -v kubectl &> /dev/null
then
    eval $(kubectl completion bash)
fi

# Auto completion minikube
if command -v minikube &> /dev/null
then
    eval $(minikube completion bash)
fi

# Auto completion crictl
if command -v crictl &> /dev/null
then
    eval $(crictl completion bash)
fi
EOF
```

# Project cninf

## Create and init project

````bash
mkdir -p $HOME/Workspace/cninf
kubebuilder init --domain uman.test --repo github.com/cedvict/cninf --plugins=go/v4
````

## Create api

````bash
kubebuilder create api --group cninf --version v1 --kind Store --resource --controller
````