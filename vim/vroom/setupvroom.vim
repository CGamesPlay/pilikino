let s:dir = expand('<sfile>:h:h')
execute 'set rtp ^='.s:dir
runtime! plugin/**/*.vim
