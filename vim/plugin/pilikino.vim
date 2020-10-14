" This plugin file sets up user commands and plugin mappings.

command! -nargs=* Pilikino call pilikino#search({ 'query': <q-args> })

imap <silent> <Plug>PilikinoInsertLink @<Esc>:call pilikino#insert_result()<CR>
