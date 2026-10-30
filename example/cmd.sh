#!/usr/bin/bash

function foo() {
    file=$(cli arg file) ### The argument is required and validated, so we know it's not empty

    if [ ! -f "$file" ]; then
        cli print --error "Invalid file path: $file"
        return 1
    fi

    contents=$(cat "$file")
    color=$(cli flag color)

    if cli flag reverse -x; then
        contents=$(echo "$contents" | rev)
    fi

    cli print "<fg=$color>$contents</>"
}

"$@"