package gpu

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Vertex represents the layout of data for a single point in 2D space.
type Vertex struct {
	Position [2]float32
	Color    [4]float32
	TexCoord [2]float32
}

// Mesh handles the VAO, VBO, and EBO for rendering geometry.
type Mesh struct {
	VAO        uint32
	VBO        uint32
	EBO        uint32
	IndexCount int32
}

// NewMesh creates a new Mesh object on the GPU.
func NewMesh(vertices []Vertex, indices []uint32) (*Mesh, error) {
	var vao, vbo, ebo uint32

	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*int(unsafe.Sizeof(Vertex{})), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.GenBuffers(1, &ebo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*4, gl.Ptr(indices), gl.STATIC_DRAW)

	// Position attribute (location 0)
	gl.VertexAttribPointerWithOffset(0, 2, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), 0)
	gl.EnableVertexAttribArray(0)

	// Color attribute (location 1)
	gl.VertexAttribPointerWithOffset(1, 4, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), 2*4)
	gl.EnableVertexAttribArray(1)

	// TexCoord attribute (location 2)
	gl.VertexAttribPointerWithOffset(2, 2, gl.FLOAT, false, int32(unsafe.Sizeof(Vertex{})), 6*4)
	gl.EnableVertexAttribArray(2)

	gl.BindVertexArray(0)

	return &Mesh{
		VAO:        vao,
		VBO:        vbo,
		EBO:        ebo,
		IndexCount: int32(len(indices)),
	}, nil
}

// NewRectMesh is a helper to create a rectangular mesh.
func NewRectMesh(x, y, w, h float32, r, g, b, a float32) (*Mesh, error) {
	vertices := []Vertex{
		// Pos          // Color        // UV
		{[2]float32{x, y}, [4]float32{r, g, b, a}, [2]float32{0, 0}},         // TL
		{[2]float32{x + w, y}, [4]float32{r, g, b, a}, [2]float32{1, 0}},     // TR
		{[2]float32{x + w, y + h}, [4]float32{r, g, b, a}, [2]float32{1, 1}}, // BR
		{[2]float32{x, y + h}, [4]float32{r, g, b, a}, [2]float32{0, 1}},     // BL
	}

	indices := []uint32{
		0, 1, 2,
		2, 3, 0,
	}

	return NewMesh(vertices, indices)
}

func (m *Mesh) Draw() {
	gl.BindVertexArray(m.VAO)
	gl.DrawElements(gl.TRIANGLES, m.IndexCount, gl.UNSIGNED_INT, nil)
	gl.BindVertexArray(0)
}

func (m *Mesh) Delete() {
	gl.DeleteVertexArrays(1, &m.VAO)
	gl.DeleteBuffers(1, &m.VBO)
	gl.DeleteBuffers(1, &m.EBO)
}

// Scissor helpers for clipping
func SetClip(x, y, width, height int32, windowHeight int32) {
	gl.Enable(gl.SCISSOR_TEST)
	// OpenGL scissor coordinates start from bottom-left.
	// We convert from top-left.
	gl.Scissor(x, windowHeight-(y+height), width, height)
}

func ClearClip() {
	gl.Disable(gl.SCISSOR_TEST)
}
