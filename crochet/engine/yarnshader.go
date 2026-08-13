package engine

import (
	"math"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/SoyStudios/moonshot/crochet/yarn"
)

// GPU lighting + instanced batching.
//
// Yarn is drawn as two primitives: a unit cylinder (segments) and a unit sphere
// (joints/tips), both carrying vertex normals. Rather than issue one draw call
// per segment, the renderer *batches*: every segment/sphere is accumulated as a
// per-instance transform, grouped by (colour, material). At the end of the
// frame each group is drawn with a single rl.DrawMeshInstanced call, so a whole
// fabric costs only a handful of draw calls (one per distinct colour) instead
// of thousands.
//
// A GLSL material does per-fragment diffuse + specular shading; the per-instance
// model matrix arrives through the instanceTransform vertex attribute, and the
// group's colour/sheen/ambient come in as uniforms.
//
// If the shader fails to compile (rare, on a limited driver) the renderer falls
// back to raylib's flat immediate-mode cylinders so the demo still runs.

// The vertex shader also derives, per fragment, where on the tube the fragment
// sits: the angle around it and the distance along it. Those feed a helical
// "ply" phase so the fragment shader can carve the twisted-fibre look of yarn.
const yarnVertexShader = `
#version 330
layout(location = 0) in vec3 vertexPosition;
layout(location = 2) in vec3 vertexNormal;
layout(location = 6) in mat4 instanceTransform;
uniform mat4 mvp;
out vec3 fragPosition;
out vec3 fragNormal;
out vec3 fragTangent;
out float fragPhase;

const float PLIES = 3.0;  // visible twisted strands around the yarn
const float TWIST = 5.5;  // how fast the plies spiral along the length

void main() {
    mat3 m3 = mat3(instanceTransform);
    vec4 world = instanceTransform*vec4(vertexPosition, 1.0);
    fragPosition = world.xyz;
    fragNormal = normalize(transpose(inverse(m3))*vertexNormal);

    // Around-tube tangent (object space) carried to world space.
    vec3 tanObj = normalize(vec3(-vertexPosition.z, 0.0, vertexPosition.x));
    fragTangent = normalize(m3*tanObj);

    // Helical ply phase: angle around the tube + twist along its length,
    // measured in radius units so the twist looks the same at any thickness.
    float radius = length(instanceTransform[0].xyz);
    float segLen = length(instanceTransform[1].xyz);
    float along = vertexPosition.y*segLen;
    float ang = atan(vertexPosition.z, vertexPosition.x);
    fragPhase = ang*PLIES + (along/max(radius, 1e-4))*TWIST;

    gl_Position = mvp*world;
}
`

const yarnFragmentShader = `
#version 330
in vec3 fragPosition;
in vec3 fragNormal;
in vec3 fragTangent;
in float fragPhase;
uniform vec4 colDiffuse;   // per-group yarn colour (raylib sets from material)
uniform vec3 lightDir;     // direction toward the key light (normalized)
uniform vec3 lightColor;
uniform float ambient;     // base brightness in shadow
uniform float sheen;       // specular strength (matte 0 .. glossy 1)
uniform vec3 viewPos;
out vec4 finalColor;

void main() {
    // Fake the plied-fibre relief: perturb the normal across the tube so the
    // light catches the ridges, and darken the grooves between plies.
    float g = sin(fragPhase);
    float gc = cos(fragPhase);
    vec3 N = normalize(fragNormal + fragTangent*(gc*0.40));

    vec3 L = normalize(lightDir);
    float diff = max(dot(N, L), 0.0);
    vec3 V = normalize(viewPos - fragPosition);
    vec3 H = normalize(L + V);
    float spec = pow(max(dot(N, H), 0.0), 20.0)*sheen*(0.5 + 0.5*gc);

    float ao = 0.78 + 0.22*(0.5 + 0.5*g);   // groove shadowing between plies
    vec3 lit = colDiffuse.rgb*(ambient + (1.0 - ambient)*diff)*ao + spec*lightColor;
    finalColor = vec4(lit, colDiffuse.a);
}
`

// batchKey groups instances that can be drawn together: same colour and same
// material response.
type batchKey struct {
	col            yarn.Color
	sheen, ambient float32
}

type yarnRenderer struct {
	lit      bool
	shader   rl.Shader
	material rl.Material
	cylinder rl.Mesh
	sphere   rl.Mesh

	locLightDir   int32
	locLightColor int32
	locAmbient    int32
	locSheen      int32
	locViewPos    int32

	curSheen, curAmbient float32
	cylBatches           map[batchKey][]rl.Matrix
	sphBatches           map[batchKey][]rl.Matrix
}

