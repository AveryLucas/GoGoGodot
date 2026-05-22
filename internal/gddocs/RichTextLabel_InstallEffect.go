/*
# effect.gd
class_name MyCustomEffect
extends RichTextEffect

var bbcode = "my_custom_effect"

# ...
*/

package main

import "github.com/AveryLucas/gogogd/classdb/RichTextEffect"

type MyCustomEffect struct {
	RichTextEffect.Extension[MyCustomEffect]

	Bbcode string
}

func NewMyCustomEffect() *MyCustomEffect {
	return &MyCustomEffect{
		Bbcode: "my_custom_effect",
	}
}
