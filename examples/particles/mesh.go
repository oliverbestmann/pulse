package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/oliverbestmann/go3d/glm"
)

type Vec = glm.Vec3[float32]

type Triangle struct {
	A, B, C    Vec
	Na, Nb, Nc Vec
}

func calculateNormal(a, b, c Vec) Vec {
	u := a.Sub(b)
	v := c.Sub(b)
	n := u.Cross(v)
	return n.Normalize()
}

type Mesh struct {
	Name      string
	Triangles []Triangle
}

func LoadMeshes(obj string) ([]Mesh, error) {
	var vertices []Vec
	var normals []Vec

	var mesh Mesh
	var meshes []Mesh

	finalize := func() {
		if len(mesh.Triangles) == 0 {
			return
		}
		meshes = append(meshes, mesh)
	}

	for line := range strings.Lines(obj) {
		if strings.HasPrefix(line, "o ") {
			finalize()
			mesh = Mesh{}
		}

		if strings.HasPrefix(line, "v ") {
			fields := strings.Fields(line[2:])
			if len(fields) != 3 {
				return nil, errors.New("expected three coordinates in vertex")
			}

			x, errX := strconv.ParseFloat(fields[0], 64)
			y, errY := strconv.ParseFloat(fields[1], 64)
			z, errZ := strconv.ParseFloat(fields[2], 64)

			if errX != nil || errY != nil || errZ != nil {
				return nil, errors.Join(errX, errY, errZ)
			}

			vertices = append(vertices, Vec{float32(x), float32(y), float32(z)})
			continue
		}

		if strings.HasPrefix(line, "vn ") {
			fields := strings.Fields(line[3:])
			if len(fields) != 3 {
				return nil, errors.New("expected three coordinates in normal")
			}

			x, errX := strconv.ParseFloat(fields[0], 64)
			y, errY := strconv.ParseFloat(fields[1], 64)
			z, errZ := strconv.ParseFloat(fields[2], 64)

			if errX != nil || errY != nil || errZ != nil {
				return nil, errors.Join(errX, errY, errZ)
			}

			n := Vec{float32(x), float32(y), float32(z)}
			n.Normalize()
			normals = append(normals, n)

			continue
		}

		if strings.HasPrefix(line, "f ") {
			fields := strings.Fields(line[2:])
			if len(fields) != 3 {
				return nil, errors.New("expected three coordinates in vertex")
			}

			a, errA := parseVertexIndex(fields[0])
			b, errB := parseVertexIndex(fields[1])
			c, errC := parseVertexIndex(fields[2])

			if errA != nil || errB != nil || errC != nil {
				return nil, errors.Join(errA, errB, errC)
			}

			tri := Triangle{
				A: vertices[a.Vertex-1],
				B: vertices[b.Vertex-1],
				C: vertices[c.Vertex-1],
			}

			if a.Normal > 0 && b.Normal > 0 && c.Normal > 0 {
				tri.Na = normals[a.Normal-1]
				tri.Nb = normals[b.Normal-1]
				tri.Nc = normals[c.Normal-1]
			} else {
				tri.Na = calculateNormal(tri.A, tri.B, tri.C)
				tri.Nb = tri.Na
				tri.Nc = tri.Na
			}

			mesh.Triangles = append(mesh.Triangles, tri)

			continue
		}
	}

	finalize()

	return meshes, nil
}

type vertexIndex struct {
	Vertex int
	Normal int
}

func parseVertexIndex(input string) (vertexIndex, error) {
	parts := strings.Split(input, "/")
	if len(parts) < 1 {
		return vertexIndex{}, fmt.Errorf("invalid vertex index: %q", input)
	}

	var err error
	var res vertexIndex

	res.Vertex, err = strconv.Atoi(parts[0])
	if err != nil {
		return vertexIndex{}, fmt.Errorf("parse vertex index %q: %w", parts[0], err)
	}

	if len(parts) >= 2 {
		res.Normal, err = strconv.Atoi(parts[2])
		if err != nil {
			return vertexIndex{}, fmt.Errorf("parse normal index %q: %w", parts[2], err)
		}
	}

	return res, nil
}
