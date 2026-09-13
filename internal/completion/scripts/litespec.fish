# Fish completions for litespec

# Disable file completions
complete -c litespec -f

function __litespec_complete
    set -l commands (litespec __complete litespec "" 2>/dev/null | string match -r '^[^\t]+')
    set -l words (commandline -xpc) (commandline -ct)
    if not __fish_seen_subcommand_from $commands; and test (commandline -ct) = ""
        set words litespec ""
    end
    litespec __complete $words 2>/dev/null | string replace \t "	"
end

complete -c litespec -a '(__litespec_complete)'