// newYarnRenderer builds the GPU resources. Must be called after InitWindow.
func newYarnRenderer() *yarnRenderer {
	r := &yarnRenderer{
		cylinder:   rl.GenMeshCylinder(1, 1, 12),
		sphere:     rl.GenMeshSphere(1, 8, 12),
		cylBatches: map[batchKey][]rl.Matrix{},
		sphBatches: map[batchKey][]rl.Matrix{},
	}
	r.shader = rl.LoadShaderFromMemory(yarnVertexShader, yarnFragmentShader)
	r.locLightDir = rl.GetShaderLocation(r.shader, "lightDir")
	r.lit = r.locLightDir >= 0
	if !r.lit {
		return r
	}
	r.locLightColor = rl.GetShaderLocation(r.shader, "lightColor")
	r.locAmbient = rl.GetShaderLocation(r.shader, "ambient")
	r.locSheen = rl.GetShaderLocation(r.shader, "sheen")
	r.locViewPos = rl.GetShaderLocation(r.shader, "viewPos")

	// Route the per-instance model matrix through the instanceTransform
	// attribute (this is what turns on GPU instancing in DrawMeshInstanced).
	locs := unsafe.Slice(r.shader.Locs, 32)
	locs[rl.ShaderLocMatrixModel] = rl.GetShaderLocationAttrib(r.shader, "instanceTransform")

	r.material = rl.LoadMaterialDefault()
	r.material.Shader = r.shader

	toLight := rl.Vector3Normalize(rl.NewVector3(0.45, 1.0, 0.35))
	rl.SetShaderValue(r.shader, r.locLightDir, []float32{toLight.X, toLight.Y, toLight.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(r.shader, r.locLightColor, []float32{1.0, 0.98, 0.94}, rl.ShaderUniformVec3)
	return r
}

// beginFrame updates the view position and resets the per-frame batches.
func (r *yarnRenderer) beginFrame(camPos rl.Vector3) {
	if !r.lit {
		return
	}
	rl.SetShaderValue(r.shader, r.locViewPos, []float32{camPos.X, camPos.Y, camPos.Z}, rl.ShaderUniformVec3)
	for k := range r.cylBatches {
		r.cylBatches[k] = r.cylBatches[k][:0]
	}
	for k := range r.sphBatches {
		r.sphBatches[k] = r.sphBatches[k][:0]
	}
}

// setMaterial selects the sheen/ambient for the geometry about to be queued.
func (r *yarnRenderer) setMaterial(m yarn.Material) {
	r.curSheen = float32(m.Sheen)
	r.curAmbient = float32(m.Ambient)
}

// segment queues a yarn section from a to b as a cylinder of the given radius.
func (r *yarnRenderer) segment(a, b rl.Vector3, radius float32, col yarn.Color) {
	if !r.lit {
		rl.DrawCylinderEx(a, b, radius, radius, 6, toRL(col))
		return
	}
	k := batchKey{col, r.curSheen, r.curAmbient}
	r.cylBatches[k] = append(r.cylBatches[k], cylinderTransform(a, b, radius))
}

// node queues a rounded joint sphere.
func (r *yarnRenderer) node(pos rl.Vector3, radius float32, col yarn.Color) {
	if !r.lit {
		rl.DrawSphere(pos, radius, toRL(col))
		return
	}
	k := batchKey{col, r.curSheen, r.curAmbient}
	tf := rl.MatrixMultiply(
		rl.MatrixScale(radius, radius, radius),
		rl.MatrixTranslate(pos.X, pos.Y, pos.Z),
	)
	r.sphBatches[k] = append(r.sphBatches[k], tf)
}

// flush draws all queued geometry as a handful of instanced draw calls.
func (r *yarnRenderer) flush() {
	if !r.lit {
		return
	}
	for k, mats := range r.cylBatches {
		if len(mats) > 0 {
			r.drawBatch(r.cylinder, k, mats)
		}
	}
	for k, mats := range r.sphBatches {
		if len(mats) > 0 {
			r.drawBatch(r.sphere, k, mats)
		}
	}
}

func (r *yarnRenderer) drawBatch(mesh rl.Mesh, k batchKey, mats []rl.Matrix) {
	rl.SetShaderValue(r.shader, r.locAmbient, []float32{k.ambient}, rl.ShaderUniformFloat)
	rl.SetShaderValue(r.shader, r.locSheen, []float32{k.sheen}, rl.ShaderUniformFloat)
	r.material.GetMap(rl.MapDiffuse).Color = toRL(k.col)
	rl.DrawMeshInstanced(mesh, r.material, mats, len(mats))
}

func (r *yarnRenderer) unload() {
	rl.UnloadMesh(&r.cylinder)
	rl.UnloadMesh(&r.sphere)
	if r.lit {
		rl.UnloadShader(r.shader)
	}
}

// cylinderTransform maps the unit cylinder (radius 1, base at the origin,
// extending up +Y by 1) so it spans a→b with the given radius.
func cylinderTransform(a, b rl.Vector3, radius float32) rl.Matrix {
	dir := rl.Vector3Subtract(b, a)
	length := float32(math.Sqrt(float64(dir.X*dir.X + dir.Y*dir.Y + dir.Z*dir.Z)))
	if length < 1e-6 {
		length = 1e-6
	}
	scale := rl.MatrixScale(radius, length, radius)
	rot := alignYTo(dir, length)
	trans := rl.MatrixTranslate(a.X, a.Y, a.Z)
	return rl.MatrixMultiply(rl.MatrixMultiply(scale, rot), trans)
}

// alignYTo returns a rotation that turns +Y onto the (length-known) direction.
func alignYTo(dir rl.Vector3, length float32) rl.Matrix {
	d := rl.NewVector3(dir.X/length, dir.Y/length, dir.Z/length)
	axis := rl.Vector3CrossProduct(rl.NewVector3(0, 1, 0), d)
	axisLen := float32(math.Sqrt(float64(axis.X*axis.X + axis.Y*axis.Y + axis.Z*axis.Z)))
	if axisLen < 1e-6 {
		if d.Y >= 0 {
			return rl.MatrixScale(1, 1, 1) // identity
		}
		return rl.MatrixRotate(rl.NewVector3(1, 0, 0), math.Pi)
	}
	angle := float32(math.Acos(clampf(float64(d.Y), -1, 1)))
	return rl.MatrixRotate(rl.Vector3Normalize(axis), angle)
}

func clampf(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func toRL(c yarn.Color) rl.Color { return rl.NewColor(c.R, c.G, c.B, c.A) }

func darker(c yarn.Color, f float64) yarn.Color {
	return yarn.Color{
		R: uint8(float64(c.R) * f),
		G: uint8(float64(c.G) * f),
		B: uint8(float64(c.B) * f),
		A: c.A,
	}
}
