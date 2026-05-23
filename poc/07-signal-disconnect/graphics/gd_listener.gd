extends Node

# Connects to sibling Emitter's tick signal in the normal GDScript way:
# `emitter.tick.connect(self._on_tick)`. This produces a Callable BOUND
# to `self` — when this node is freed, Godot should auto-disconnect it.

const FREE_AT_FRAME = 3

func _ready():
	var emitter = get_parent().get_node("Emitter")
	emitter.tick.connect(_on_tick)
	print("[POC7] GdListener._ready — connected to Emitter.tick (self-bound callable)")

func _process(_delta):
	if get_tree().get_frame() >= FREE_AT_FRAME and is_inside_tree():
		print("[POC7] GdListener self-freeing at scene-frame=", get_tree().get_frame())
		queue_free()

func _on_tick(n):
	print("[POC7][GD]  GdListener received Tick(", n, ")")
