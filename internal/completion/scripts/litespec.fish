# Fish completions for litespec

# Disable file completions
complete -c litespec -f

# Top-level commands
complete -c litespec -n '__fish_use_subcommand' -a '(litespec __complete litespec 2>/dev/null | string replace \t "	")' -d 'command'

# init
complete -c litespec -n '__fish_seen_subcommand_from init' -l tools -d 'Tool IDs (comma-separated)'
complete -c litespec -n '__fish_seen_subcommand_from init' -a '(litespec __complete litespec init --tools 2>/dev/null | string replace \t "	")'

# new - no completions for free-text name

# list
complete -c litespec -n '__fish_seen_subcommand_from list' -l specs -d 'List specs instead of changes'
complete -c litespec -n '__fish_seen_subcommand_from list' -l changes -d 'List changes (default)'
complete -c litespec -n '__fish_seen_subcommand_from list' -l sort -d 'Sort by recent or name'
complete -c litespec -n '__fish_seen_subcommand_from list' -l json -d 'Output as JSON'

# status
complete -c litespec -n '__fish_seen_subcommand_from status' -l json -d 'Output as JSON'
complete -c litespec -n '__fish_seen_subcommand_from status; and not __fish_seen_argument' -a '(litespec __complete litespec status "" 2>/dev/null | string replace \t "	")' -d 'change'

# validate
complete -c litespec -n '__fish_seen_subcommand_from validate' -l all -d 'Validate all changes and specs'
complete -c litespec -n '__fish_seen_subcommand_from validate' -l changes -d 'Validate all changes only'
complete -c litespec -n '__fish_seen_subcommand_from validate' -l specs -d 'Validate all specs only'
complete -c litespec -n '__fish_seen_subcommand_from validate' -l strict -d 'Treat warnings as errors'
complete -c litespec -n '__fish_seen_subcommand_from validate' -l json -d 'Output as JSON'
complete -c litespec -n '__fish_seen_subcommand_from validate' -l type -d 'Disambiguate name: change|spec'

# instructions
complete -c litespec -n '__fish_seen_subcommand_from instructions' -l json -d 'Output as JSON'
complete -c litespec -n '__fish_seen_subcommand_from instructions; and not __fish_seen_argument' -a '(litespec __complete litespec instructions "" 2>/dev/null | string replace \t "	")' -d 'artifact'

# archive
complete -c litespec -n '__fish_seen_subcommand_from archive' -l allow-incomplete -d 'Archive even with incomplete tasks'
complete -c litespec -n '__fish_seen_subcommand_from archive; and not __fish_seen_argument' -a '(litespec __complete litespec archive "" 2>/dev/null | string replace \t "	")' -d 'change'

# preview
complete -c litespec -n '__fish_seen_subcommand_from preview' -l json -d 'Output as JSON'
complete -c litespec -n '__fish_seen_subcommand_from preview; and not __fish_seen_argument' -a '(litespec __complete litespec preview "" 2>/dev/null | string replace \t "	")' -d 'change'

# update
complete -c litespec -n '__fish_seen_subcommand_from update' -l tools -d 'Tool IDs (comma-separated)'
complete -c litespec -n '__fish_seen_subcommand_from update' -a '(litespec __complete litespec update --tools 2>/dev/null | string replace \t "	")'

# completion
complete -c litespec -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish' -d 'shell'
