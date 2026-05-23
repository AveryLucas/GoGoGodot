extends Node

# Calls every exported method on the sibling Probe node and prints the result.
# Output is prefixed [GD3] to disambiguate from Go's [POC3] traces.
# Run order matters: this script's _ready fires *after* Probe.Ready (the Go
# side's stderr trace confirms ordering).

func _ready():
	var probe = get_parent().get_node("Probe")
	print("[GD3] caller _ready — probe=", probe)

	# 1. PascalCase Go names should expose as snake_case in GDScript.
	print("[GD3] has_method('add_ints') = ", probe.has_method("add_ints"))
	print("[GD3] has_method('AddInts')  = ", probe.has_method("AddInts"))
	print("[GD3] has_method('as_hidden')= ", probe.has_method("as_hidden"))

	# 2. int / int -> int
	print("[GD3] add_ints(3, 4) = ", probe.add_ints(3, 4))

	# 3. string -> string
	print("[GD3] greet('world') = ", probe.greet("world"))

	# 4. Vector2 + float -> Vector2
	var v = probe.scale_vec(Vector2(2.0, 3.0), 1.5)
	print("[GD3] scale_vec((2,3), 1.5) = ", v)

	# 5. Array[int] -> int
	print("[GD3] sum_slice([10,20,30]) = ", probe.sum_slice([10, 20, 30]))

	# 6. bool -> bool
	print("[GD3] toggle(true) = ", probe.toggle(true))

	# 7. void return
	probe.no_return("from GDScript")
	print("[GD3] no_return called")

	# 8. Exported field readable as snake_case property?
	print("[GD3] probe.touch_count = ", probe.touch_count)
	print("[GD3] probe.TouchCount  = ", probe.TouchCount)

	print("[GD3] DONE")
