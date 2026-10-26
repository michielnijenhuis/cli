#!/usr/bin/env bash

source "$HOME/dotfiles/.zshrc.base"

function hello() {
    greeting="$(cli flag greeting)"
    exclamate="$(cli flag exclamate)"
    name="$(cli arg name)"

    message="${greeting}, ${name}"
    if [[ "${exclamate}" == "true" ]]; then
        message="${message}!"
    fi

    echo "${message}"
}

function lines() {
    number_of_lines="$(cli flag number-of-lines)"

    cli print --warning "number of lines: ${number_of_lines}"
}

function a_b() {
    echo "a:b"
}

function a_b_c() {
    echo "a:b:c"
}

function a_b_d() {
    echo "a:b:d"
}

function input() {
    answer="$(cli input)"

    cli print --success "You entered: $answer"
}

function choose() {
    answer="$(cli choose foo bar baz)"
    if [[ "${answer}" == "" ]]; then
        cli print --error "You did not choose anything"
        return 1
    fi

    cli print --success "You chose: $answer"
}

function foo() {
    answer="$(gum input)"
    cli print --info "${answer}"

    choice="$(gum choose foo bar baz)"
    cli print --info "You chose: ${choice}"
}

cli_run "$@"