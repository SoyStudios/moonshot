package engine

import (
	"math"
	"sort"
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

// The fibre detail (fuzz, colour variation) is sampled from a small baked noise
// texture (bound through the material's normal-map slot as texture2) rather than
// computed per fragment, which is far cheaper. A per-instance UV offset keeps
// the tiling from repeating visibly across segments. The helical "ply" phase
// stays analytic — it's only a couple of trig ops.
const yarnVertexShader = `
#version 330
layout(location = 0) in vec3 vertexPosition;
layout(location = 1) in vec2 vertexTexCoord;
layout(location = 2) in vec3 vertexNormal;
layout(location = 6) in mat4 instanceTransform;
uniform mat4 mvp;
out vec3 fragPosition;
out vec3 fragNormal;
out vec3 fragTangent;
out vec2 fragUV;
out float fragPhase;

const float PLIES = 3.0;  // visible twisted strands around the yarn
const float TWIST = 5.5;  // how fast the plies spiral along the length

void main() {
    mat3 m3 = mat3(instanceTransform);
    vec4 world = instanceTransform*vec4(vertexPosition, 1.0);
    fragPosition = world.xyz;
    fragNormal = normalize(transpose(inverse(m3))*vertexNormal);

    vec3 tanObj = normalize(vec3(-vertexPosition.z, 0.0, vertexPosition.x));
    fragTangent = normalize(m3*tanObj);

    float radius = length(instanceTransform[0].xyz);
    float segLen = length(instanceTransform[1].xyz);
    float along = vertexPosition.y*segLen;
    float ang = atan(vertexPosition.z, vertexPosition.x);
    fragPhase = ang*PLIES + (along/max(radius, 1e-4))*TWIST;

    // Per-instance UV jitter so the baked noise doesn't obviously tile. The
    // seed is quantised to a coarse grid so tiny per-frame physics drift can't
    // change it — otherwise the fibre pattern swims and flickers.
    vec3 seed = floor(instanceTransform[3].xyz*4.0);
    float off = fract(sin(dot(seed, vec3(12.9, 78.2, 37.7)))*43758.5);
    fragUV = vertexTexCoord + vec2(off, off*1.7);

    gl_Position = mvp*world;
}
`

const yarnFragmentShader = `
#version 330
in vec3 fragPosition;
in vec3 fragNormal;
in vec3 fragTangent;
in vec2 fragUV;
in float fragPhase;
uniform vec4 colDiffuse;   // per-group yarn colour (raylib sets from material)
uniform vec3 lightDir;     // direction toward the key light (normalized)
uniform vec3 lightColor;
uniform float ambient;     // base brightness in shadow
uniform float sheen;       // sheen strength (matte 0 .. glossy 1)
uniform vec3 viewPos;
uniform sampler2D texture2; // baked fibre noise (material normal-map slot)
out vec4 finalColor;

void main() {
    vec3 N = normalize(fragNormal);
    vec3 T = normalize(fragTangent);          // around the tube
    vec3 B = normalize(cross(N, T));          // along the tube

    // Coarse plied-fibre relief across the tube.
    float g  = sin(fragPhase);
    float gc = cos(fragPhase);
    N = normalize(N + T*(gc*0.28));

    // Fuzzy micro-roughness from the baked noise texture (two lookups for two
    // tangent directions), so the surface scatters light like spun fibre.
    float nA = texture(texture2, fragUV*vec2(3.0, 1.6)).r - 0.5;
    float nB = texture(texture2, fragUV*vec2(3.0, 1.6) + vec2(0.5, 0.27)).r - 0.5;
    N = normalize(N + (T*nA + B*nB)*0.6);

    vec3 L = normalize(lightDir);
    vec3 V = normalize(viewPos - fragPosition);

    // Soft wrapped diffuse (matte, no hard shadow edge).
    float diff = clamp(dot(N, L)*0.5 + 0.5, 0.0, 1.0);

    // Fabric sheen: a dim, broad highlight plus a faint grazing rim.
    vec3 H = normalize(L + V);
    float broad = pow(max(dot(N, H), 0.0), 5.0);
    float rim = pow(1.0 - max(dot(normalize(fragNormal), V), 0.0), 3.0);
    float spec = (broad*0.10 + rim*0.12)*sheen;

    float ao   = 0.82 + 0.18*(0.5 + 0.5*g);
    float tint = 0.93 + 0.07*(nA + 0.5);
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
layout(location = 1) in vec2 vertexTexCoord;
layout(location = 2) in vec3 vertexNormal;
layout(location = 6) in mat4 instanceTransform;
uniform mat4 mvp;
uniform float shellOffset;   // push out along the normal (object units)
uniform float shellFrac;     // 0..1 how far out this shell is
out vec3 fragNormal;
out vec3 fragPosition;
out vec2 fragUV;
out float vFrac;
void main() {
    mat3 m3 = mat3(instanceTransform);
    vec3 posExp = vertexPosition + vertexNormal*shellOffset;
    vec4 world = instanceTransform*vec4(posExp, 1.0);
    fragNormal = normalize(transpose(inverse(m3))*vertexNormal);
    fragPosition = world.xyz;
    vFrac = shellFrac;
    // Quantised seed: stable under tiny physics jitter so the fuzz doesn't swim.
    vec3 seed = floor(instanceTransform[3].xyz*4.0);
    float off = fract(sin(dot(seed, vec3(12.9, 78.2, 37.7)))*43758.5);
    fragUV = vertexTexCoord + vec2(off, off*1.7);
    gl_Position = mvp*world;
}
`

const haloFragmentShader = `
#version 330
in vec3 fragNormal;
in vec3 fragPosition;
in vec2 fragUV;
in float vFrac;
uniform vec4 colDiffuse;
uniform vec3 lightDir;
uniform float ambient;
uniform sampler2D texture2; // baked fibre noise (material normal-map slot)
out vec4 finalColor;

void main() {
    // A fibre reaches this shell only if its noise "length" clears the shell's
    // height; the threshold rises outward so tips thin to a sparse fuzz. A soft
    // edge (rather than a hard discard) antialiases the thin fibres so they
    // don't shimmer as the geometry moves sub-pixel.
    float n = texture(texture2, fragUV*vec2(4.0, 2.2)).r;
    float edge = 0.34 + vFrac*0.6;
    float cover = smoothstep(edge - 0.05, edge + 0.10, n);
    float a = cover*(1.0 - vFrac)*0.55;
    if (a < 0.004) discard;

    vec3 N = normalize(fragNormal);
    vec3 L = normalize(lightDir);
    float diff = clamp(dot(N, L)*0.5 + 0.5, 0.0, 1.0);
    vec3 col = colDiffuse.rgb*(ambient + (1.0 - ambient)*diff)*1.15;
    finalColor = vec4(col, a);
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
	noiseTex rl.Texture2D

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
		cylinder:   rl.GenMeshCylinder(1, 1, 8),
		sphere:     rl.GenMeshSphere(1, 6, 10),
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

	// Bake the fibre noise once into a small tiling texture, sampled by the
	// shaders instead of computed per fragment.
	img := rl.GenImagePerlinNoise(256, 256, 0, 0, 9.0)
	r.noiseTex = rl.LoadTextureFromImage(img)
	rl.UnloadImage(img)
	rl.SetTextureFilter(r.noiseTex, rl.FilterBilinear)
	rl.SetTextureWrap(r.noiseTex, rl.WrapRepeat)

	r.material = rl.LoadMaterialDefault()
	r.material.Shader = r.shader
	r.material.GetMap(rl.MapNormal).Texture = r.noiseTex

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
	r.haloMaterial.GetMap(rl.MapNormal).Texture = r.noiseTex
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
	r.sphBatches[k] = append(r.sphBatches[k], sphereTransform(pos, radius))
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
	// Only the cylinders get a halo — they form the yarn's silhouette. Skipping
	// the joint spheres roughly halves the fill the shells cost, for no visible
	// difference (the cylinder fuzz already covers the joints). Iterate the
	// colour batches in a stable order: Go map order is randomised each frame,
	// which with alpha blending would flicker at colour boundaries.
	keys := sortedBatchKeys(r.cylBatches)
	for _, s := range haloShells {
		rl.SetShaderValue(r.haloShader, r.hShellOffset, []float32{s.offset}, rl.ShaderUniformFloat)
		rl.SetShaderValue(r.haloShader, r.hShellFrac, []float32{s.frac}, rl.ShaderUniformFloat)
		for _, k := range keys {
			mats := r.cylBatches[k]
			if len(mats) > 0 {
				r.drawHaloBatch(r.cylinder, k, mats)
			}
		}
	}
	rl.EnableDepthMask()
	rl.EndBlendMode()
}

// sortedBatchKeys returns the batch keys in a deterministic order so alpha
// blending is stable frame-to-frame.
func sortedBatchKeys(m map[batchKey][]rl.Matrix) []batchKey {
	keys := make([]batchKey, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		ak := uint32(a.col.R)<<24 | uint32(a.col.G)<<16 | uint32(a.col.B)<<8 | uint32(a.col.A)
		bk := uint32(b.col.R)<<24 | uint32(b.col.G)<<16 | uint32(b.col.B)<<8 | uint32(b.col.A)
		if ak != bk {
			return ak < bk
		}
		if a.sheen != b.sheen {
			return a.sheen < b.sheen
		}
		return a.ambient < b.ambient
	})
	return keys
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
		rl.UnloadTexture(r.noiseTex)
		rl.UnloadShader(r.shader)
		rl.UnloadShader(r.haloShader)
	}
}

// cylinderTransform maps the unit cylinder (radius 1, base at the origin,
// extending up +Y by 1) so it spans a→b with the given radius. The matrix is
// built directly from the direction vector — its columns are the world images
// of the local X/Y/Z axes — which avoids the acos + rotation-matrix cost of
// composing MatrixRotate/MatrixMultiply for thousands of segments per frame.
func cylinderTransform(a, b rl.Vector3, radius float32) rl.Matrix {
	yx, yy, yz := b.X-a.X, b.Y-a.Y, b.Z-a.Z // local +Y → dir*length
	length := float32(math.Sqrt(float64(yx*yx + yy*yy + yz*yz)))
	if length < 1e-6 {
		length = 1e-6
	}
	inv := 1 / length
	ux, uy, uz := yx*inv, yy*inv, yz*inv // unit direction

	// Reference axis not parallel to the direction.
	var rx, ry, rz float32 = 0, 1, 0
	if uy > 0.99 || uy < -0.99 {
		rx, ry, rz = 1, 0, 0
	}
	// local +X → unit(ref × dir), scaled by radius.
	cx, cy, cz := ry*uz-rz*uy, rz*ux-rx*uz, rx*uy-ry*ux
	cl := float32(math.Sqrt(float64(cx*cx + cy*cy + cz*cz)))
	if cl < 1e-6 {
		cl = 1
	}
	xhx, xhy, xhz := cx/cl, cy/cl, cz/cl
	// local +Z → dir × X (already unit), scaled by radius.
	zhx, zhy, zhz := uy*xhz-uz*xhy, uz*xhx-ux*xhz, ux*xhy-uy*xhx

	return rl.Matrix{
		M0: xhx * radius, M4: yx, M8: zhx * radius, M12: a.X,
		M1: xhy * radius, M5: yy, M9: zhy * radius, M13: a.Y,
		M2: xhz * radius, M6: yz, M10: zhz * radius, M14: a.Z,
		M3: 0, M7: 0, M11: 0, M15: 1,
	}
}

// sphereTransform is a plain uniform scale + translation.
func sphereTransform(pos rl.Vector3, radius float32) rl.Matrix {
	return rl.Matrix{
		M0: radius, M4: 0, M8: 0, M12: pos.X,
		M1: 0, M5: radius, M9: 0, M13: pos.Y,
		M2: 0, M6: 0, M10: radius, M14: pos.Z,
		M3: 0, M7: 0, M11: 0, M15: 1,
	}
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
