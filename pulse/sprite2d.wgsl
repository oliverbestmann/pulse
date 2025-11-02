struct VertexOutput {
    @location(0) uv: vec2f,
    @location(1) color: vec4f,
    @builtin(position) position: vec4f,
};

@group(0)
@binding(2)
var<uniform> view_transform: mat3x3f;

@vertex
fn vs_main(
    @builtin(vertex_index) index: u32,
    @location(0) color: vec4f,
    @location(1) uv_offset: vec2f,
    @location(2) uv_scale: vec2f,
    @location(3) tr0: vec3f,
    @location(4) tr1: vec3f,
) -> VertexOutput {
    let z = vec3f(0, 0, 1);
    let transform = view_transform * transpose(mat3x3(tr0, tr1, z));

    // index vertices as p00, p01, p10, p11, this way
    // x and y can be derived from the lower bit of index
    let x = f32((index >> 1) & 1);
    let y = f32(index & 1);

    let pos_zo = (transform * vec3f(x, y, 1)).xy;
    let pos = pos_zo * 2 - 1;

    var result: VertexOutput;
    result.position = vec4(pos.x, -pos.y, 0.0, 1.0);
    result.uv = uv_offset + uv_scale * vec2f(x, y);
    result.color = color;
    return result;
}

@group(0)
@binding(0)
var texture: texture_2d<f32>;

@group(0)
@binding(1)
var texSampler: sampler;

@fragment
fn fs_main(vertex: VertexOutput) -> @location(0) vec4f {
    let tex = textureSample(texture, texSampler, vertex.uv);
   return tex * vertex.color;
}
