package engine

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"

	"github.com/SoyStudios/moonshot/crochet/yarn"
)

// GPU lighting. Instead of shading each yarn segment on the CPU, the renderer
// draws unit cylinder/sphere meshes (which carry vertex normals) through a GLSL
// material. raylib feeds the shader the model, normal and MVP matrices plus the
// per-draw diffuse colour automatically; we supply the light and the material's
// sheen/ambient as uniforms. Lighting — diffuse + a specular highlight — is
// then computed per fragment on the GPU, so tubes look round and glossy yarn
// reads differently from matte wool.
//
// If the shader fails to compile (unusual, but possible on a limited driver)
// the renderer falls back to raylib's flat immediate-mode cylinders so the demo
// still runs.

const yarnVertexShader = `
#version 330
in vec3 vertexPosition;
in vec3 vertexNormal;
in vec4 vertexColor;
uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 matNormal;
out vec3 fragPosition;
out vec3 fragNormal;
out vec4 fragColor;
void main() {
    fragPosition = vec3(matModel*vec4(vertexPosition, 1.0));
    fragNormal   = normalize(vec3(matNormal*vec4(vertexNormal, 1.0)));
    fragColor    = vertexColor;
    gl_Position  = mvp*vec4(vertexPosition, 1.0);
}
`

const yarnFragmentShader = `
#version 330
in vec3 fragPosition;
in vec3 fragNormal;
in vec4 fragColor;
uniform vec4 colDiffuse;   // per-draw yarn colour (raylib sets from material)
uniform vec3 lightDir;     // direction toward the key light (normalized)
uniform vec3 lightColor;
uniform float ambient;     // base brightness in shadow
uniform float sheen;       // specular strength (matte 0 .. glossy 1)
uniform vec3 viewPos;
out vec4 finalColor;
void main() {
    vec3 N = normalize(fragNormal);
    vec3 L = normalize(lightDir);
    float diff = max(dot(N, L), 0.0);

    vec3 V = normalize(viewPos - fragPosition);
    vec3 H = normalize(L + V);
    float spec = pow(max(dot(N, H), 0.0), 24.0)*sheen;

    vec3 base = colDiffuse.rgb*fragColor.rgb;
    vec3 lit  = base*(ambient + (1.0 - ambient)*diff) + spec*lightColor;
    finalColor = vec4(lit, colDiffuse.a*fragColor.a);
}
`

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
}

// newYarnRenderer builds the GPU resources. Must be called after InitWindow
// (a GL context has to exist).
func newYarnRenderer() *yarnRenderer {
	r := &yarnRenderer{
		cylinder: rl.GenMeshCylinder(1, 1, 10),
		sphere:   rl.GenMeshSphere(1, 8, 12),
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

	r.material = rl.LoadMaterialDefault()
	r.material.Shader = r.shader

	// The key light is fixed in world space (from the upper-left front).
	toLight := rl.Vector3Normalize(rl.NewVector3(0.45, 1.0, 0.35))
	rl.SetShaderValue(r.shader, r.locLightDir, []float32{toLight.X, toLight.Y, toLight.Z}, rl.ShaderUniformVec3)
	rl.SetShaderValue(r.shader, r.locLightColor, []float32{1.0, 0.98, 0.94}, rl.ShaderUniformVec3)
	return r
}

// beginFrame updates the per-frame view position uniform.
func (r *yarnRenderer) beginFrame(camPos rl.Vector3) {
	if !r.lit {
		return
	}
	rl.SetShaderValue(r.shader, r.locViewPos, []float32{camPos.X, camPos.Y, camPos.Z}, rl.ShaderUniformVec3)
}

// setMaterial selects the sheen/ambient for the strand about to be drawn.
func (r *yarnRenderer) setMaterial(m yarn.Material) {
	if !r.lit {
		return
	}
	rl.SetShaderValue(r.shader, r.locAmbient, []float32{float32(m.Ambient)}, rl.ShaderUniformFloat)
	rl.SetShaderValue(r.shader, r.locSheen, []float32{float32(m.Sheen)}, rl.ShaderUniformFloat)
}

// segment draws a yarn section from a to b as a lit cylinder of the given radius.
func (r *yarnRenderer) segment(a, b rl.Vector3, radius float32, col yarn.Color) {
	if !r.lit {
		rl.DrawCylinderEx(a, b, radius, radius, 8, toRL(col))
		return
	}
	tf := cylinderTransform(a, b, radius)
	r.material.GetMap(rl.MapDiffuse).Color = toRL(col)
	rl.DrawMesh(r.cylinder, r.material, tf)
}

// node draws a rounded joint sphere.
func (r *yarnRenderer) node(pos rl.Vector3, radius float32, col yarn.Color) {
	if !r.lit {
		rl.DrawSphere(pos, radius, toRL(col))
		return
	}
	tf := rl.MatrixMultiply(
		rl.MatrixScale(radius, radius, radius),
		rl.MatrixTranslate(pos.X, pos.Y, pos.Z),
	)
	r.material.GetMap(rl.MapDiffuse).Color = toRL(col)
	rl.DrawMesh(r.sphere, r.material, tf)
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

// alignYTo returns a rotation that turns +Y onto the (already length-known)
// direction vector.
func alignYTo(dir rl.Vector3, length float32) rl.Matrix {
	d := rl.NewVector3(dir.X/length, dir.Y/length, dir.Z/length)
	axis := rl.Vector3CrossProduct(rl.NewVector3(0, 1, 0), d)
	axisLen := float32(math.Sqrt(float64(axis.X*axis.X + axis.Y*axis.Y + axis.Z*axis.Z)))
	if axisLen < 1e-6 {
		// Parallel to +Y (same or opposite direction).
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
