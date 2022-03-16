FROM gitpod/workspace-full
RUN sudo apt-get update
RUN sudo apt-get install -y
RUN sudo apt-get install -y build-essential
RUN  echo "Compiling all modules, including geth ..."
RUN make all
RUN go env -w GO11MODULE=auto
