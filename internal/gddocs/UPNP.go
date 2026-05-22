/*
var upnp = UPNP.new()
upnp.discover()
upnp.add_port_mapping(7777)
*/

package main

import "github.com/AveryLucas/gogogd/classdb/UPNP"

func ExampleUPNP() {
	var upnp = UPNP.New()
	upnp.Discover()
	upnp.AddPortMapping(7777)
}
