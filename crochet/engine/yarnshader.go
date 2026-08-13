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
uniform float sheen;       // sheen strength (matte 0 .. glossy 1)
uniform vec3 viewPos;
out vec4 finalColor;

// Cheap coherent value noise for the fuzzy micro-surface.
float hash(vec3 p) {
    p = fract(p*vec3(0.1031, 0.1030, 0.0973));
    p += dot(p, p.yzx + 33.33);
    return fract((p.x + p.y)*p.z);
}
float vnoise(vec3 p) {
    vec3 i = floor(p), f = fract(p);
    f = f*f*(3.0 - 2.0*f);
    return mix(mix(mix(hash(i + vec3(0,0,0)), hash(i + vec3(1,0,0)), f.x),
                   mix(hash(i + vec3(0,1,0)), hash(i + vec3(1,1,0)), f.x), f.y),
               mix(mix(hash(i + vec3(0,0,1)), hash(i + vec3(1,0,1)), f.x),
                   mix(hash(i + vec3(0,1,1)), hash(i + vec3(1,1,1)), f.x), f.y), f.z);
}

void main() {
    vec3 N = normalize(fragNormal);
    vec3 T = normalize(fragTangent);          // around the tube
    vec3 B = normalize(cross(N, T));          // along the tube

    // Coarse plied-fibre relief across the tube.
    float g  = sin(fragPhase);
    float gc = cos(fragPhase);
    N = normalize(N + T*(gc*0.28));

    // Fuzzy micro-roughness: jitter the normal with fine coherent noise so the
    // surface scatters light like spun fibre instead of moulded plastic.
    float nA = vnoise(fragPosition*38.0) - 0.5;
    float nB = vnoise(fragPosition*38.0 + 19.3) - 0.5;
    N = normalize(N + (T*nA + B*nB)*0.6);

    vec3 L = normalize(lightDir);
    vec3 V = normalize(viewPos - fragPosition);

    // Soft wrapped diffuse: fibres scatter light around the terminator, so the
    // shading is matte with no hard shadow edge.
    float diff = clamp(dot(N, L)*0.5 + 0.5, 0.0, 1.0);

    // Fabric sheen: a dim, broad highlight plus a faint grazing rim — never the
    // tight hot spot of plastic.
    vec3 H = normalize(L + V);
    float broad = pow(max(dot(N, H), 0.0), 5.0);
    float rim = pow(1.0 - max(dot(normalize(fragNormal), V), 0.0), 3.0);
    float spec = (broad*0.10 + rim*0.12)*sheen;

    float ao   = 0.82 + 0.18*(0.5 + 0.5*g);        // groove shadowing
    float tint = 0.93 + 0.07*(nA + 0.5);           // subtle colour variation
    vec3 lit = colDiffuse.rgb*tint*(ambient + (1.0 - ambient)*diff)*ao + spec*lightColor;
    finalColor = vec4(lit, colDiffuse.a);
}
`

// The halo shader draws "shell fur": the yarn geometry re-drawn a few times,
// each shell pushed out along the normal, with a noise mask that keeps fewer
// fragments the further out it goes — so sparse fibre tips stick out around the
// silhouette and fuzz the surface.
const haloVertexShader = `
#version 330
layout(location = 0) in vec3 vertexPosition;
layout(location = 2) in vec3 vertexNormal;
layout(location = 6) in mat4 instanceTransform;
uniform mat4 mvp;
uniform float shellOffset;   // push out along the normal (object units)
uniform float shellFrac;     // 0..1 how far out this shell is
out vec3 fragBase;           // un-expanded world pos, for stable fibre noise
out vec3 fragNormal;
out vec3 fragPosition;
out float vFrac;
void main() {
    mat3 m3 = mat3(instanceTransform);
    vec3 posExp = vertexPosition + vertexNormal*shellOffset;
    vec4 world = instanceTransform*vec4(posExp, 1.0);
    fragBase = (instanceTransform*vec4(vertexPosition, 1.0)).xyz;
    fragNormal = normalize(transpose(inverse(m3))*vertexNormal);
    fragPosition = world.xyz;
    vFrac = shellFrac;
    gl_Position = mvp*world;
}
`

const haloFragmentShader = `
#version 330
in vec3 fragBase;
in vec3 fragNormal;
in vec3 fragPosition;
in float vFrac;
uniform vec4 colDiffuse;
uniform vec3 lightDir;
uniform float ambient;
out vec4 finalColor;

float hash(vec3 p) {
    p = fract(p*vec3(0.1031, 0.1030, 0.0973));
    p += dot(p, p.yzx + 33.33);
    return fract((p.x + p.y)*p.z);
}
float vnoise(vec3 p) {
    vec3 i = floor(p), f = fract(p);
    f = f*f*(3.0 - 2.0*f);
    return mix(mix(mix(hash(i + vec3(0,0,0)), hash(i + vec3(1,0,0)), f.x),
                   mix(hash(i + vec3(0,1,0)), hash(i + vec3(1,1,0)), f.x), f.y),
               mix(mix(hash(i + vec3(0,0,1)), hash(i + vec3(1,0,1)), f.x),
                   mix(hash(i + vec3(0,1,1)), hash(i + vec3(1,1,1)), f.x), f.y), f.z);
}

