# 2023-02-04

Picking up where I left off.

## Re-read code and add comments

After this long hiatus, it might make sense to dive deeply into the code
and see what I can surmise.

## Test scenarios

Build out scenarios to test if the stack machine works as expected with
the environment.

## Asteroids

This is where I left off the last time. Problems I need to address:

* How to procedurally generate asteroids
* How does rigid body physics relate to asteroids breaking
* How Mining relates to deformation/mass changes

### Procedurally generating asteroids

Next attempt:

https://cglab.ca/~sander/misc/ConvexGeneration/convex.html

Much simpler idea (sometimes the simplest ideas don't come to us):

Pretty much: draw a circle with random steps and random radii.

Too "spikey".

Current idea:

Start with a line with random (clamped) length and angle.
Add another line at the end of the previous.
Repeat until sum of all angles higher than threshold (to identify closed).
Close the loop.

Now, find all lines with length larger than threshold.
Subdivide lines.

### Breaking asteroids

https://stackoverflow.com/questions/3623703/how-can-i-split-a-polygon-by-a-line

Subdivide new parts.

## Notes

Demo bot code:

BEGIN EV
	// Read bot’s current energy level and push it to the stack
	RDE
	PSH CON 1000
	GEQ
END
BEGIN EX
	PSH CON 500
	REP
END

BEGIN EV
	// If total velocity is >= 200
	RDX
	RDY
	ABS
	PSH CON 200
	GEQ
END
BEGIN EX
	PSH CON 0
	POP REG 0
	// thrust in opposite direction
	RDX
	NEG
	RDY
	NEG
	THR
END

// Counter for turning
// Register 0 holds counter
BEGIN EV
	// If reg0 <= 80
	PSH REG 0
	PSH CON 80
	LEQ
END
BEGIN EX
	// reg0++
	PSH REG 0
	PSH CON 1
	ADD
	POP REG 0
END

// Turning every 80 ticks
BEGIN EV
	// if reg0 > 80
	PSH REG 0
	PSH CON 80
	GRT
END
BEGIN EX
	// reset reg0 to 0
	// turn by 10 degrees
	PSH CON 0
	POP REG 0
	PSH CON 10
	TRN
	PSH CON 500
	IMP
END

// Create impulse every 20 ticks
// counter in reg1
BEGIN EV
	// if reg1 <= 20
	PSH REG 1
	PSH CON 20
	LEQ
END
BEGIN EX
	// reg1++
	PSH REG 1
	PSH CON 1
	ADD
	POP REG 1
END

// Create impulse in current direction
BEGIN EV
	// if reg1 > 20
	PSH REG 1
	PSH CON 20
	GRT
END
BEGIN EX
	PSH CON 500
	IMP
	PSH CON 0
	POP REG 1
END
