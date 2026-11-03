#!/usr/bin/bash

function foo() {
    file=$(cli arg file) ### The argument is required and validated, so we know it's not empty.

    if [ ! -f "$file" ]; then
        cli print --error "Invalid file path: $file" ### Use the error flag to pretty print an error.
        return 1
    fi

    contents=$(cat "$file")
    color=$(cli flag color)

    if cli flag reverse -q; then ### Exits with 0 if true, 1 if false. Silence the output ("true" or "false") since we don't need it.
        contents=$(echo "$contents" | rev)
    fi

    ### Print in the requested color. We could also set `<bg=$color>`, or provide options (e.g. `<options=bold,underline>`),
    ### or combine them (e.g. `<fg=white;bg=blue;options=bold,underline>text</>`). HEX color values are also accepted,
    ### e.g. `<fg=#0000ff>Blue!</>`.
    cli print "<fg=$color>$contents</>"
}

"$@"