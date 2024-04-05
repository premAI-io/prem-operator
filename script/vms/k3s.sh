#!/bin/sh -eu

set_vars() {
  echo "QEMU K3s Ports(SSH=${K3S_SSH_PORT:=2222} API=${K3S_API_PORT:=16443} web=${K3S_WEB_PORT:=8080})"
  echo "Kubeconfig: ${KUBECONFIG:=kubeconfig}"
  echo "Image: ${IMG:=container.tar}"

  rund=${XDG_RUNTIME_DIR-/tmp/$USER}/prem-operator
  pid_file="${rund}/qemu.pid"
  here=$(dirname "$0")

  echo "Runtime dir: $rund"
}

# Function for 'boot-qemu-daemon' subcommand
boot_qemu_daemon() {
  echo "Booting QEMU as a daemon..."

  mkdir -p $rund

  if [ -f "$pid_file" ] && ps -p $(cat "$pid_file") > /dev/null 2>&1; then
      echo "QEMU is already running. Exiting."
      exit 1
  fi

  if [ $# -lt 1 ]; then
    QCOW=temp.img
    rm -f temp.img
    qemu-img create -F qcow2 -f qcow2 -b k3s-base.img temp.img
  else
    QCOW=$1
  fi

  qemu-system-x86_64 \
    -machine accel=kvm,type=q35 \
    -cpu host \
    -m 8G \
    -device virtio-net-pci,netdev=net0 \
    -netdev user,id=net0,hostfwd=tcp::${K3S_SSH_PORT}-:22,hostfwd=tcp::${K3S_API_PORT}-:6443,hostfwd=tcp::${K3S_WEB_PORT}-:80 \
    -drive if=virtio,format=qcow2,file=$QCOW,cache=none \
    -drive if=virtio,format=qcow2,file=seed.img \
    -serial file:serial.log \
    -pidfile "$pid_file" \
    -display none \
    -daemonize

  echo "Started QEMU waiting for SSH"
  local ssh_opts="-o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=5"
  local timeout=120
  local start_time=$(date +%s)

  while ! ssh $ssh_opts -p "${K3S_SSH_PORT}" -i ssh_key ubuntu@0.0.0.0 true; do
    local current_time=$(date +%s)
    local elapsed_time=$((current_time - start_time))

    if [ "$elapsed_time" -ge "$timeout" ]; then
      echo "Serial log output:"
      tail -n 100 serial.log
      echo "Killing QEMU..."
      kill $(cat "$pid_file")
      rm -f "$pid_file"
      exit 1
    fi

    echo "Retrying SSH..."
  done

  echo "SSH connection established."
}

boot_qemu() {
  echo "Booting QEMU..."
 
  if [ $# -lt 1 ]; then
    QCOW=temp.img
    rm -f temp.img
    qemu-img create -F qcow2 -f qcow2 -b k3s-base.img temp.img
  else
    QCOW=$1
  fi

  qemu-system-x86_64 \
    -machine accel=kvm,type=q35 \
    -cpu host \
    -m 8G \
    -device virtio-net-pci,netdev=net0 \
    -netdev user,id=net0,hostfwd=tcp::${K3S_SSH_PORT}-:22,hostfwd=tcp::${K3S_API_PORT}-:6443,hostfwd=tcp::${K3S_WEB_PORT}-:80 \
    -drive if=virtio,format=qcow2,file=$QCOW,cache=none \
    -drive if=virtio,format=qcow2,file=seed.img \
    -nographic
}

stop_qemu() {
  if [ -f "$pid_file" ]; then
      pid=$(cat "$pid_file")

      if ps -p $pid > /dev/null 2>&1; then
          # Check if the process is QEMU
          if ps -p $pid -o comm= | grep -q "qemu-system"; then
              echo "QEMU pid $pid found. Killing it..."

              kill $pid
          else
              echo "PID file found, but the process is not QEMU. Skipping termination."
          fi
      else
          echo "PID file found, but the process is not running. Removing the PID file..."
      fi

      rm -f "$pid_file"
  else
      echo "No QEMU PID file found."
  fi
}

shutdown_qemu() {
  local timeout=120
  local interval=2

  echo "Shutdown QEMU..."
  ssh -o BatchMode=yes -p $K3S_SSH_PORT -i ssh_key ubuntu@0.0.0.0 sudo systemctl poweroff
  
  if [ -f "$pid_file" ]; then
    local pid=$(cat "$pid_file")
    
    echo "Waiting for QEMU to exit..."
    while [ $timeout -gt 0 ]; do
        if kill -0 "$pid" 2>/dev/null; then
          sleep $interval
        else
          echo "QEMU Process $pid has exited."
          return
        fi
        timeout=$((timeout - interval))
    done
  else
    echo "No QEMU PID file found."
  fi

  echo "Timeout reached. QEMU process $pid is still running."
  stop_qemu
}

install_k3s() {
  echo "Installing k3s..."

  k3sup install --ip ${K3S_IP-0.0.0.0} --ssh-key ssh_key --ssh-port ${K3S_SSH_PORT-22} --user ${K3S_USER-ubuntu}
  sed -i -e s/:6443/:${K3S_API_PORT}/ ./kubeconfig
  k3sup ready --kubeconfig=./kubeconfig
}

cloud_init() {
  echo "Could init"

  if [ ! -f ssh_key ]; then
    ssh-keygen -t ed25519 -f ssh_key -C "pou3" -N ""
  else
    echo "ssh_key exists, skipping key creation"
  fi
  local pub_key=$(cat ssh_key.pub)
  cat > user-data.yaml <<EOF
#cloud-config
growpart:
  mode: auto
  devices: ["/"]

sudo: ALL=(ALL) NOPASSWD:ALL
ssh_authorized_keys:
  - $pub_key
EOF

  cloud-localds -d qcow2 seed.img user-data.yaml metadata.yaml
  if [ ! -f ubuntu-base.img ]; then
    curl https://cloud-images.ubuntu.com/minimal/releases/jammy/release/ubuntu-22.04-minimal-cloudimg-amd64.img -o ubuntu-base.img
  else
    echo "ubuntu-base.img exists, skipping download"
  fi
  qemu-img create -F qcow2 -f qcow2 -b ubuntu-base.img k3s-base.img 200G

  ssh-keygen -R [0.0.0.0]:2222
  boot_qemu_daemon k3s-base.img
  install_k3s
  shutdown_qemu
}

setup_image() {
  echo "Setup Image"

  export KUBECONFIG

  kubectl apply -f ./priv-shell.yaml --prune -l premai.io/util=shell
  kubectl -n shell wait --for=condition=Ready pods/shell
  kubectl cp $IMG shell/shell:/tmp
  kubectl -n shell exec -it pods/shell -- k3s ctr images import /tmp/$IMG
}

tests() {
  echo "e2e tests"

  export KUBECONFIG=$(realpath $KUBECONFIG)

  kubectl -n prem-operator-system wait --for=condition=Available deployments/prem-operator-controller-manager
  cd ../../tests && go run github.com/onsi/ginkgo/v2/ginkgo -r -v ./e2e
}

clean() {
  rm -r ./*.img
  rm -r ./ssh_key*
  rm -r ./serial.log
  rm -r ./kubeconfig
  rm -r ./user-data.yaml
}

show_help() {
  echo "Usage: $0 <command> [options]"
  echo ""
  echo "Commands:"
  echo "  boot-qemu-daemon  Boot QEMU as a daemon"
  echo "  boot-qemu         Boot QEMU"
  echo "  stop-qemu         Stop QEMU"
  echo "  shutdown-qemu     Shutdown QEMU"
  echo "  install-k3s       Install k3s to a running system"
  echo "  cloud-init        Create a new VM and install K3s on it"
  echo "  setup-image       Add the operator container image to a running system"
  echo "  tests             Run e2e tests"
  echo "  clean             Clean up all created files"
  echo "  help              Show this help message"
}

# Check if at least one argument is provided
if [ $# -lt 1 ]; then
    show_help
    exit 1
fi

original_pwd=$(pwd)
cd $(dirname $0)

cleanup() {
  cd "$original_pwd"
}
trap cleanup EXIT

set_vars

case "$1" in
  boot-qemu-daemon)
    boot_qemu_daemon "${@:2}"
    ;;
  boot-qemu)
    boot_qemu "${@:2}"
    ;;
  stop-qemu)
    stop_qemu "${@:2}"
    ;;
  shutdown-qemu)
    shutdown_qemu "${@:2}"
    ;;
  install-k3s)
    install_k3s "${@:2}"
    ;;
  cloud-init)
    cloud_init "${@:2}"
    ;;
  setup-image)
    setup_image "${@:2}"
    ;;
  tests)
    tests "${@:2}"
    ;;
  clean)
    clean
    ;;
  help)
    show_help
    ;;
  *)
    echo "Error: Unknown command '$1'"
    show_help
    exit 1
    ;;
esac
