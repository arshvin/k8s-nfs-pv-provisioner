#!/usr/bin/env bash
set -eo pipefail

if [[ ! -d /home/vagrant ]]; then
    echo "ERROR: do not run this script directly!"
    echo "script should be run as part of vagrant provision process"
    exit 1
fi

# check we have an argument which points to to project root
if [[ -z $1 ]]; then
    echo "ERROR: project root argument is not provided."
    echo "Usage: $0 /path/to/project"
    exit 1
fi

PROJECT_ROOT=$1
if [[ ! -d ${PROJECT_ROOT} ]]; then
    echo "ERROR: provided project root is not a directory!"
    exit 1
fi

# check we have ansible-playbook installed
if [[ -z $(which ansible-playbook) ]]; then
    echo "Ansible is not installed, performing installation..."
    yum -q clean all
    yum install -y -q epel-release ansible
    yum install python-pip -y --nogpgcheck
    #upgrade ansible 2.7 to latest version by pip
    python -m pip install --upgrade ansible    
fi

# vagrant box provision
env PYTHONUNBUFFERED=1 ANSIBLE_HOST_KEY_CHECKING=0 ANSIBLE_FORCE_COLOR=1 ANSIBLE_ROLES_PATH=${PROJECT_ROOT}/vm_provisionnig/roles \
        ansible-playbook \
        "${PROJECT_ROOT}/vm_provisioning/vagrant.yml"
