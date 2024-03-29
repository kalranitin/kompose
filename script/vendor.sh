#!/usr/bin/env bash
set -e

cd "$(dirname "$BASH_SOURCE")/.."
rm -rf vendor/
source 'script/.vendor-helpers.sh'

clone git github.com/Sirupsen/logrus v0.8.2
clone git github.com/codegangsta/cli 6086d7927ec35315964d9fea46df6c04e6d697c1
clone git github.com/docker/distribution c6c9194e9c6097f84b0ff468a741086ff7704aa3
clone git github.com/docker/docker 58b270c338e831ac6668a29788c72d202f9fc251
clone git github.com/docker/libtrust 9cbd2a1374f46905c68a4eb3694a130610adc62a
clone git github.com/flynn/go-shlex 3f9db97f856818214da2e1057f8ad84803971cff
clone git github.com/fsouza/go-dockerclient 39d9fefa6a7fd4ef5a4a02c5f566cb83b73c7293
clone git github.com/gorilla/context 215affda49addc4c8ef7e2534915df2c8c35c6cd
clone git github.com/gorilla/mux f15e0c49460fd49eebe2bcc8486b05d1bef68d3a
clone git github.com/opencontainers/runc b40c7901845dcec5950ecb37cb9de178fc2c0604
clone git github.com/stretchr/testify a1f97990ddc16022ec7610326dd9bce31332c116
clone git github.com/davecgh/go-spew 5215b55f46b2b919f50a1df0eaa5886afe4e3b3d
clone git github.com/pmezard/go-difflib d8ed2627bdf02c080bf22230dbb337003b7aba2d
clone git golang.org/x/crypto 4d48e5fa3d62b5e6e71260571bf76c767198ca02 https://github.com/golang/crypto.git
clone git golang.org/x/net 3a29182c25eeabbaaf94daaeecbc7823d86261e7 https://github.com/golang/net.git
clone git gopkg.in/check.v1 11d3bc7aa68e238947792f30573146a3231fc0f1
clone git github.com/Azure/go-ansiterm 70b2c90b260171e829f1ebd7c17f600c11858dbe
clone git github.com/cloudfoundry-incubator/candiedyaml 55a459c2d9da2b078f0725e5fb324823b2c71702

# These are required for the kubernetes dependencies and will need to be kept in sync with
# github.com/kubernetes/kubernetes/Godeps/Godeps.json
clone git bitbucket.org/bertimus9/systemstat 1468fd0db20598383c9393cccaa547de6ad99e5e
clone hg bitbucket.org/ww/goautoneg 75cd24fc2f2c2a2088577d12123ddee5f54e0675
clone git github.com/beorn7/perks b965b613227fddccbfffe13eae360ed3fa822f8d
clone git github.com/blang/semver 31b736133b98f26d5e078ec9eb591666edfd091f
clone git github.com/docker/libcontainer 5dc7ba0f24332273461e45bc49edcb4d5aa6c44c # Replace with runc/libcontainer
clone git github.com/google/cadvisor a8085bf9276c22f16dbcd7aa56f0d4d0626a0b2e
clone git github.com/emicklei/go-restful 1f9a0ee00ff93717a275e15b30cf7df356255877
clone git github.com/ghodss/yaml 73d445a93680fa1a78ae23a5839bad48f32ba1ee
clone git github.com/golang/glog 44145f04b68cf362d9c4df2182967c2275eaefed
clone git github.com/golang/protobuf 7f07925444bb51fa4cf9dfe6f7661876f8852275
clone git github.com/matttproud/golang_protobuf_extensions fc2b8d3a73c4867e51861bbdd5ae3c1f0869dd6a
clone git github.com/pborman/uuid ca53cad383cad2479bbba7f7a1a05797ec1386e4
clone git github.com/juju/ratelimit 772f5c38e468398c4511514f4f6aa9a4185bc0a0
clone git github.com/google/gofuzz bbcb9da2d746f8bdbd6a936686a0a6067ada0ec5
clone git github.com/prometheus/client_golang 3b78d7a77f51ccbc364d4bc170920153022cfd08
clone git github.com/prometheus/client_model fa8ad6fec33561be4280a8f0514318c79d7f6cb6
clone git github.com/prometheus/common ef7a9a5fb138aa5d3a19988537606226869a0390
clone git github.com/prometheus/procfs 490cc6eb5fa45bf8a8b7b73c8bc82a8160e8531d
clone git github.com/spf13/pflag 8e7dc108ab3a1ab6ce6d922bbaff5657b88e8e49
clone git speter.net/go/exp/math/dec/inf 42ca6cd68aa922bc3f32f1e056e61b65945d9ad7 https://code.google.com/p/go-decimal-inf.exp
clone git github.com/ugorji/go/codec f1f1a805ed361a0e078bb537e4ea78cd37dcf065 https://github.com/ugorji/go.git
clone git github.com/opencontainers/runc 97bc9a7faf3dd660d9be90a2880b2e37f3cdbf38
clone git gopkg.in/yaml.v2 d466437aa4adc35830964cffc5b5f262c63ddcb4
clone git k8s.io/kubernetes v1.1.4 https://github.com/kubernetes/kubernetes.git

clean && mv vendor/src/* vendor
