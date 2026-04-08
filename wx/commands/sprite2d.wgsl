struct VertexOutput {
    @location(0) uv: vec2f,
    @location(1) color: vec4f,
    @location(2) @interpolate(flat) scissor_min: vec2f,
    @location(3) @interpolate(flat) scissor_max: vec2f,
    @builtin(position) position: vec4f,
};

struct Region {
    pos: vec2f,
    size: vec2f,
};

struct Regions {
    rects: array<Region, 4>,
    transforms: array<mat3x3f, 8>,
    twod: array<array<vec3f, 2>, 4>,
}

fn decode_region(in: vec2<u32>) -> Region {
    let x = f32((in[0] ) & 0xffff);
    let y = f32((in[0] >> 16) & 0xffff);
    let width = f32((in[1] >> 0) & 0xffff);
    let height = f32((in[1] >> 16) & 0xffff);

    var rect: Region;
    rect.pos = vec2f(x, y);
    rect.size = vec2f(width, height);
    return rect;
}

struct VertexUniforms {
    // transforms the resulting coordinates into the 0 to 1 scale
    target_texture_size: vec2f,

    // the source texture size
    source_texture_size: vec2f,
}

@group(0)
@binding(2)
var<uniform> vertex_uniforms: VertexUniforms;

@vertex
fn vs_main(
    @builtin(vertex_index) index: u32,
    @location(0) color: vec4f,
    @location(1) tr0: vec3f,
    @location(2) tr1: vec3f,
    @location(3) source_region_in: vec2<u32>,
    @location(4) target_region_in: vec2<u32>,
) -> VertexOutput {
    // decode (sub) image source and target region
    let source_region = decode_region(source_region_in);
    let target_region = decode_region(target_region_in);

    // apply sprite transform
    let model_transform = transpose(
        mat3x3(tr0, tr1, vec3f(0, 0, 1)),
    );

    // need to scale with the source region
    let view_transform = mat3x3(
        vec3f(source_region.size.x, 0, 0),
        vec3f(0, source_region.size.y, 0),
        vec3f(0, 0, 1),
    );

    // between 0 and 1
    // index vertices as p00, p01, p10, p11, this way
    // x and y can be derived from the lower bit of index
    let x = f32((index >> 1) & 1);
    let y = f32(index & 1);
    let vertex = vec2f(x, y);

    // target position relative to target region in pixel space
    let pos_px = (model_transform * view_transform * vec3f(vertex, 1)).xy;

    // calculate vertex position in ndc space (-1 to +1)
    let pos_zo = ((target_region.pos + pos_px) / vertex_uniforms.target_texture_size);
    let pos = pos_zo * 2.0 - 1.0;

    // convert the source region to to 0 to 1 uv values
    let texsize = 1.0 / vertex_uniforms.source_texture_size;
    let uv_offset = source_region.pos * texsize;
    let uv_scale = source_region.size * texsize * vertex;

    var result: VertexOutput;
    result.uv = uv_offset + uv_scale;
    result.color = color;
    result.position = vec4(pos.x, -pos.y, 0.0, 1.0);

    // configure our own scissor rect to discard fragment samples
    // if they do not fall into the target region
    result.scissor_min = target_region.pos;
    result.scissor_max = target_region.pos + target_region.size;

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
    if vertex.position.x < vertex.scissor_min.x ||
       vertex.position.x > vertex.scissor_max.x ||
       vertex.position.y < vertex.scissor_min.y ||
       vertex.position.y > vertex.scissor_max.y {

       discard;
    }

    let tex = textureSample(texture, texSampler, vertex.uv);
    return tex * vertex.color;
}
