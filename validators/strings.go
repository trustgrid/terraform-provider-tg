package validators

import (
	"fmt"
	"strings"

	"github.com/asaskevich/govalidator"
)

func IsHostname(i interface{}, k string) (warnings []string, errors []error) {
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
