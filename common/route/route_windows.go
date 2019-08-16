package route

import (
	"errors"
	"fmt"
	"os/exec"
)

func AddRoute(dest, netmask, gateway string) error {
	out, err := exec.Command("route", "add", dest, "mask", netmask, gateway, "metric", "5").Output()
	if err != nil {
		if len(out) != 0 {
			return errors.New(fmt.Sprintf("%v, output: %s", err, out))
		}
		return err
	}
	return nil
}
