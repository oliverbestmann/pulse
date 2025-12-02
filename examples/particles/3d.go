package main

/*
	type Vertex3 struct {
		pos    [4]float32
		normal [3]float32
	}

	var VertexBufferLayout = wgpu.VertexBufferLayout{
		ArrayStride: uint64(unsafe.Sizeof(Vertex3{})),
		StepMode:    wgpu.VertexStepModeVertex,
		Attributes: []wgpu.VertexAttribute{
			{
				// vertex position
				Format:         wgpu.VertexFormatFloat32x4,
				Offset:         0,
				ShaderLocation: 0,
			},
			{
				// vertex normal
				Format:         wgpu.VertexFormatFloat32x3,
				Offset:         4 * 4,
				ShaderLocation: 1,
			},
		},
	}

	func prepareVertices(meshes []Mesh) []Vertex3 {
		var instances []Vertex3
		for _, mesh := range meshes {
			for _, tri := range mesh.Triangles {
				instances = append(instances,
					Vertex3{
						pos:    tri.A.Extend(1),
						normal: tri.Na,
					},
					Vertex3{
						pos:    tri.B.Extend(1),
						normal: tri.Nb,
					},
					Vertex3{
						pos:    tri.C.Extend(1),
						normal: tri.Nc,
					},
				)
			}
		}

		return instances
	}

	func generateViewProjectionMatrix(aspectRatio float32, time time.Duration) glm.Mat4[float32] {
		sin := float32(math.Sin(time.Seconds()))
		cos := float32(math.Cos(time.Seconds()))

		projection := glm.Perspective(math.Pi/4, aspectRatio, 0.1, 100)
		view := glm.LookAt(
			glm.Vec3[float32]{sin * 10, cos * 10, 5},
			glm.Vec3[float32]{0, 0, 1},
			glm.Vec3[float32]{0, 0, 1},
		)

		return projection.Mul(view)
	}

//go:embed shader.wgsl
var shader string

//go:embed shader2d.wgsl
var shader2d string

	type Context struct {
		vertexBuf  *wgpu.Buffer
		uniformBuf *wgpu.Buffer
		sprites   *wgpu.RenderPipeline

		gopher *Texture

		startTime   time.Time
		vertexCount uint32
		pipeline2d  *wgpu.RenderPipeline
	}

	func InitState(s *ViewState) (s *Context, err error) {
		defer func() {
			if err != nil {
				s.Destroy()
				s = nil
			}
		}()

		s = &Context{}
		s.startTime = time.Now()

		meshes, err := LoadMeshes(_scene_obj)
		if err != nil {
			return s, fmt.Errorf("load meshes: %w", err)
		}

		instances := prepareVertices(meshes)
		s.vertexCount = uint32(len(instances))

		s.vertexBuf, err = s.device.TryCreateBufferInit(&wgpu.BufferInitDescriptor{
			Label:    "Vertex3 Buffer",
			Contents: wgpu.ToBytes(instances),
			Usage:    wgpu.BufferUsageVertex,
		})
		if err != nil {
			return s, err
		}

		mxTotal := generateViewProjectionMatrix(float32(s.config.Width)/float32(s.config.Height), time.Since(s.startTime))
		s.uniformBuf, err = s.device.TryCreateBufferInit(&wgpu.BufferInitDescriptor{
			Label:    "Uniform Buffer",
			Contents: wgpu.ToBytes(mxTotal[:]),
			Usage:    wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
		})
		if err != nil {
			return s, err
		}

		shader, err := s.device.TryCreateShaderModule(&wgpu.ShaderModuleDescriptor{
			Label:          "shader.wgsl",
			WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{Code: shader},
		})
		if err != nil {
			return s, err
		}
		defer shader.Release()

		s.sprites, err = s.device.TryCreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
			Label: "3D",
			Vertex: wgpu.VertexState{
				Module:     shader,
				EntryPoint: "vs_main",
				Buffers:    []wgpu.VertexBufferLayout{VertexBufferLayout},
			},
			Fragment: &wgpu.FragmentState{
				Module:     shader,
				EntryPoint: "fs_main",
				Targets: []wgpu.ColorTargetState{
					{
						Format:    s.config.Format,
						Blend:     nil,
						WriteMask: wgpu.ColorWriteMaskAll,
					},
				},
			},
			Primitive: wgpu.PrimitiveState{
				Topology:  wgpu.PrimitiveTopologyTriangleList,
				FrontFace: wgpu.FrontFaceCCW,
				CullMode:  wgpu.CullModeNone,
			},
			DepthStencil: &wgpu.DepthStencilState{
				Format:            wgpu.TextureFormatDepth32Float,
				DepthWriteEnabled: true,
				DepthCompare:      wgpu.CompareFunctionLess,
				StencilFront: wgpu.StencilFaceState{
					Compare:     wgpu.CompareFunctionAlways,
					FailOp:      wgpu.StencilOperationKeep,
					DepthFailOp: wgpu.StencilOperationKeep,
					PassOp:      wgpu.StencilOperationKeep,
				},
				StencilBack: wgpu.StencilFaceState{
					Compare:     wgpu.CompareFunctionAlways,
					FailOp:      wgpu.StencilOperationKeep,
					DepthFailOp: wgpu.StencilOperationKeep,
					PassOp:      wgpu.StencilOperationKeep,
				},
			},
			Multisample: wgpu.MultisampleState{
				Count:                  4,
				Mask:                   0xFFFFFFFF,
				AlphaToCoverageEnabled: false,
			},
		})
		if err != nil {
			return s, err
		}

		s.pipeline2d = createDrawSpritePipeline(s)

		src, _, _ := image.Decode(bytes.NewReader(gopherImage))
		s.gopher = ImportTexture(s, src)

		return s, nil
	}

	func createDrawSpritePipeline(s *Context) *wgpu.RenderPipeline {
		shader, err := s.device.TryCreateShaderModule(&wgpu.ShaderModuleDescriptor{
			Label:          "shader2d.wgsl",
			WGSLDescriptor: &wgpu.ShaderModuleWGSLDescriptor{Code: shader2d},
		})
		if err != nil {
			must(err)
		}

		defer shader.Release()

		desc := wgpu.RenderPipelineDescriptor{
			Label: "2D",
			Vertex: wgpu.VertexState{
				Module:     shader,
				EntryPoint: "vs_main",
				Buffers: []wgpu.VertexBufferLayout{
					{
						ArrayStride: uint64(unsafe.Sizeof(SpriteInstance{})),
						StepMode:    wgpu.VertexStepModeVertex,
						Attributes: []wgpu.VertexAttribute{
							{
								// position
								Format:         wgpu.VertexFormatFloat32x2,
								Offset:         0,
								ShaderLocation: 0,
							},
							{
								// uv
								Format:         wgpu.VertexFormatFloat32x2,
								Offset:         uint64(unsafe.Offsetof(SpriteInstance{}.U)),
								ShaderLocation: 1,
							},
						},
					},
				},
			},
			Fragment: &wgpu.FragmentState{
				Module:     shader,
				EntryPoint: "fs_main",
				Targets: []wgpu.ColorTargetState{
					{
						Format:    s.config.Format,
						Blend:     nil,
						WriteMask: wgpu.ColorWriteMaskAll,
					},
				},
			},
			Primitive: wgpu.PrimitiveState{
				Topology:  wgpu.PrimitiveTopologyTriangleList,
				FrontFace: wgpu.FrontFaceCCW,
				CullMode:  wgpu.CullModeNone,
			},
			DepthStencil: nil,
			Multisample: wgpu.MultisampleState{
				Count:                  4,
				Mask:                   0xFFFFFFFF,
				AlphaToCoverageEnabled: false,
			},
		}

		sprites, err := s.device.TryCreateRenderPipeline(&desc)
		if err != nil {
			must(err)
		}

		return sprites
	}

	func (s *Context) Resize(width, height int) {
		if width > 0 && height > 0 {
			s.config.Width = uint32(width)
			s.config.Height = uint32(height)

			s.depthView.Release()
			s.depthTexture.Release()
			s.depthTexture, s.depthView, _ = createDepthTexture(s)

			s.multisampleTextureView.Release()
			s.multisampleTexture.Release()
			s.multisampleTexture, s.multisampleTextureView, _ = createMultisampleTexture(s.device, s.config)

			mxTotal := generateViewProjectionMatrix(float32(width)/float32(height), time.Since(s.startTime))
			s.queue.TryWriteBuffer(s.uniformBuf, 0, wgpu.ToBytes(mxTotal[:]))

			s.surface.Configure(s.adapter, s.device, s.config)
		}
	}

	func DrawTriangles(s *Context, target *wgpu.TextureView, image *Texture) {
		encoder, err := s.device.TryCreateCommandEncoder(nil)
		if err != nil {
			must(err)
		}
		defer encoder.Release()

		renderPass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
			ColorAttachments: []wgpu.RenderPassColorAttachment{
				{
					View:          s.multisampleTextureView,
					ResolveTarget: target,
					LoadOp:        wgpu.LoadOpLoad,
					StoreOp:       wgpu.StoreOpStore,
				},
			},
		})

		bindGroupLayout := s.pipeline2d.GetBindGroupLayout(0)
		defer bindGroupLayout.Release()

		bindGroup := must2(s.device.TryCreateBindGroup(&wgpu.BindGroupDescriptor{
			Layout: bindGroupLayout,
			Entries: []wgpu.BindGroupEntry{
				{
					Binding:     0,
					TextureView: s.gopher.view,
					Size:        wgpu.WholeSize,
				},
				{
					Binding: 1,
					Sampler: s.gopher.sampler,
				},
			},
		}))
		defer bindGroup.Release()

		var x, y, w, h float32

		x = 16
		y = 16
		w = 128
		h = 128

		vertexBuf := []SpriteInstance{
			{x, y, 0, 0},
			{x + w, y, 1, 0},
			{x, y + h, 0, 1},
			{x + w, y + h, 1, 1},
		}

		indexBuf := []uint32{
			0, 2, 3,
			0, 3, 1,
		}

		bufInstances := must2(s.device.TryCreateBufferInit(&wgpu.BufferInitDescriptor{
			Label:    "QuadVertices",
			Contents: wgpu.ToBytes(vertexBuf),
			Usage:    wgpu.BufferUsageVertex,
		}))

		bufIndices := must2(s.device.TryCreateBufferInit(&wgpu.BufferInitDescriptor{
			Label:    "QuadIndices",
			Contents: wgpu.ToBytes(indexBuf),
			Usage:    wgpu.BufferUsageIndex,
		}))

		renderPass.SetPipeline(s.pipeline2d)
		renderPass.SetBindGroup(0, bindGroup, nil)
		renderPass.SetVertexBuffer(0, bufInstances, 0, wgpu.WholeSize)
		renderPass.SetIndexBuffer(bufIndices, wgpu.IndexFormatUint32, 0, bufIndices.GetSize())
		renderPass.DrawIndexed(6, 1, 0, 0, 0)
		if err := renderPass.TryEnd(); err != nil {
			renderPass.Release()
			must(err)
		}
		renderPass.Release() // must release

		cmdBuffer, err := encoder.TryFinish(nil)
		if err != nil {
			must(err)
		}
		defer cmdBuffer.Release()

		s.queue.Submit(cmdBuffer)
	}

	func must2[T any](value T, err error) T {
		must(err)
		return value
	}

	type SpriteInstance struct {
		X float32
		Y float32
		U float32
		V float32
	}

	func (s *Context) DrawTriangles(screen *wgpu.Texture) error {
		// update the camera matrix
		width, height := s.config.Width, s.config.Height
		mxTotal := generateViewProjectionMatrix(float32(width)/float32(height), time.Since(s.startTime))
		s.queue.TryWriteBuffer(s.uniformBuf, 0, wgpu.ToBytes(mxTotal[:]))

		view, err := screen.TryCreateView(nil)
		if err != nil {
			return err
		}
		defer view.Release()

		encoder, err := s.device.TryCreateCommandEncoder(nil)
		if err != nil {
			return err
		}
		defer encoder.Release()

		renderPass := encoder.BeginRenderPass(&wgpu.RenderPassDescriptor{
			ColorAttachments: []wgpu.RenderPassColorAttachment{
				{
					View:          s.multisampleTextureView,
					ResolveTarget: view,
					LoadOp:        wgpu.LoadOpClear,
					StoreOp:       wgpu.StoreOpStore,
					ClearValue:    wgpu.Color{R: 0.1, G: 0.2, B: 0.3, A: 1.0},
				},
			},

			DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
				View:            s.depthView,
				DepthLoadOp:     wgpu.LoadOpClear,
				DepthStoreOp:    wgpu.StoreOpStore,
				DepthClearValue: 1.0,
			},
		})

		bindGroupLayout := s.sprites.GetBindGroupLayout(0)
		defer bindGroupLayout.Release()

		bindGroup, err := s.device.TryCreateBindGroup(&wgpu.BindGroupDescriptor{
			Layout: bindGroupLayout,
			Entries: []wgpu.BindGroupEntry{
				{
					Binding: 0,
					Buffer:  s.uniformBuf,
					Size:    wgpu.WholeSize,
				},
				{
					Binding:     1,
					TextureView: s.gopher.view,
					Size:        wgpu.WholeSize,
				},
				{
					Binding: 2,
					Sampler: s.gopher.sampler,
				},
			},
		})
		if err != nil {
			return err
		}

		renderPass.SetPipeline(s.sprites)
		renderPass.SetBindGroup(0, bindGroup, nil)
		renderPass.SetVertexBuffer(0, s.vertexBuf, 0, wgpu.WholeSize)
		renderPass.Draw(s.vertexCount, 1, 0, 0)
		if err := renderPass.TryEnd(); err != nil {
			renderPass.Release()
			return err
		}
		renderPass.Release() // must release

		cmdBuffer, err := encoder.TryFinish(nil)
		if err != nil {
			return err
		}
		defer cmdBuffer.Release()

		s.queue.Submit(cmdBuffer)

		DrawTriangles(s, view, s.gopher)

		s.surface.Present()

		return nil
	}

	func (s *Context) Destroy() {
		if s.sprites != nil {
			s.sprites.Release()
			s.sprites = nil
		}
		if s.uniformBuf != nil {
			s.uniformBuf.Release()
			s.uniformBuf = nil
		}
		if s.vertexBuf != nil {
			s.vertexBuf.Release()
			s.vertexBuf = nil
		}
		if s.config != nil {
			s.config = nil
		}
		if s.queue != nil {
			s.queue.Release()
			s.queue = nil
		}
		if s.device != nil {
			s.device.Release()
			s.device = nil
		}
		if s.adapter != nil {
			s.adapter.Release()
			s.adapter = nil
		}
		if s.surface != nil {
			s.surface.Release()
			s.surface = nil
		}
	}
*/
