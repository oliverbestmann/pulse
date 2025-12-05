struct VertexOutput {
    @location(0) color: vec4f,
    @builtin(position) position: vec4f,
};

@group(0) @binding(0) var<uniform> view_transform: mat3x3<f32>;

@group(0) @binding(1) var<storage, read> model_transforms: array<mat3x3<f32>>;

@vertex
fn vs_main(
    @location(0) position: vec2f,
    @location(1) color: vec4f,
    @location(2) transform_idx: u32,
) -> VertexOutput {
    let model_transform = model_transforms[transform_idx];

    // calculate position between 0 and 1, then convert to ndc
    let pos = view_transform * model_transform * vec3f(position, 1);

    var result: VertexOutput;
    result.position = vec4(pos.xy, 0.0, 1.0);
    result.color = color;
    return result;
}

@fragment
fn fs_main(vertex: VertexOutput) -> @location(0) vec4f {
    return vertex.color;
}
