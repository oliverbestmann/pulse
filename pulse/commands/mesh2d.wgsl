struct VertexOutput {
    @location(0) color: vec4f,
    @builtin(position) position: vec4f,
};

@vertex
fn vs_main(
    @location(0) position: vec2f,
    @location(1) color: vec4f,
) -> VertexOutput {
    // calculate position between 0 and 1, then convert to ndc
    let pos = position * vec2(2, -2) + vec2(-1, 1);

    var result: VertexOutput;
    result.position = vec4(pos.xy, 0.0, 1.0);
    result.color = color;
    return result;
}

@fragment
fn fs_main(vertex: VertexOutput) -> @location(0) vec4f {
    return vertex.color;
}
