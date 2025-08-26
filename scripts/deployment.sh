#!/bin/bash
set -e

# Start Minikube if not running
if ! minikube status &> /dev/null; then
  echo "Starting Minikube..."
  minikube start --memory 10240 --cpus 4 --driver=docker
fi

# Enable Nginx Ingress
echo "Enabling Nginx Ingress..."
minikube addons enable ingress

docker build -f ./docker/Dockerfile -t kart-challenge:latest .
minikube image load kart-challenge:latest

# Mount local folder ./promocodes to /mnt/promocodes in Minikube
# Run this in the background
echo "Mounting local ./promocodes to /mnt/promocodes in Minikube..."
minikube mount ./promocodes:/mnt/promocodes &
MOUNT_PID1=$!
minikube mount ./db:/mnt/db &
MOUNT_PID2=$!

# Apply all Kubernetes resources via Kustomization
echo "Applying Kubernetes resources..."
kubectl apply -k ./deployment/k8s/

echo "Deployment complete! Access your app at http://kart.local/"

# Keep mount process alive
echo "Promocodes mount is running in background (PID: $MOUNT_PID1)"
echo "DB migrations mount is running in background (PID: $MOUNT_PID2)"

# --- Function to clean up on Ctrl+C ---
cleanup() {
    echo
    echo "Stopping all background processes..."
    kill $MOUNT_PID1 $MOUNT_PID2 2>/dev/null
    wait $MOUNT_PID1 $MOUNT_PID2 2>/dev/null
    echo "All mounts and tunnel stopped."
    exit 0
}

# Trap Ctrl+C (SIGINT) and call cleanup
trap cleanup SIGINT

echo "Press Ctrl+C to stop all mounts and the tunnel."
# Wait for all background processes to finish
wait $MOUNT_PID1 $MOUNT_PID2
