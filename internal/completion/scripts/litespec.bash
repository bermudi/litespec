#!/bin/bash
_litespec() {
	local words cword
	if type _get_comp_words_by_ref &>/dev/null; then
		_get_comp_words_by_ref -w words -i cword
	else
		words=("${COMP_WORDS[@]}")
		cword=$COMP_CWORD
	fi

	local -a args=("${words[@]:0:$((cword+1))}")

	local IFS=$'\n'
	local candidates
	candidates=$(litespec __complete "${args[@]}" 2>/dev/null)

	COMPREPLY=()
	while IFS=$'\t' read -r cand desc; do
		if [[ -n "$cand" ]]; then
			COMPREPLY+=("$cand")
		fi
	done <<<"$candidates"
}

complete -F _litespec litespec
