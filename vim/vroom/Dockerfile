FROM alpine:3.12
RUN apk add --no-cache neovim git ncurses python3 python3-dev build-base
RUN \
  python3 -m ensurepip && \
  pip3 install neovim
RUN \
  git clone --depth=1 https://github.com/google/vroom.git && \
  cd vroom && \
  python3 setup.py build install && \
  cd .. && \
  rm -rf vroom
COPY pilikino /usr/bin/
WORKDIR /vroom
VOLUME /vroom
CMD ./vim/vroom/run-docker.sh
