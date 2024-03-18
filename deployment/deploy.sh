#!/bin/sh

# Apply Terraform Configuration
terraform init && terraform apply -auto-approve

# Get VM IP and Redis IP
VM_IP=$(terraform output -raw vm_ip)
REDIS_IP=$(terraform output -raw redis_ip)

# Create Ansible Inventory
echo "[gcp_vm]" > ansible_inventory
echo "$VM_IP ansible_user=dennis ansible_ssh_private_key_file=~/.ssh/id_rsa" >> ansible_inventory

# Run Ansible Playbook
ansible-playbook -i ansible_inventory playbook.yml -e redis_host=$REDIS_IP