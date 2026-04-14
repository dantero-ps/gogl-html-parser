package gpu

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Shader represents an OpenGL shader program.
type Shader struct {
	Program uint32
}

// NewShader creates, compiles, and links a shader program from file paths.
func NewShader(vertexPath, fragmentPath string) (*Shader, error) {
	vSource, err := os.ReadFile(vertexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read vertex shader: %w", err)
	}

	fSource, err := os.ReadFile(fragmentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fragment shader: %w", err)
	}

	vShader, err := compileShader(string(vSource)+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	defer gl.DeleteShader(vShader)

	fShader, err := compileShader(string(fSource)+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}
	defer gl.DeleteShader(fShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vShader)
	gl.AttachShader(program, fShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))
		return nil, fmt.Errorf("failed to link program: %v", log)
	}

	return &Shader{Program: program}, nil
}

// NewShaderFromSource creates, compiles, and links a shader program from source strings.
func NewShaderFromSource(vertexSrc, fragmentSrc string) (*Shader, error) {
	vShader, err := compileShader(vertexSrc+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}
	defer gl.DeleteShader(vShader)

	fShader, err := compileShader(fragmentSrc+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
	}
	defer gl.DeleteShader(fShader)

	program := gl.CreateProgram()
	gl.AttachShader(program, vShader)
	gl.AttachShader(program, fShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)
		logMsg := strings.Repeat("\x00", int(logLength))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(logMsg))
		return nil, fmt.Errorf("failed to link program: %v", logMsg)
	}

	return &Shader{Program: program}, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)
		log := strings.Repeat("\x00", int(logLength))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))
		return 0, fmt.Errorf("failed to compile shader (type %d): %v", shaderType, log)
	}

	return shader, nil
}

func (s *Shader) Use() {
	gl.UseProgram(s.Program)
}

func (s *Shader) SetFloat(name string, value float32) {
	gl.Uniform1f(gl.GetUniformLocation(s.Program, gl.Str(name+"\x00")), value)
}

func (s *Shader) SetVec2(name string, x, y float32) {
	gl.Uniform2f(gl.GetUniformLocation(s.Program, gl.Str(name+"\x00")), x, y)
}

func (s *Shader) SetVec4(name string, x, y, z, w float32) {
	gl.Uniform4f(gl.GetUniformLocation(s.Program, gl.Str(name+"\x00")), x, y, z, w)
}

func (s *Shader) SetMat4(name string, value *float32) {
	gl.UniformMatrix4fv(gl.GetUniformLocation(s.Program, gl.Str(name+"\x00")), 1, false, value)
}

func (s *Shader) SetInt(name string, value int32) {
	gl.Uniform1i(gl.GetUniformLocation(s.Program, gl.Str(name+"\x00")), value)
}

func (s *Shader) SetBool(name string, value bool) {
	val := int32(0)
	if value {
		val = 1
	}
	gl.Uniform1i(gl.GetUniformLocation(s.Program, gl.Str(name+"\x00")), val)
}

func (s *Shader) Delete() {
	gl.DeleteProgram(s.Program)
}
