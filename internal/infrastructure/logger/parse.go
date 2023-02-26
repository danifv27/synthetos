package logger

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/speijnik/go-errortree"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/infrastructure/logger/logrus"
)

// URI "logger:logrus?level=<logrus_level>ḉ&output=[plain|json]"
func Parse(URI string) (logger.Logger, error) {
	var level string
	var l logger.Logger
	var rcerror error

	u, err := url.Parse(URI)
	if err != nil {
		return nil, errortree.Add(rcerror, "Parse", err)
	}
	if u.Scheme != "logger" {
		return nil, errortree.Add(rcerror, "Parse", fmt.Errorf("invalid scheme %s", URI))
	}
	switch u.Opaque {
	case "logrus":
		level = u.Query().Get("level")
		if level == "" {
			return nil, errortree.Add(rcerror, "Parse", errors.New("missing level query argument"))
		}
		output := u.Query().Get("output")
		if output == "" {
			return nil, errortree.Add(rcerror, "Parse", errors.New("missing output query argument"))
		}
		l = logrus.NewLogger()
		l.SetLevel(level)
		if err = l.SetFormat(fmt.Sprintf("logrus:hooked?output=%s", url.QueryEscape(output))); err != nil {
			return nil, errortree.Add(rcerror, "Parse", err)
		}
	default:
		return nil, errortree.Add(rcerror, "Parse", fmt.Errorf("unsupported logger implementation %q", u.Opaque))
	}

	return l, nil
}
