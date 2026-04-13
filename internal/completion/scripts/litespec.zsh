#compdef litespec

_litespec() {
  _arguments -C \
    '1:command:->command' \
    '*::arg:->args'

  case $state in
    command)
      local -a values
      local -a raw
      local IFS=$'\n'
      raw=(${(f)"$(litespec __complete litespec "" 2>/dev/null)"})

      local cand desc
      for line in "${raw[@]}"; do
        IFS=$'\t' read -r cand desc <<< "$line"
        if [[ -n "$cand" ]]; then
          values+=("${cand}${desc:+\:${desc//:/\\:}}")
        fi
      done

      _describe 'command' values
      ;;
    args)
      local -a values
      local IFS=$'\n'
      local -a raw
      raw=(${(f)"$(litespec __complete "${words[@]}" 2>/dev/null)"})

      local cand desc
      for line in "${raw[@]}"; do
        IFS=$'\t' read -r cand desc <<< "$line"
        if [[ -n "$cand" ]]; then
          values+=("${cand}${desc:+\:${desc//:/\\:}}")
        fi
      done

      _describe 'argument' values
      ;;
  esac
}

_litespec "$@"
