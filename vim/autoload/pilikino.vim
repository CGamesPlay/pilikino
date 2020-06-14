" A lot of this file comes directly from the FZF base plugin.

let s:save_cpo = &cpo
set cpo&vim

let s:default_binary = 'pilikino'
let s:is_win = has('win32') || has('win64')
let s:default_layout = { 'up': '40%' }
" Map from expected key name to action in vim
let s:default_actions = {
  \ 'enter': 'e',
  \ 'ctrl-x': 'topleft sp',
  \ 'ctrl-v': 'botright vert sp',
  \ 'ctrl-t': 'tab e',
  \ }

" Windows CMD.exe shell escaping from fzf
function! s:shellesc_cmd(arg)
  let escaped = substitute(a:arg, '[&|<>()@^]', '^&', 'g')
  let escaped = substitute(escaped, '%', '%%', 'g')
  let escaped = substitute(escaped, '"', '\\^&', 'g')
  let escaped = substitute(escaped, '\(\\\+\)\(\\^\)', '\1\1\2', 'g')
  return '^"'.substitute(escaped, '\(\\\+\)$', '\1\1', '').'^"'
endfunction

" Windows-safe shell escaping from fzf
function! s:shellesc(arg, ...)
  let shell = get(a:000, 0, &shell)
  if shell =~# 'cmd.exe$'
    return s:shellesc_cmd(a:arg)
  elseif s:is_win
    let shellslash = &shellslash
    try
      set noshellslash
      return shellescape(a:arg)
    finally
      let &shellslash = shellslash
    endtry
  endif
  return shellescape(a:arg)
endfunction

function! s:binary_path()
  if !exists('s:binary_path_')
    if executable(s:default_binary)
      let s:binary_path_ = s:default_binary
    else
      throw 'pilikino executable not found'
    endif
  endif
  return s:shellesc(s:binary_path_)
endfunction

function! s:use_sh()
  let [shell, shellslash, shellcmdflag, shellxquote] = [&shell, &shellslash, &shellcmdflag, &shellxquote]
  if s:is_win
    set shell=cmd.exe
    set noshellslash
    let &shellcmdflag = has('nvim') ? '/s /c' : '/c'
    let &shellxquote = has('nvim') ? '"' : '('
  else
    set shell=sh
  endif
  return [shell, shellslash, shellcmdflag, shellxquote]
endfunction

function! s:error(msg)
  echohl ErrorMsg
  echom a:msg
  echohl None
endfunction

" Return a collection of information about the active tabs/windows, used to
" decide if the layout has changed substantially from one time to another.
function! s:getpos()
  return {'tab': tabpagenr(), 'win': winnr(), 'cnt': winnr('$'), 'tcnt': tabpagenr('$')}
endfunction

" Does the dictionary have any of the keys in it?
function! s:present(dict, ...)
  for key in a:000
    if !empty(get(a:dict, key, ''))
      return 1
    endif
  endfor
  return 0
endfunction

" Check to verify there is a reasonable amount of space to create a split
function! s:splittable(opts)
  return s:present(a:opts, 'up', 'down') && &lines > 15 ||
        \ s:present(a:opts, 'left', 'right') && &columns > 40
endfunction

" Parse a value either as an absolute or a percentage of the max value.
function! s:parse_size(max, val)
  if a:val =~ '%$'
    return a:max * str2nr(a:val[:-2]) / 100
  else
    return min([a:max, str2nr(a:val)])
  endif
endfunction

" Create a new split with the provided options
function! s:split(opts)
  let directions = {
  \ 'up':    ['topleft', 'resize', &lines],
  \ 'down':  ['botright', 'resize', &lines],
  \ 'left':  ['vertical topleft', 'vertical resize', &columns],
  \ 'right': ['vertical botright', 'vertical resize', &columns] }
  let ppos = s:getpos()
  let opts = copy(s:default_layout)
  call extend(opts, a:opts)
  try
    if s:present(opts, 'window')
      execute 'keepalt' opts.window
    elseif !s:splittable(opts)
      execute (tabpagenr()-1).'tabnew'
    else
      for [dir, triple] in items(directions)
        let val = get(opts, dir, '')
        if !empty(val)
          let [cmd, resz, max] = triple
          let sz = s:parse_size(max, val)
          execute cmd sz.'new'
          execute resz sz
          return [ppos, {}]
        endif
      endfor
    endif
    return [ppos, { '&l:wfw': &l:wfw, '&l:wfh': &l:wfh }]
  finally
    setlocal winfixwidth winfixheight
  endtry
endfunction

