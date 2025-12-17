@group(0) @binding(0) var<storage, read> buf: array<u32>;

const width: u32 = 1024 * 2;
const height: u32 = 1024 * 2;

struct VertexOut {
    @builtin(position) clip: vec4f,
    @location(0) uv: vec2f,
}

@vertex
fn vertex(@builtin(vertex_index) index: u32) -> VertexOut {
    var out: VertexOut;

    switch (index) {
        case 0: {
            out.clip = vec4f(-1, 3, 0, 1);
            out.uv = vec2f(0, 2);

        }
        case 1: {
            out.clip = vec4f(-1, -1, 0, 1);
            out.uv = vec2f(0, 0);
        }

        case 2, default: {
            out.clip = vec4f(3, -1, 0, 1);
            out.uv = vec2f(2, 0);
        }
    }

    return out;
}

@fragment
fn fragment(in: VertexOut) -> @location(0) vec4f {
    let x = min(width - 1, u32(in.uv.x * f32(width) + 0.5));
    let y = min(height - 1, u32(in.uv.y * f32(height) + 0.5));

    let value = f32(buf[y * width + x]) / 256;
    return vec4f(value, value, value, 1);
}
