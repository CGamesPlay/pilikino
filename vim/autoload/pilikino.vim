" A lot of this file comes directly from the FZF base plugin.

let s:save_cpo = &cpo
set cpo&vim

if !exists('g:pilikino_directory')
  let g:pilikino_directory = '.'
end
if !exists('g:pilikino_layout')
  let g:pilikino_layout = { 'up': '40%' }
end
if !exists('g:pilikino_actions')
  " Map from expected key name to action in vim
  let g:pilikino_actions = {
    \ 'enter': 'e',
    \ 'ctrl-x': 'topleft sp',
    \ 'ctrl-v': 'botright vert sp',
    \ 'ctrl-t': 'tab e',
    \ }
end
if !exists('g:pilikino_link_template')
  let g:pilikino_link_template = '{filename}'
end

let s:default_binary = 'pilikino'
let s:is_win = has('win32') || has('win64')

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
  let opts = a:opts
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
  try
    let [shell, shellslash, shellcmdflag, shellxquote] = s:use_sh()
    let opts = a:opts
    function! s:exit_cb(job, status, ...) closure
      if has_key(opts, 'on_exit')
        call opts.on_exit(a:status)
      end
    endfunction
    if has('nvim')
      let termopts = { 'on_exit': funcref('s:exit_cb') }
      call termopen([&shell, &shellcmdflag, a:opts.command], termopts)
      let buf = bufnr('')
      startinsert
    else
      let termopts = { 'curwin': 1, 'exit_cb': funcref('s:exit_cb') }
      let buf = term_start([&shell, &shellcmdflag, a:opts.command], termopts)
      if !has('patch-8.0.1261') && !has('nvim') && !s:is_win
        call term_wait(buf, 20)
      endif
    endif
    setlocal nospell bufhidden=wipe nobuflisted nonumber
    return buf
  finally
    let [&shell, &shellslash, &shellcmdflag, &shellxquote] = [shell, shellslash, shellcmdflag, shellxquote]
  endtry
endfunction

" Executes Pilikino.
"
" args.callback, required, function to call with the results on success
" args.cancel, optional, function to call when user aborts search
" args.args, optional, a list of strings to pass as arguments
" args.layout, optional, the layout to use for the split
function! pilikino#exec(args) abort
  let args = a:args
  let split = {}
  if has_key(args, 'layout')
    let split = copy(args.layout)
  end
  let tempfile = tempname()
  if s:is_win
    " Need to write a batch file to set up the file redirection, apparently
    throw 'not implemented on windows'
  end
  try
    let command = s:binary_path().' search'
  catch
    " Clean up the stack trace in case program not installed
    throw v:exception
  endtry
  if has_key(a:args, 'args')
    let command .= ' '.join(map(a:args.args, { _, v -> s:shellesc(v) }), " ")
  end
  let command .= ' > '.tempfile
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
    elseif has_key(args, 'cancel')
      call args.cancel()
    end
  endfunction
  call s:in_split(split)
endfunction

" opts.actions is a dictionary of key names mapping to the desired command to
" run, which will be executed with the selected filename (with the directory
" prepended to it). The action can also be a funcref which will be called with
" the unmodified filename.
" opts.cancel is a funcref which will be called if the search is aborted.
function! pilikino#search(...) abort
  let default_opts = {
    \ 'actions': g:pilikino_actions,
    \ 'directory': g:pilikino_directory,
    \ 'layout': g:pilikino_layout,
    \ }
  let opts = a:0 > 0 ? copy(a:1) : {}
  call extend(opts, default_opts, 'keep')

  let opts.args = [
    \ '--expect', join(keys(opts.actions), ","),
    \ '--directory', opts.directory,
    \ ]

  function! opts.callback(results) closure
    let Action = opts.actions[a:results[0]]
    if type(Action) == v:t_string
      let filename = simplify(opts.directory.'/'.a:results[1])
      exec Action.' ' filename
    else
      let filename = a:results[1]
      call Action(filename)
    end
  endfunction
  call pilikino#exec(opts)
endfunction

function! pilikino#format_link(filename, template) abort
  let title = fnamemodify(a:filename, ":r")
  let filename = a:filename
  try
    let Replacer = function(a:template)
  catch
  endtry
  if exists('Replacer')
    return Replacer(filename)
  end
  let result = a:template
  let result = substitute(result, "{title}", title, "g")
  let result = substitute(result, "{filename}", filename, "g")
  return result
endfunction

" Perform a pilikino search, and insert the found filename into the current
" buffer. This method is expected to be called with the cursor positioned over
" a placeholder character, which will be replaced with the link.
function! pilikino#insert_result(...)
  let default_opts = { 'template': g:pilikino_link_template }
  let opts = a:0 > 0 ? a:1 : {}
  call extend(opts, default_opts, 'keep')
  function! s:do_insert(result) closure
    undojoin
    let link = pilikino#format_link(a:result, opts.template)
    let line = getline('.')
    let column = col('.')
    " The cursor is presently sitting on an inserted placeholder, which should
    " be replaced with the link.
    call setline('.', strpart(line, 0, column - 1).link.strpart(line, column))
    " Now move the cursor to end of the link and press "a" to resume insert
    " mode.
    call cursor(line('.'), column + len(link) - 1)
    call feedkeys("a", "n")
  endfunction
  function! s:do_cancel() closure
    " Delete the placeholder character and reenter insert mode, without even
    " making an undo checkpoint.
    undojoin | call feedkeys("s", "n")
  endfunction
  call pilikino#search({
    \ 'actions': { 'enter': funcref('s:do_insert') },
    \ 'cancel': funcref('s:do_cancel'),
    \ })
endfunction

let &cpo = s:save_cpo
