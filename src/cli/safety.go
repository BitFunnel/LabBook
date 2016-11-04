package cli

import (
	"errors"
	"fmt"
	"os/user"
	"strings"

	"regexp"
)

func checkCurrentUserSafe() error {
	// TODO: Extend to Windows.
	user, userErr := user.Current()
	if userErr != nil {
		return errors.New("Could not verify user was not root. To skip this " +
			"check, run with the --allow-sudo flag")
	}

	// TODO: Add --allow-sudo flag.
	if user.Username == "root" || user.Name == "System Administrator" {
		return fmt.Errorf("To avoid destroying important files, running as " +
			"root or administrator is disabled by default. To enable, run " +
			"with the --allow-sudo flag")
	}

	return nil
}

func checkDirectorySafe(path string) error {
	unsafe, pathCheckErr := pathUnsafe(path)
	if pathCheckErr != nil {
		return pathCheckErr
	} else if unsafe {
		// TODO: Add --allow-unsafe-paths.
		return fmt.Errorf("Refusing to write to path '%s' because it is a "+
			"systems directory. To enable this, use the flag "+
			"--allow-unsafe-paths", path)
	}

	return nil
}

func pathUnsafe(path string) (bool, error) {
	// TODO: Extend to Windows.
	match, matchErr := regexp.Match(inSystemsDirectory, []byte(path))
	if matchErr != nil {
		return false, matchErr
	}

	return match, nil
}

// Pattern matching a directories that lie in a set of blacklisted directories
// that we refuse to write to.
var inSystemsDirectory string

// TODO: Extend all this to Windows.
func init() {
	// Generate a regex that will check if a path prefixed (i.e., rooted in)
	// one of a set of blacklisted systems directories. The trailing slashes
	// are important here because we want to disallow something like
	// `/bin/cow/` but not `/bincow/`.
	systemsDirectories := []string{
		"/bin/", "/usr/", "/sbin/", "/etc/", "/dev/", "/proc/", "/sys/",
		"/mnt/", "/media/", "/var/", "/lib/", "/boot/",
	}
	const root = "/"

	// NOTE: We're adding the path `/` here, as it does not conform to the
	// regex pattern of the paths above.
	inSystemsDirectory = fmt.Sprintf(
		"(^%s|%s)",
		strings.Join(systemsDirectories, "|^"),
		"^"+root+"$")
}
