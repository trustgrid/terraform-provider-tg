package validators

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
)

func IsHostname(i any, k string) (warnings []string, errors []error) {
	v, ok := i.(string)

	switch {
	case !ok:
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
	case !govalidator.IsDNSName(v):
		errors = append(errors, fmt.Errorf("expected %s to contain a valid hostname, got: %s", k, v))
	case !strings.Contains(v, "."):
		errors = append(errors, fmt.Errorf("expected %s to contain a valid hostname, got: %s", k, v))
	case strings.HasPrefix(v, "."), strings.HasSuffix(v, "."):
		errors = append(errors, fmt.Errorf("expected %s to contain a valid hostname, got: %s", k, v))
	}

	return warnings, errors
}

var lowercase = regexp.MustCompile(`^[a-z0-9-]+$`)

func IsNodeName(i any, k string) (warnings []string, errors []error) {
	v, ok := i.(string)

	switch {
	case !ok:
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
	case !lowercase.MatchString(v):
		errors = append(errors, fmt.Errorf("expected %s to contain only lowercase letters, numbers, and dashes, got: %s", k, v))
	case strings.HasPrefix(v, "-"), strings.HasSuffix(v, "-"):
		errors = append(errors, fmt.Errorf("expected %s to not start or end with a dash, but got: %s", k, v))
	}

	return warnings, errors
}
