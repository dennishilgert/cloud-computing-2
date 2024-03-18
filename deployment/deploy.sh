#!/bin/sh

echo "Starting Deployment process"

# Check if both files exist
echo "Checking prerequisites ..."
if [ ! -f "../service-account.json" ] || [ ! -f "./main.tf" ]; then
    echo "Please make sure the following files exist:"
    echo "deployment/main.tf"
    echo "service-account.json"
    echo "Both can be created from the respective samples."
fi

# Remove sample terraform file
echo "Removing sample Terraform file ..."
rm main.sample.tf

# Initializing Terraform
echo "Initializing Terraform ..."
terraform init

# Apply Terraform configuration
echo "Applying Terraform configuration (this can take up to 6 minutes) ..."
terraform apply -auto-approve

# Get VM IP and Redis IP
echo "Reading output variables of Terraform execution ..."
VM_IP=$(terraform output -raw vm_ip)
REDIS_IP=$(terraform output -raw redis_ip)

# Create Ansible Inventory
echo "Creating dynamic Ansible inventory ..."
echo "[gcp_vm]" > ansible_inventory
echo "$VM_IP ansible_user=dennis ansible_ssh_private_key_file=~/.ssh/id_rsa" >> ansible_inventory

# Run Ansible Playbook
echo "Running Ansible playbook (this can take up to 20 minutes) ..."
ansible-playbook -i ansible_inventory playbook.yml -e redis_host=$REDIS_IP

echo "Deployed."