/*
extends AudioEffect

@export var strength = 4.0

func _instantiate():
	var effect = CustomAudioEffectInstance.new()
	effect.base = self

	return effect
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/AudioEffect"
	"github.com/AveryLucas/gogogd/classdb/AudioEffectAmplify"
	"github.com/AveryLucas/gogogd/variant/Object"
)

type MyAudioEffect struct {
	AudioEffect.Extension[MyAudioEffect]
}

func (e *MyAudioEffect) Instantiate() AudioEffect.Instance {
	effect := AudioEffectAmplify.New()
	Object.Set(effect, "base", e)
	return effect.AsAudioEffect()
}
