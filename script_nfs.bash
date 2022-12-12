#!/bin/bash -xe
# This script requires:
# - a working kubectl context pointing to your quicklab cluster and an admin user.
# - the quicklab ssh private key
# - the binary jq, helm, and oc
# - being connected to the Red Hat VPN
# The NFS provisioner will be deployed to the current namespace in the current cluster
( command -v jq && command -v helm && command -v oc) || ( echo "This script needs jq, helm and oc"; exit 1)
test -n "$1" || (echo "Please provide the quicklab private key as first argument" && exit 1)

host=upi-0.$(oc whoami --show-server | sed 's#https://api\.\([^:]*\):6443#\1#')
user=quicklab

# Configure the UPI host as an NFS server
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

# Grant permissions to the service account to hostmount-anyuid
oc adm policy add-scc-to-user hostmount-anyuid \
   system:serviceaccount:$(oc project -q):nfs-subdir-external-provisioner

# Install the NFS-subdir external provisioner
helm repo add nfs-subdir-external-provisioner \
   https://kubernetes-sigs.github.io/nfs-subdir-external-provisioner/
helm upgrade nfs-subdir-external-provisioner \
     nfs-subdir-external-provisioner/nfs-subdir-external-provisioner \
     --set nfs.server=$host \
     --set nfs.path=/srv/ocpstorage \
     --wait --atomic --install

# Make the NFS storageclass the default
oc annotate --overwrite storageclass nfs-client storageclass.kubernetes.io/is-default-class="true"
