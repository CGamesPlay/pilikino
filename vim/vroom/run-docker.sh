#!/bin/sh
set -e

cd pkg/pilikino/testdata
# NOTE: gvim doesn't support testing imaps, so the test suite fails there
vroom --neovim --crawl ../../../vim/vroom
# When manually debugging:
# nvim --cmd "set rtp^=/vroom/vim"
