#!/bin/sh

set -e -x

# Install certstrap
go get -v github.com/square/certstrap

# Place keys and certificates here
depot_path="$HOME/.docker/consul-certs"
mkdir -p ${depot_path}

# CA to distribute to consul agent and servers
certstrap --depot-path ${depot_path} init --passphrase '' --common-name consulCA
mv -f ${depot_path}/consulCA.crt ${depot_path}/server-ca.crt
mv -f ${depot_path}/consulCA.key ${depot_path}/server-ca.key

# Server certificate to share across the consul cluster
server_cn=server.dc1.cf.internal
certstrap --depot-path ${depot_path} request-cert --passphrase '' --common-name $server_cn
certstrap --depot-path ${depot_path} sign $server_cn --CA server-ca
mv -f ${depot_path}/$server_cn.key ${depot_path}/server.key
mv -f ${depot_path}/$server_cn.csr ${depot_path}/server.csr
mv -f ${depot_path}/$server_cn.crt ${depot_path}/server.crt

# Agent certificate to distribute to jobs that access consul
certstrap --depot-path ${depot_path} request-cert --passphrase '' --common-name 'consul agent'
certstrap --depot-path ${depot_path} sign consul_agent --CA server-ca
mv -f ${depot_path}/consul_agent.key ${depot_path}/agent.key
mv -f ${depot_path}/consul_agent.csr ${depot_path}/agent.csr
mv -f ${depot_path}/consul_agent.crt ${depot_path}/agent.crt