void main() {
    // A fibre reaches this shell only if its noise "length" clears the shell's
    // height; the threshold rises outward so tips thin to a sparse fuzz.
    float n = vnoise(fragBase*46.0);
    if (n < 0.34 + vFrac*0.6) discard;

    vec3 N = normalize(fragNormal);
    vec3 L = normalize(lightDir);
    float diff = clamp(dot(N, L)*0.5 + 0.5, 0.0, 1.0);
    vec3 col = colDiffuse.rgb*(ambient + (1.0 - ambient)*diff)*1.15;
    finalColor = vec4(col, (1.0 - vFrac)*0.55);
}
`

// haloShell is one fur layer: how far out (object units) and its 0..1 fraction.
type haloShell struct{ offset, frac float32 }

var haloShells = []haloShell{{0.55, 0.5}, {1.3, 1.0}}

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

	// Fuzzy-fibre halo (shell fur).
	fuzz         bool
	haloShader   rl.Shader
	haloMaterial rl.Material
	hLightDir    int32
	hAmbient     int32
	hShellOffset int32
	hShellFrac   int32

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

	// Fuzzy-fibre halo shader.
	r.fuzz = true
	r.haloShader = rl.LoadShaderFromMemory(haloVertexShader, haloFragmentShader)
	r.hLightDir = rl.GetShaderLocation(r.haloShader, "lightDir")
	r.hAmbient = rl.GetShaderLocation(r.haloShader, "ambient")
	r.hShellOffset = rl.GetShaderLocation(r.haloShader, "shellOffset")
	r.hShellFrac = rl.GetShaderLocation(r.haloShader, "shellFrac")
	hlocs := unsafe.Slice(r.haloShader.Locs, 32)
	hlocs[rl.ShaderLocMatrixModel] = rl.GetShaderLocationAttrib(r.haloShader, "instanceTransform")
	r.haloMaterial = rl.LoadMaterialDefault()
	r.haloMaterial.Shader = r.haloShader
	rl.SetShaderValue(r.haloShader, r.hLightDir, []float32{toLight.X, toLight.Y, toLight.Z}, rl.ShaderUniformVec3)
	return r
}

// toggleFuzz turns the fibre halo on or off.
func (r *yarnRenderer) toggleFuzz() { r.fuzz = !r.fuzz }

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

// flush draws all queued geometry as a handful of instanced draw calls,
// followed by the translucent fibre-halo shells.
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
	if r.fuzz {
		r.drawHalo()
	}
}

func (r *yarnRenderer) drawBatch(mesh rl.Mesh, k batchKey, mats []rl.Matrix) {
	rl.SetShaderValue(r.shader, r.locAmbient, []float32{k.ambient}, rl.ShaderUniformFloat)
	rl.SetShaderValue(r.shader, r.locSheen, []float32{k.sheen}, rl.ShaderUniformFloat)
	r.material.GetMap(rl.MapDiffuse).Color = toRL(k.col)
	rl.DrawMeshInstanced(mesh, r.material, mats, len(mats))
}

// drawHalo re-draws the batched geometry as a few expanding, translucent shells
// so a fuzzy fibre halo forms around the yarn. Depth writes are disabled so the
// shells blend instead of occluding each other.
func (r *yarnRenderer) drawHalo() {
	rl.BeginBlendMode(rl.BlendAlpha)
	rl.DisableDepthMask()
	for _, s := range haloShells {
		rl.SetShaderValue(r.haloShader, r.hShellOffset, []float32{s.offset}, rl.ShaderUniformFloat)
		rl.SetShaderValue(r.haloShader, r.hShellFrac, []float32{s.frac}, rl.ShaderUniformFloat)
		for k, mats := range r.cylBatches {
			if len(mats) > 0 {
				r.drawHaloBatch(r.cylinder, k, mats)
			}
		}
		for k, mats := range r.sphBatches {
			if len(mats) > 0 {
				r.drawHaloBatch(r.sphere, k, mats)
			}
		}
	}
	rl.EnableDepthMask()
	rl.EndBlendMode()
}

func (r *yarnRenderer) drawHaloBatch(mesh rl.Mesh, k batchKey, mats []rl.Matrix) {
	rl.SetShaderValue(r.haloShader, r.hAmbient, []float32{k.ambient}, rl.ShaderUniformFloat)
	r.haloMaterial.GetMap(rl.MapDiffuse).Color = toRL(k.col)
	rl.DrawMeshInstanced(mesh, r.haloMaterial, mats, len(mats))
}

func (r *yarnRenderer) unload() {
	rl.UnloadMesh(&r.cylinder)
	rl.UnloadMesh(&r.sphere)
	if r.lit {
		rl.UnloadShader(r.shader)
		rl.UnloadShader(r.haloShader)
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
