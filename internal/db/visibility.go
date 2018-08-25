package db

import (
	"errors"
	"fmt"
)

type Visibility string

const (
	// VisibilityIgnore shows nothing
	VisibilityIgnore Visibility = "ignore"
	// VisibilityAnon doesn't show the name, but increments the anonymous user count
	VisibilityAnon Visibility = "anon"
	// VisibilityUser shows the user, but not the device name(s)
	VisibilityUser Visibility = "user"
	// VisibilityAll shows user and the device names
	VisibilityAll Visibility = "all"

	VisibilityInfrastructure Visibility = "infrastructure"

	VisibilityDeprecatedInfrastructure Visibility = "deprecated-infrastructure"

	VisibilityUserInfrastructure Visibility = "user-infrastructure"

	VisibilityImportantInfrastructure Visibility = "important-infrastructure"

	VisibilityCriticalInfrastructure Visibility = "critical-infrastructure"
)

// don't forget all Visibilities to this array...
var validVsibilities = [...]Visibility{VisibilityIgnore, VisibilityAnon, VisibilityUser, VisibilityAll,
	VisibilityInfrastructure, VisibilityDeprecatedInfrastructure, VisibilityUserInfrastructure, VisibilityImportantInfrastructure, VisibilityCriticalInfrastructure}

func (v *Visibility) UnmarshalJSON(byteValue []byte) error {
	if len(byteValue) < 2 {
		return errors.New("Visibility must be a JSON string value.")
	}
	value := Visibility(byteValue[1 : len(byteValue)-1])
	for _, validV := range validVsibilities {
		if validV == value {
			*v = value
			return nil
		}
	}

	return fmt.Errorf("Visibility was '%s' but must be one of %s", value, validVsibilities)
}
