# Taken from https://github.com/microsoft/DeepSpeed/blob/master/docker/Dockerfile
FROM nvidia/cuda:12.2.2-devel-ubuntu20.04

ENV DEBIAN_FRONTEND=noninteractive

# Taken from https://www.notion.so/premai/Mamba-Deployment-Image-62731eb173f8452ebee8c9e19a0cef1d?pvs=4
RUN apt update && apt install python3 python3-pip -y

# Taken from DeepSpeed Dockerfile
RUN pip install torch==2.1.2
# DeepSpeed plus precompiled kernels (I hope)
RUN pip install deepspeed-mii==0.2.0 deepspeed-kernels

# Patch DeepSpeed to work with MIG
COPY ./0001-cuda-Use-MIG-device-handle-to-get-stats.patch /tmp/0001.patch
RUN cd /usr/local/lib/python3.8/dist-packages/deepspeed && patch -p1 < /tmp/0001.patch

# Patch DeepSpeed-MII
COPY ./0001-Add-kubernetes-health-check-route-to-REST-server.patch /tmp/0001.patch
RUN cd /usr/local/lib/python3.8/dist-packages && patch -p1 < /tmp/0001.patch

COPY ./serve-rest-api.py ./serve-rest-api.py

ENTRYPOINT [ "python3", "serve-rest-api.py" ]
CMD ["--uri", "microsoft/phi-1_5"]
