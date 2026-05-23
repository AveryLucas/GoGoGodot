extends Node

# Connects to all three Go-declared signals on the sibling Probe.
# Prints what it receives so we can verify marshaling.

func _ready():
	var probe = get_parent().get_node("Probe")

	# Verify the engine sees the signals (and at the snake_case names we expect).
	print("[GD5] has_signal('ping')    = ", probe.has_signal("ping"))
	print("[GD5] has_signal('damaged') = ", probe.has_signal("damaged"))
	print("[GD5] has_signal('moved')   = ", probe.has_signal("moved"))
	print("[GD5] has_signal('Ping')    = ", probe.has_signal("Ping"))

	probe.ping.connect(_on_ping)
	probe.damaged.connect(_on_damaged)
	probe.moved.connect(_on_moved)
	print("[GD5] connected to ping / damaged / moved")

func _on_ping():
	print("[GD5][GDS] ping received")

func _on_damaged(amount):
	print("[GD5][GDS] damaged received: amount=", amount)

func _on_moved(who, where):
	print("[GD5][GDS] moved received: who=", who, " where=", where)
