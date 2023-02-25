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
		rcerror = errortree.Add(rcerror, "parse", err)
		return nil, rcerror
	}
	if u.Scheme != "logger" {
		rcerror = errortree.Add(rcerror, "parse", fmt.Errorf("invalid scheme %s", URI))
		return nil, rcerror
	}
	switch u.Opaque {
	case "logrus":
		level = u.Query().Get("level")
		if level == "" {
			rcerror = errortree.Add(rcerror, "parse", errors.New("missing level query argument"))
			return nil, rcerror
		}
		output := u.Query().Get("output")
		if output == "" {
			rcerror = errortree.Add(rcerror, "parse", errors.New("missing output query argument"))
			return nil, rcerror
		}
		l = logrus.NewLogger()
		l.SetLevel(level)
		if err = l.SetFormat(fmt.Sprintf("logrus:hooked?output=%s", url.QueryEscape(output))); err != nil {
			rcerror = errortree.Add(rcerror, "parse", err)
			return nil, rcerror
		}
	default:
		rcerror = errortree.Add(rcerror, "parse", fmt.Errorf("unsupported logger implementation %q", u.Opaque))
		return nil, rcerror
	}

	return l, nil
}
