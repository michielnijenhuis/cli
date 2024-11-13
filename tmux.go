package cli

import (
	"fmt"
	"strings"
)

type Tmux struct {
	io *IO
}

func (t Tmux) SessionExists(session string) bool {
	cmd := fmt.Sprintf("tmux has -t %s", session)
	_, err := t.io.ShPipe(cmd)

	return err == nil
}

func (t Tmux) NewSessionCommand(session string) string {
	return fmt.Sprintf("tmux new-session -s %s", session)
}

func (t Tmux) NewDetachedSessionCommand(session string) string {
	return t.NewSessionCommand(session) + " -d"
}

func (t Tmux) AttachSessionCommand(session string) string {
	return fmt.Sprintf("tmux attach -t %s", session)
}

func (t Tmux) NewWindowCommand(session string, window string) string {
	return fmt.Sprintf("tmux new-window -t %s -n %s -a", session, window)
}

func (t Tmux) SelectWindowCommand(session string, window string) string {
	return fmt.Sprintf("tmux select-window -t %s:%s", session, window)
}

func (t Tmux) RenameWindowCommand(session string, window string, name string) string {
	return fmt.Sprintf("tmux rename-window -t %s:%s %s", session, window, name)
}

func (t Tmux) SendKeysCommand(session string, window string, keys string) string {
	return fmt.Sprintf("tmux send-keys -t %s:%s \"%s\" Enter", session, window, strings.Join(StringToInputArgs(keys), " "))
}

func (t Tmux) SplitWindowHorizontallyCommand(session string, window string) string {
	return fmt.Sprintf("tmux split-window -h -t %s:%s", session, window)
}

func (t Tmux) SplitWindowVerticallyCommand(session string, window string) string {
	return fmt.Sprintf("tmux split-window -v -t %s:%s", session, window)
}

func (t Tmux) SelectPaneCommand(session string, window string, pane int) string {
	return fmt.Sprintf("tmux select-pane -t %s:%s.%d", session, window, pane)
}

func (t Tmux) SendKeysToPaneCommand(session string, window string, pane int, keys string) string {
	return fmt.Sprintf("tmux send-keys -t %s:%s.%d %s Enter", session, window, pane, keys)
}

func (t Tmux) KillWindowCommand(session string, window string) string {
	return fmt.Sprintf("tmux kill-window -t %s:%s", session, window)
}

func (t Tmux) KillSessionCommand(session string) string {
	return fmt.Sprintf("tmux kill-session -t %s", session)
}

func (t Tmux) CreateSplitWindowCommand(session string, window string, keys []string) string {
	cmds := []string{
		t.NewWindowCommand(session, window),
		t.SplitWindowHorizontallyCommand(session, window),
	}

	if len(keys) > 0 {
		cmds = append(cmds, t.SendKeysCommand(session, window+".1", keys[0]))
	}

	if len(keys) > 1 {
		cmds = append(cmds, t.SendKeysCommand(session, window+".2", keys[1]))
	}

	return strings.Join(cmds, " && ")
}

func (t Tmux) Exec(cmd string) error {
	return t.io.Zsh(cmd)
}

func (t Tmux) ExecMultiple(cmds []string) error {
	return t.io.Zsh(strings.Join(cmds, "; "))
}
