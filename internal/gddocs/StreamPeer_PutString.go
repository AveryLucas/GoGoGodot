/*
[gdscript]
put_data("Hello world".to_ascii_buffer())
[/gdscript]
[csharp]
PutData("Hello World".ToAsciiBuffer());
[/csharp]
*/

package main

import "github.com/AveryLucas/gogogd/classdb/StreamPeer"

var streamPeer StreamPeer.Instance

func StreamPeer_PutString() {
	streamPeer.PutData([]byte("Hello world"))
}
