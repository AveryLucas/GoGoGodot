extends Node

# Inspects the Player class to verify the wrapper layer didn't disrupt Godot's
# class hierarchy view.

func _ready():
	var player = get_parent().get_node("Player")

	# 1. What does Godot think Player's class is?
	print("[POC8][GD] player.get_class() = ", player.get_class())

	# 2. Is Player considered a Node2D? (Should be true — Node2DExt extends
	#    Node2D.Extension[T] which extends Node2D.)
	print("[POC8][GD] player is Node2D    = ", player is Node2D)
	print("[POC8][GD] player is Node      = ", player is Node)
	print("[POC8][GD] player is Sprite2D  = ", player is Sprite2D)  # should be false

	# 3. Is the user method exposed at the right name?
	print("[POC8][GD] has_method('take_damage') = ", player.has_method("take_damage"))

	# 4. Can we call it and get the return?
	var new_hp = player.take_damage(7)
	print("[POC8][GD] take_damage(7) returned ", new_hp)

	# 5. Are forwarder methods on the wrapper EXPOSED to GDScript?
	#    They shouldn't be — they're Go-side conveniences. Godot's Node2D
	#    already gives GDScript its own position accessors.
	print("[POC8][GD] has_method('set_position') = ", player.has_method("set_position"))
	#    (Always true; that's a Node2D built-in, not the wrapper's.)

	# 6. Are exported fields readable as properties?
	print("[POC8][GD] player.hp = ", player.hp)

	# 7. Is the Node2D-level property visible?
	print("[POC8][GD] player.position = ", player.position)
