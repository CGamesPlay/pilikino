" This plugin file sets up user commands and plugin mappings.

command! Pilikino call pilikino#search()

imap <silent> <Plug>PilikinoInsertLink @<Esc>:call pilikino#insert_result()<CR>
