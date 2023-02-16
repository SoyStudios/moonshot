extends Node

@export
var bot_scene : PackedScene
@onready
var camera = $Camera2D
var cameraScrollSpeed : float = 10
var cameraZoomStep : float = .1

# Called when the node enters the scene tree for the first time.
func _ready():
	var bot1 = bot_scene.instantiate()
	bot1.position.x = 100
	bot1.position.y = 100
	var bot2 = bot_scene.instantiate()
	bot2.position.x = -1000
	bot2.position.y = 1000
	add_child(bot1)
	add_child(bot2)

# getScroll returns for how much we should scroll
# based on the current zoom factor
#
# the more zoomed out we are, the more we move
func getScroll():
	return cameraScrollSpeed / camera.zoom.x

# Called every frame. 'delta' is the elapsed time since the previous frame.
func _process(delta):
	if Input.is_action_pressed("scroll up"):
		camera.position.y -= getScroll()
	if Input.is_action_pressed("scroll left"):
		camera.position.x -= getScroll()
	if Input.is_action_pressed("scroll down"):
		camera.position.y += getScroll()
	if Input.is_action_pressed("scroll right"):
		camera.position.x += getScroll()
	if Input.is_action_pressed("zoom in"):
		camera.zoom.x += cameraZoomStep
		camera.zoom.y += cameraZoomStep
	if Input.is_action_pressed("zoom out"):
		camera.zoom.x -= cameraZoomStep
		camera.zoom.y -= cameraZoomStep
