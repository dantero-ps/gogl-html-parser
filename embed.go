package goglweb

import _ "embed"

//go:embed assets/shaders/vertex.glsl
var defaultVertexShader string

//go:embed assets/shaders/fragment.glsl
var defaultFragmentShader string
