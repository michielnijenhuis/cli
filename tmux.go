package cli

import (
	"fmt"
	"strings"
)

type Tmux struct {
	io *IO
}

func (t *Tmux) SessionExists(session string) bool {
	cmd := fmt.Sprintf("tmux ls 2>/dev/null | grep -E \"^%s:\"", session)

	res, err := t.io.ShPipe(cmd)
	if err != nil {
		return false
	}

	if res == "" {
		return false
	}

	return true
}

func (t *Tmux) NewSessionCommand(session string, detached bool) string {
	cmd := fmt.Sprintf("tmux new-session -d -s %s", session)
	if detached {
		cmd = fmt.Sprintf("%s -d", cmd)
	}

	return cmd
}

func (t *Tmux) AttachSessionCommand(session string) string {
	return fmt.Sprintf("tmux attach -t %s", session)
}

func (t *Tmux) NewWindowCommand(session string, window string) string {
	return fmt.Sprintf("tmux new-window -t %s -n %s -a", session, window)
}

func (t *Tmux) SelectWindowCommand(session string, window string) string {
	return fmt.Sprintf("tmux select-window -t %s:%s", session, window)
}

func (t *Tmux) RenameWindowCommand(session string, window string, name string) string {
	return fmt.Sprintf("tmux rename-window -t %s:%s %s", session, window, name)
}

func (t *Tmux) SendKeysCommand(session string, window string, keys string) string {
	return fmt.Sprintf("tmux send-keys -t %s:%s \"%s\" Enter", session, window, strings.Join(StringToInputArgs(keys), " "))
}

func (t *Tmux) SplitWindowHorizontallyCommand(session string, window string) string {
	return fmt.Sprintf("tmux split-window -h -t %s:%s", session, window)
}

func (t *Tmux) SplitWindowVerticallyCommand(session string, window string) string {
	return fmt.Sprintf("tmux split-window -v -t %s:%s", session, window)
}

func (t *Tmux) SelectPaneCommand(session string, window string, pane int) string {
	return fmt.Sprintf("tmux select-pane -t %s:%s.%d", session, window, pane)
}

func (t *Tmux) SendKeysToPaneCommand(session string, window string, pane int, keys string) string {
	return fmt.Sprintf("tmux send-keys -t %s:%s.%d %s Enter", session, window, pane, keys)
}

func (t Tmux) KillWindowCommand(session string, window string) string {
	return fmt.Sprintf("tmux kill-window -t %s:%s", session, window)
}

func (t Tmux) KillSessionCommand(session string) string {
	return fmt.Sprintf("tmux kill-session -t %s", session)
}

func (t Tmux) Exec(cmd string) error {
	return t.io.Zsh(cmd)
}

func (t Tmux) ExecMultiple(cmds []string) error {
	return t.io.Zsh(strings.Join(cmds, "; "))
}
