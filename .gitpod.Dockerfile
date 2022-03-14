FROM gitpod/workspace-full
RUN sudo apt-get update
RUN sudo apt-get install -y
RUN go env -w GO11MODULE=auto