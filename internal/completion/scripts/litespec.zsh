#compdef litespec

_litespec() {
  _arguments -C \
    '1:command:->command' \
    '*::arg:->args'

  case $state in
    command)
      local -a cmd_cands cmd_descs
      local -a raw
      local IFS=$'\n'
      raw=(${(f)"$(litespec __complete litespec "" 2>/dev/null)"})

      local cand desc
      for line in "${raw[@]}"; do
        IFS=$'\t' read -r cand desc <<< "$line"
        if [[ -n "$cand" ]]; then
          cmd_cands+=("$cand")
          cmd_descs+=("${desc:-}")
        fi
      done

      _describe 'command' cmd_cands cmd_descs
      ;;
    args)
      local -a candidates
      local IFS=$'\n'
      candidates=(${(f)"$(litespec __complete "${words[@]}" 2>/dev/null)"})

      local -a comp_cands
      local -a comp_descs
      local cand desc
      for line in "${candidates[@]}"; do
        IFS=$'\t' read -r cand desc <<< "$line"
        if [[ -n "$cand" ]]; then
          comp_cands+=("$cand")
          comp_descs+=("${desc:-}")
        fi
      done

      _describe 'candidate' comp_cands comp_descs
      ;;
  esac
}

_litespec "$@"
