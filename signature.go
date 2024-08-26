package cli

import (
	"fmt"
	"regexp"
	"strings"
)

type ParsedSignature struct {
	Expression  string
	Name        string
	Aliases     []string
	Description string
	Arguments   []*InputArgument
	Options     []*InputOption
}

func ParseSignature(expr string) (*ParsedSignature, error) {
	signature := &ParsedSignature{
		Expression: expr,
	}

	err := signature.parseNameAndDescription()
	if err != nil {
		return signature, err
	}

	matches := regexp.MustCompile(`\{\s*(.*?)\s*\}`).FindAllStringSubmatch(expr, -1)
	if len(matches) > 0 {
		tokens := make([]string, 0, len(matches))
		for _, match := range matches {
			tokens = append(tokens, match[1])
		}

		err = signature.parseParameters(tokens)
		if err != nil {
			return nil, err
		}
	}

	return signature, nil
}

func (s *ParsedSignature) parseNameAndDescription() error {
	matches := regexp.MustCompile(`(?m)^([\w|\|]+)(?:\s*:\s*([^{]+))?`).FindAllStringSubmatch(s.Expression, -1)

	if len(matches) == 0 {
		return fmt.Errorf("unable to determine command name from signature \"%s\"", s.Expression)
	}

	name := matches[0][1]
	name = strings.TrimSpace(name)
	parts := strings.Split(name, "|")
	s.Name = parts[0]

	if len(parts) > 1 {
		s.Aliases = parts[1:]
	}

	if len(matches[0]) > 1 {
		s.Description = strings.TrimSpace(matches[0][2])
	}

	return nil
}

func (s *ParsedSignature) parseParameters(tokens []string) error {
	args := make([]*InputArgument, 0)
	opts := make([]*InputOption, 0)

	for _, token := range tokens {
		matches := regexp.MustCompile(`^-{2,}(.*)`).FindAllString(token, -1)
		if len(matches) > 0 {
			opt := s.parseOption(matches[0])
			opts = append(opts, opt)
		} else {
			arg := s.parseArgument(token)
			args = append(args, arg)
		}
	}

	s.Arguments = args
	s.Options = opts

	return nil
}

func (s *ParsedSignature) parseOption(token string) *InputOption {
	parsedToken, description := s.extractDescription(token)
	shortcutAndName := regexp.MustCompile(`\s*\|\s*`).Split(parsedToken, 2)
	var shortcut string

	if len(shortcutAndName) > 1 {
		shortcut = shortcutAndName[0]
		parsedToken = shortcutAndName[1]
	}

	shortcut = strings.TrimPrefix(shortcut, "--")
	parsedToken = strings.TrimPrefix(parsedToken, "--")

	if strings.HasSuffix(parsedToken, "=") {
		return &InputOption{
			Name:        parsedToken[:len(parsedToken)-1],
			Shortcut:    shortcut,
			Mode:        InputOptionBool,
			Description: description,
		}
	}

	if strings.HasSuffix(parsedToken, "=*") {
		return &InputOption{
			Name:        parsedToken[:len(parsedToken)-2],
			Shortcut:    shortcut,
			Mode:        InputOptionBool | InputOptionIsArray,
			Description: description,
		}
	}

	matches := regexp.MustCompile(`(.+)=*(.+)`).FindAllString(parsedToken, -1)
	if len(matches) > 0 {
		name := strings.TrimPrefix(matches[0], "--")

		var defaultValue InputType
		if len(matches) > 1 {
			defaultValue = regexp.MustCompile(`,s?`).Split(matches[1], -1)
		}

		return &InputOption{
			Name:         name,
			Shortcut:     shortcut,
			Mode:         InputOptionBool | InputOptionIsArray,
			Description:  description,
			DefaultValue: defaultValue,
		}
	}

	matches = regexp.MustCompile(`(.+)=(.+)`).FindAllString(parsedToken, -1)
	if len(matches) > 0 {
		return &InputOption{
			Name:         matches[0],
			Shortcut:     shortcut,
			Mode:         InputOptionOptional,
			Description:  description,
			DefaultValue: matches[1],
		}
	}

	return &InputOption{
		Name:        parsedToken,
		Shortcut:    shortcut,
		Mode:        InputOptionBool,
		Description: description,
	}
}

// TODO
func (s *ParsedSignature) parseArgument(token string) *InputArgument {
	parsedToken, description := s.extractDescription(token)

	if strings.HasSuffix(parsedToken, "?*") {
		return &InputArgument{
			Name:        parsedToken[:len(parsedToken)-2],
			Mode:        InputArgumentIsArray,
			Description: description,
		}
	}

	if strings.HasSuffix(parsedToken, "*") {
		return &InputArgument{
			Name:        parsedToken[:len(parsedToken)-1],
			Mode:        InputArgumentIsArray | InputArgumentRequired,
			Description: description,
		}
	}

	if strings.HasSuffix(parsedToken, "?") {
		return &InputArgument{
			Name:        parsedToken[:len(parsedToken)-1],
			Mode:        InputArgumentOptional,
			Description: description,
		}
	}

	matches := regexp.MustCompile(`(.+)\=\*(.+)`).FindAllString(parsedToken, -1)
	if len(matches) > 0 {
		return &InputArgument{
			Name:         matches[1],
			Mode:         InputArgumentIsArray,
			Description:  description,
			DefaultValue: regexp.MustCompile(`,s?`).Split(matches[2], -1),
		}
	}

	matches = regexp.MustCompile(`(.+)\=(.+)`).FindAllString(parsedToken, -1)
	if len(matches) > 0 {
		return &InputArgument{
			Name:         matches[1],
			Mode:         InputArgumentOptional,
			Description:  description,
			DefaultValue: matches[2],
		}
	}

	return &InputArgument{
		Name:        parsedToken,
		Mode:        InputArgumentRequired,
		Description: description,
	}
}

func (s *ParsedSignature) extractDescription(token string) (string, string) {
	parts := regexp.MustCompile(`\s+:\s+`).Split(token, 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return token, ""
}
