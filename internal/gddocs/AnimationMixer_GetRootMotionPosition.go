/*
[gdscript]
var current_rotation

func _process(delta):
	if Input.is_action_just_pressed("animate"):
		current_rotation = get_quaternion()
		state_machine.travel("Animate")
	var velocity = current_rotation * animation_tree.get_root_motion_position() / delta
	set_velocity(velocity)
	move_and_slide()
[/gdscript]
*/

package main

import (
	"github.com/AveryLucas/gogogd/classdb/AnimationNodeStateMachinePlayback"
	"github.com/AveryLucas/gogogd/classdb/AnimationTree"
	"github.com/AveryLucas/gogogd/classdb/CharacterBody3D"
	"github.com/AveryLucas/gogogd/classdb/Input"
	"github.com/AveryLucas/gogogd/variant/Float"
	"github.com/AveryLucas/gogogd/variant/Quaternion"
	"github.com/AveryLucas/gogogd/variant/Vector3"
)

var delta Float.X
var current_rotation Quaternion.IJKX

var animationTree AnimationTree.Instance
var animationNodeStateMachinePlayback AnimationNodeStateMachinePlayback.Instance
var characterBody3D CharacterBody3D.Instance

func AnimationMixer_GetRootMotionPosition() {
	if Input.IsActionJustPressed("animate", false) {
		current_rotation = characterBody3D.AsNode3D().Quaternion()
		animationNodeStateMachinePlayback.Travel("Animate")
	}
	var velocity = Vector3.DivX(Quaternion.Rotate(animationTree.AsAnimationMixer().GetRootMotionPosition(), current_rotation), delta)
	characterBody3D.SetVelocity(velocity)
	characterBody3D.MoveAndSlide()
}
