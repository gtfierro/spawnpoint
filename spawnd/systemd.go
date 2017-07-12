package main

import (
	"fmt"
	"io/ioutil"

	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

// SystemD
// Given a docker image name, we want to build up a systemd unit file that we can use to manage
// the running docker container

// [Unit]
// Description=<name here>
//
// [Service]
// Restart=always
// RestartSec=5s
// ExecStart=/usr/bin/docker run \
//     --name %p \
//     -v /etc/hod:/etc/hod \
//     -p 80:80 \
//     <container name from spawnpoint>
//
// ExecStop=/usr/bin/docker stop -t 5 %p ; /usr/bin/docker rm -f %p
//
// [Install]
// WantedBy=multi-user.target

func MakeSystemD(image_name string, container *docker.Container) error {
	var options []*unit.UnitOption

	var opt *unit.UnitOption

	// [Unit]
	opt = unit.NewUnitOption("Unit", "Description", container.Name)
	options = append(options, opt)

	// [Service]
	opt = unit.NewUnitOption("Service", "Restart", "always")
	options = append(options, opt)
	opt = unit.NewUnitOption("Service", "RestartSec", "5s")
	options = append(options, opt)

	fmt.Printf("%+v\n", container)

	execstart := fmt.Sprintf("/usr/bin/docker run --name %%p -e BW2_AGENT=172.17.0.1:28589 -e BW2_DEFAULT_ENTITY=entity.key %s ./svcexe", image_name)
	opt = unit.NewUnitOption("Service", "ExecStart", execstart)
	options = append(options, opt)
	opt = unit.NewUnitOption("Service", "ExecStop", "/usr/bin/docker stop -t 5 %p ; /usr/bin/docker rm -f %p")
	options = append(options, opt)

	// [Install]
	opt = unit.NewUnitOption("Install", "WantedBy", "multi-user.target")
	options = append(options, opt)

	rdr := unit.Serialize(options)

	unitfile, _ := ioutil.ReadAll(rdr)
	unitfilename := fmt.Sprintf("/etc/systemd/system/%s.service", container.Name)
	err := ioutil.WriteFile(unitfilename, unitfile, 0644)
	if err != nil {
		fmt.Println(errors.Wrapf(err, "Failed to make systemd file for %s", container.Name))
		return err
	}

	// connect to dbus
	conn, err := dbus.New()
	if err != nil {
		fmt.Println(errors.Wrap(err, "Could not open systemd dbus conn"))
		return err
	}
	err = conn.Reload() // scan for new unit files
	if err != nil {
		fmt.Println(errors.Wrap(err, "Could not reload unit files"))
		return err
	}

	return nil
}
