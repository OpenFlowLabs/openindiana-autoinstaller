package installd

import "fmt"

func NotSupportedError(functionality string) error {
	return fmt.Errorf("functionality %s is currently not Supported", functionality)
}

func InvalidConfiguration(section string) error {
	return fmt.Errorf("invalid Configuration %s is not correct please fix", section)
}
