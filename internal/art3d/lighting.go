package art3d

import rl "github.com/gen2brain/raylib-go/raylib"

func LoadLighting() rl.Shader {
	return rl.LoadShaderFromMemory(`#version 330
in vec3 vertexPosition;
in vec2 vertexTexCoord;
in vec3 vertexNormal;
in vec4 vertexColor;
uniform mat4 mvp;
uniform mat4 matModel;
uniform mat4 matNormal;
out vec2 fragTexCoord;
out vec4 fragColor;
out vec3 normal;
void main() {
 fragTexCoord = vertexTexCoord;
 fragColor = vertexColor;
 normal = normalize(vec3(matNormal * vec4(vertexNormal, 0.0)));
 gl_Position = mvp * vec4(vertexPosition, 1.0);
}`, `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
in vec3 normal;
uniform sampler2D texture0;
uniform vec4 colDiffuse;
out vec4 finalColor;
void main() {
 vec4 base = texture(texture0, fragTexCoord) * colDiffuse * fragColor;
 float key = max(dot(normalize(normal), normalize(vec3(-0.4, 0.8, 0.6))), 0.0);
 float rim = pow(max(dot(normalize(normal), normalize(vec3(0.2, 0.3, -0.9))), 0.0), 2.0);
 vec3 light = vec3(0.57, 0.61, 0.68) + key * vec3(0.66, 0.61, 0.53) + rim * vec3(0.15, 0.42, 0.56);
 finalColor = vec4(base.rgb * light, base.a);
}`)
}
