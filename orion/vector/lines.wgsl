
@group(0) @binding(0)
var<uniform> config: LineConfig;

@group(0) @binding(1)
var<storage, read> points: array<vec2f>;

struct LineConfig {
    projection: mat3x3<f32>,
    color: vec4f,
    thickness: f32,

    // number of points (equal to the number of instances)
    points_count: u32,
}

const circle_triangle_count: u32 = 32;

struct VertexIn {
    // line segment index
    @builtin(instance_index) index: u32,
    @builtin(vertex_index) v_index : u32,
}

struct VertexOut {
    @builtin(position) clip: vec4f,
    @location(0) color: vec4f,
}

@vertex
fn vertex(in: VertexIn) -> VertexOut {
    // for the first segment, we only draw the endcap at the current point,
    // we do not draw the segment to the previous point
    if in.index == 0 && in.v_index < 6 {
        var out: VertexOut;
        out.clip = vec4f(0, 0, 0, 1);
        out.color = vec4f(0, 1, 0, 1);
        return out;
    }

    var color = config.color;

    // get the two points of the line segment to render
    let base = points[in.index];
    let prev = points[in.index-1];

    // calculate a vector orthogonal to the line
    let dir = normalize(prev - base);
    let ortho = vec2f(-dir.y, dir.x) * (config.thickness * 0.5);

    // depending on the vertex index (0 to 3), we need to
    // calculate the position of the vertex in clip space
    //  0     2/5
    //  1/3   4

    var pos: vec2f;
    switch (in.v_index) {
        case 0: {
            pos = base - ortho;
        }
        case 1, 3: {
            pos = base + ortho;
        }
        case 2, 5: {
            pos = prev - ortho;
        }
        case 4: {
            pos = prev + ortho;
        }

        default: {
            // end cap at a. triangles in a circle
            let v = in.v_index - 6;
            let r = config.thickness * 0.5;

            // slice per triangle
            let slice = 3.1415926 * 2.0 / f32(circle_triangle_count);
            let tri = v / 3;

            switch (v % 3) {
                case 0: {
                    pos = base;
                }
                case 1: {
                    let a = f32(tri) * slice;
                    pos = base + vec2f(sin(a), cos(a)) * r;
                }
                default: {
                    let a = f32(tri+1) * slice;
                    pos = base + vec2f(sin(a), cos(a)) * r;
                }
            }
        }
    }

    let clip = config.projection * vec3f(pos.xy, 1);

    var out: VertexOut;
    out.clip = vec4f(clip.xy, 0, 1);
    out.color = color;
    return out;
}

@fragment
fn fragment(in: VertexOut) -> @location(0) vec4f {
    return in.color;
}