" Creates a new split, switches to it, and runs opts.func. The function will
" be passed a callback which closes the split and restores the window layout.
" The opts.callback will be called with the arguments passed to the callback.
function! s:in_split(opts) abort
  let winrest = winrestcmd()
  let pbuf = bufnr('') " previous buffer
  let [ppos, winopts] = s:split(a:opts)

  let data = {
    \ 'buf': bufnr(''), 'pbuf': pbuf, 'ppos': ppos, 'winopts': winopts,
    \ 'opts': a:opts,
    \ 'winrest': winrest, 'lines': &lines, 'columns': &columns,
    \ }

  function! data.switch_back(inplace)
    if a:inplace && bufnr('') == self.buf
      if bufexists(self.pbuf)
        execute 'keepalt b' self.pbuf
      endif
      " No other listed buffer
      if bufnr('') == self.buf
        enew
      endif
    endif
  endfunction
  function! data.cleanup(...)
    if s:getpos() == self.ppos " {'window': 'enew'}
      for [opt, val] in items(self.winopts)
        execute 'let' opt '=' val
      endfor
      call self.switch_back(1)
    else
      if bufnr('') == self.buf
        " We use close instead of bd! since Vim does not close the split when
        " there's no other listed buffer (nvim +'set nobuflisted')
        close
      endif
      execute 'tabnext' self.ppos.tab
      execute self.ppos.win.'wincmd w'
    endif

    if bufexists(self.buf)
      execute 'bd!' self.buf
    endif

    " If screen was resized, don't attempt ot restore old window configuration
    if &lines == self.lines && &columns == self.columns && s:getpos() == self.ppos
      execute self.winrest
    endif

    if has_key(self.opts, 'callback')
      call call(self.opts.callback, a:000)
    endif
    " Unclear why this is here, since it was theoretically handled above?
    call self.switch_back(s:getpos() == self.ppos)
  endfunction

  function! s:bound_cb(...) closure
    call call(data.cleanup, a:000)
  endfunction

  call a:opts.func(funcref('s:bound_cb'))
endfunction

" Runs opts.command in a terminal in the current buffer, and call
" opts.on_exit with the exit status when it finishes. Returns the buffer
" number of the terminal window.
function! s:run_terminal(opts) abort
  let has_vim8_term = has('terminal') && has('patch-8.0.995')
  let has_nvim_term = has('nvim-0.2.1') || has('nvim') && !s:is_win
  if !has_nvim_term && !has_vim8_term
    throw 'unsupported vim version'
  end
  if has('nvim')
    throw 'nvim not implemented'
  end
  try
    let [shell, shellslash, shellcmdflag, shellxquote] = s:use_sh()
    let opts = a:opts
    function! s:exit_cb(job, status) closure
      if has_key(opts, 'on_exit')
        call opts.on_exit(a:status)
      end
    endfunction
    let termopts = { 'curwin': 1, 'exit_cb': funcref('s:exit_cb') }
    let buf = term_start([&shell, &shellcmdflag, a:opts.command], termopts)
    if !has('patch-8.0.1261') && !has('nvim') && !s:is_win
      call term_wait(buf, 20)
    endif
    setlocal nospell bufhidden=wipe nobuflisted nonumber
    return buf
  finally
    let [&shell, &shellslash, &shellcmdflag, &shellxquote] = [shell, shellslash, shellcmdflag, shellxquote]
  endtry
endfunction

" Executes Pilikino. If the execution is successful and the user selects a
" file, this function calls args.callback with the results.
"
" You can pass arguments to the command directly by setting these keys:
" expect, directory.
function! pilikino#exec(args) abort
  let args = a:args
  let split = {}
  let tempfile = tempname()
  if s:is_win
    " Need to write a batch file to set up the file redirection, apparently
    throw 'not implemented on windows'
  end
  let command = s:binary_path().' search'
  let stringArgs = ['expect', 'directory']
  for argName in stringArgs
    if has_key(args, argName)
      let command .= ' --'.argName.' '.s:shellesc(args[argName])
    endif
  endfor
  try
    let command .= ' > '.tempfile
  catch
    throw v:exception
  endtry
  function! split.func(next) closure
    let termopts = { 'command': command, 'next': a:next, 'output': tempfile }
    function! termopts.on_exit(status) abort
      let result = []
      try
        if filereadable(self.output)
          let result = readfile(self.output)
        endif
      finally
        silent! call delete(self.output)
      endtry
      if a:status > 1 && a:status != 130
        call s:error('Failed to execute pilikino')
      else
        call self.next(a:status, result)
      end
    endfunction
    call s:run_terminal(termopts)

    setlocal statusline=Pilikino
  endfunction
  function! split.callback(status, result) closure
    if a:status == 0
      call args.callback(a:result)
    end
  endfunction
  call s:in_split(split)
endfunction

" opts.actions is a dictionary of key names mapping to the desired command
" to run, which will receive the selected filename.
function! pilikino#search(...)
  let opts = a:0 > 0 ? a:1 : {}
  let actions = has_key(opts, 'actions') ? opts.actions : s:default_actions
  let directory = has_key(opts, 'directory') ? opts.directory : '.'
  let directory = '/Users/rpatterson/Seafile/Notes/'
  let args = { 'expect': join(keys(actions), ','), 'directory': directory }
  function! args.callback(results) closure
    let action = actions[a:results[0]]
    exec action.' ' simplify(directory.'/'.a:results[1])
  endfunction
  call pilikino#exec(args)
endfunction

let &cpo = s:save_cpo
