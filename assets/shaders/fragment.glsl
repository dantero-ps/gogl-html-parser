#version 410 core
in vec4 vColor;
in vec2 vTexCoord;

uniform sampler2D uTexture;
uniform bool uUseTexture;

out vec4 FragColor;

void main() {
    if (uUseTexture) {
        FragColor = texture(uTexture, vTexCoord) * vColor;
    } else {
        FragColor = vColor;
    }
}
