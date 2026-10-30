# CLI

CLI Toolkit inspired by PHP's symfony/console and Golang's spf13/cobra.

## Installation

```bash
$ go install github.com/michielnijenhuis/cli/cmd
```

## Documentation

Check out the `example` directory to learn more about how to use this package.

## Features

- Automatic --help and --version flag implementation.
- A print command that makes printing with colors easy:
    - ```bash
      $ cli print "<fg=white;bg=red>Oh no!</>"
      $ cli print --info "All is well."
      ```
- Automatic flag and argument parsing, with validation
- Completion script generation
- Configurable theme