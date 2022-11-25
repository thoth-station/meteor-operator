#!/bin/bash -xe
( command -v jq && command -v helm && command -v oc) || ( echo "This script needs jq, helm and oc"; exit 1)
test -n "$1" || (echo "Please provide the quicklab private key as first argument" && exit 1)

host=upi-0.$(kubectl config view -o jsonpath='{.clusters[0].cluster.server}' | sed 's#https://api\.\([^:]*\):6443#\1#')
user=quicklab

ssh -i "$1" $user@$host sudo bash -exs <<'ENDSSH'
yum -y install nfs-utils
systemctl enable --now rpcbind
# RHEL 7 || RHEL 8:
systemctl enable --now nfs || systemctl enable --now nfs-server
mkdir -p /srv/ocpstorage
chmod 770 /srv/ocpstorage

export="/srv/ocpstorage 10.0.0.0/8(rw,sync,no_root_squash)"
grep -q "$export" /etc/exports || echo "$export" >> /etc/exports

exportfs -ra

firewall-cmd --permanent --add-service mountd
firewall-cmd --permanent --add-service rpc-bind
firewall-cmd --permanent --add-service nfs
firewall-cmd --reload
ENDSSH

helm repo add nfs-subdir-external-provisioner \
   https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/
helm upgrade nfs-subdir-external-provisioner \
     nfs-subdir-external-provisioner/nfs-subdir-external-provisioner \
     --set nfs.server=$host \
     --set nfs.path=/srv/ocpstorage \
     --wait --atomic --install

oc adm policy add-scc-to-user hostmount-anyuid \
   system:serviceaccount:aicoe-meteor:nfs-subdir-external-provisioner
oc annotate --overwrite storageclass nfs-client storageclass.kubernetes.io/is-default-class="true"
