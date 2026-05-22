/*
upnp.delete_port_mapping(port)
*/

package main

import "github.com/AveryLucas/gogogd/classdb/UPNP"

func ExampleUPNP_Delete(upnp UPNP.Instance, port int) {
	upnp.DeletePortMapping(port)
}
